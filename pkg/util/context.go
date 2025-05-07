package util

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"gorm.io/gorm"
)

// context 上下文 key
type (
	traceIDCtx     struct{}
	transCtx       struct{}
	rowLockCtx     struct{}
	usernameCtx    struct{}
	userTokenCtx   struct{}
	isAdminUserCtx struct{}
	userCacheCtx   struct{}
)

func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtx{}, traceID)
}

func FromTraceID(ctx context.Context) string {
	v := ctx.Value(traceIDCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func NewTrans(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, transCtx{}, db)
}

func FromTrans(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(transCtx{})
	if v != nil {
		return v.(*gorm.DB), true
	}
	return nil, false
}

func NewRowLock(ctx context.Context) context.Context {
	return context.WithValue(ctx, rowLockCtx{}, true)
}

func FromRowLock(ctx context.Context) bool {
	v := ctx.Value(rowLockCtx{})
	return v != nil && v.(bool)
}

func NewUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, usernameCtx{}, userID)
}

func FromUserID(ctx context.Context) uint {
	v := ctx.Value(usernameCtx{})
	if v != nil {
		return v.(uint)
	}
	return 0
}

func NewUserToken(ctx context.Context, userToken string) context.Context {
	return context.WithValue(ctx, userTokenCtx{}, userToken)
}

func FromUserToken(ctx context.Context) string {
	v := ctx.Value(userTokenCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func NewIsAdminUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, isAdminUserCtx{}, true)
}

func FromIsAdminUser(ctx context.Context) bool {
	v := ctx.Value(isAdminUserCtx{})
	return v != nil && v.(bool)
}

// UserCache 用户缓存
type UserCache struct {
	Role string `json:"role"`
}

func ParseUserCache(s string) UserCache {
	var a UserCache
	if s == "" {
		return a
	}

	_ = json.Unmarshal([]byte(s), &a)
	return a
}

func (a UserCache) String() string {
	return json.MarshalToString(a)
}

func NewUserCache(ctx context.Context, userCache UserCache) context.Context {
	return context.WithValue(ctx, userCacheCtx{}, userCache)
}

func FromUserCache(ctx context.Context) UserCache {
	v := ctx.Value(userCacheCtx{})
	if v != nil { // 存在用户缓存
		return v.(UserCache)
	}

	// 不存在用户缓存（管理员/匿名用户）
	if id := FromUserID(ctx); id != 0 {
		return UserCache{Role: "admin"}
	}
	return UserCache{Role: "anonymous"}
}

func ExtendContext(newCtx context.Context, oldCtx context.Context) context.Context {
	// 复制 TraceID
	if traceID := FromTraceID(oldCtx); traceID != "" {
		newCtx = NewTraceID(newCtx, traceID)
	}
	// 复制 Transaction 数据库事务
	if db, ok := FromTrans(oldCtx); ok {
		newCtx = NewTrans(newCtx, db)
	}
	// 复制数据库行锁
	if ok := FromRowLock(oldCtx); ok {
		newCtx = NewRowLock(newCtx)
	}
	// 复制 UserID
	if id := FromUserID(oldCtx); id != 0 {
		newCtx = NewUserID(newCtx, id)
	}
	// 复制用户缓存
	if cache := FromUserCache(oldCtx); cache.Role != "" {
		newCtx = NewUserCache(newCtx, cache)
	}

	return newCtx
}
