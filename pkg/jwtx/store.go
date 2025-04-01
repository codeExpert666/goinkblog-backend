package jwtx

import (
	"context"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
)

// Storer 存储令牌的接口
type Storer interface {
	Set(ctx context.Context, tokenStr string, expiration time.Duration) error
	Delete(ctx context.Context, tokenStr string) error
	Check(ctx context.Context, tokenStr string) (bool, error)
	Close(ctx context.Context) error
}

type storeOptions struct {
	CacheNS string // 默认 "jwt"
}

type StoreOption func(*storeOptions)

func WithCacheNS(ns string) StoreOption {
	return func(o *storeOptions) {
		o.CacheNS = ns
	}
}

func NewStoreWithCache(cache cachex.Cacher, opts ...StoreOption) Storer {
	s := &storeImpl{
		c: cache,
		opts: &storeOptions{
			CacheNS: "jwt",
		},
	}
	for _, opt := range opts {
		opt(s.opts)
	}
	return s
}

type storeImpl struct {
	opts *storeOptions
	c    cachex.Cacher
}

func (s *storeImpl) Set(ctx context.Context, tokenStr string, expiration time.Duration) error {
	return s.c.Set(ctx, s.opts.CacheNS, tokenStr, "", expiration)
}

func (s *storeImpl) Delete(ctx context.Context, tokenStr string) error {
	return s.c.Delete(ctx, s.opts.CacheNS, tokenStr)
}

func (s *storeImpl) Check(ctx context.Context, tokenStr string) (bool, error) {
	return s.c.Exists(ctx, s.opts.CacheNS, tokenStr)
}

func (s *storeImpl) Close(ctx context.Context) error {
	return s.c.Close(ctx)
}
