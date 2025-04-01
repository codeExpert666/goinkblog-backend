package jwtx

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type Auther interface {
	// GenerateToken 生成一个 JWT 令牌，使用提供的主题。
	GenerateToken(ctx context.Context, subject uint) (TokenInfo, error)
	// DestroyToken 通过从令牌存储中删除令牌来使令牌失效。
	DestroyToken(ctx context.Context, accessToken string) error
	// ParseSubject 从给定的访问令牌中解析主题（或用户标识符）。
	ParseSubject(ctx context.Context, accessToken string) (uint, error)
	// Release 释放 JWTAuth 实例持有的任何资源。
	Release(ctx context.Context) error
}

const defaultKey = "CG24SDVP8OHPK395GB5G"

var ErrInvalidToken = errors.New("invalid token")

type options struct {
	signingMethod jwt.SigningMethod
	signingKey    []byte
	signingKey2   []byte
	keyFuncs      []func(*jwt.Token) (interface{}, error)
	expired       int
	tokenType     string
}

type Option func(*options)

func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

func SetSigningKey(key, oldKey string) Option {
	return func(o *options) {
		o.signingKey = []byte(key)
		if oldKey != "" && key != oldKey {
			o.signingKey2 = []byte(oldKey)
		}
	}
}

func SetExpired(expired int) Option {
	return func(o *options) {
		o.expired = expired
	}
}

func New(store Storer, opts ...Option) Auther {
	o := options{
		tokenType:     "Bearer",
		expired:       7200,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    []byte(defaultKey),
	}

	for _, opt := range opts {
		opt(&o)
	}

	o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
		// 检查签名方法是否为基于 hash 的签名方法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		// 先检查方法再返回密钥是最佳实践
		return o.signingKey, nil
	})

	if o.signingKey2 != nil {
		o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return o.signingKey2, nil
		})
	}

	return &JWTAuth{
		opts:  &o,
		store: store,
	}
}

type JWTAuth struct {
	opts  *options
	store Storer
}

func (a *JWTAuth) GenerateToken(ctx context.Context, subject uint) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second).Unix()

	token := jwt.NewWithClaims(a.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt,
		NotBefore: now.Unix(),
		Subject:   fmt.Sprintf("%d", subject),
	})

	tokenStr, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	tokenInfo := &tokenInfo{
		ExpiresAt:   expiresAt,
		TokenType:   a.opts.tokenType,
		AccessToken: tokenStr,
	}
	return tokenInfo, nil
}

func (a *JWTAuth) parseToken(tokenStr string) (*jwt.StandardClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	for _, keyFunc := range a.opts.keyFuncs {
		token, err = jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, keyFunc)
		if err != nil || token == nil || !token.Valid {
			continue
		}
		break
	}

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token.Claims.(*jwt.StandardClaims), nil
}

func (a *JWTAuth) callStore(fn func(Storer) error) error {
	if store := a.store; store != nil {
		return fn(store)
	}
	return nil
}

func (a *JWTAuth) DestroyToken(ctx context.Context, tokenStr string) error {
	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return err
	}

	return a.callStore(func(store Storer) error {
		expired := time.Until(time.Unix(claims.ExpiresAt, 0)) // 令牌剩余有效期
		return store.Set(ctx, tokenStr, expired)
	})
}

func (a *JWTAuth) ParseSubject(ctx context.Context, tokenStr string) (uint, error) {
	if tokenStr == "" {
		return 0, ErrInvalidToken
	}

	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return 0, err
	}

	err = a.callStore(func(store Storer) error {
		if exists, err := store.Check(ctx, tokenStr); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// 将 Subject 转换为 uint
	uIntSub, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}

	return uint(uIntSub), nil
}

func (a *JWTAuth) Release(ctx context.Context) error {
	return a.callStore(func(store Storer) error {
		return store.Close(ctx)
	})
}
