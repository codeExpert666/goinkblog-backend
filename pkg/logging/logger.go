// Package logging 结合 zap 与上下文 context 实现链路追踪
package logging

import (
	"context"

	"go.uber.org/zap"
)

// zap "tag" 标签值
const (
	TagKeyMain     = "main"
	TagKeyRecovery = "recovery"
	TagKeyRequest  = "request"
	TagKeyLogin    = "login"
	TagKeyLogout   = "logout"
	TagKeySystem   = "system"
	TagKeyOperate  = "operate"
	TagKeyAI       = "ai"
)

// 上下文键类型
// 上下文的键与 zap 的公共 Field 标签对应
type (
	ctxLoggerKey  struct{}
	ctxTraceIDKey struct{}
	ctxUserIDKey  struct{}
	ctxTagKey     struct{}
	ctxStackKey   struct{}
)

func NewLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

func FromLogger(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxLoggerKey{})
	if v != nil {
		if vv, ok := v.(*zap.Logger); ok {
			return vv
		}
	}
	return zap.L()
}

func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey{}, traceID)
}

func FromTraceID(ctx context.Context) string {
	v := ctx.Value(ctxTraceIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NewUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

func FromUserID(ctx context.Context) uint {
	v := ctx.Value(ctxUserIDKey{})
	if v != nil {
		if u, ok := v.(uint); ok {
			return u
		}
	}
	return 0
}

func NewTag(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, ctxTagKey{}, tag)
}

func FromTag(ctx context.Context) string {
	v := ctx.Value(ctxTagKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NewStack(ctx context.Context, stack string) context.Context {
	return context.WithValue(ctx, ctxStackKey{}, stack)
}

func FromStack(ctx context.Context) string {
	v := ctx.Value(ctxStackKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func Context(ctx context.Context) *zap.Logger {
	var fields []zap.Field
	if v := FromTraceID(ctx); v != "" {
		fields = append(fields, zap.String("trace_id", v))
	}
	if v := FromUserID(ctx); v != 0 {
		fields = append(fields, zap.Uint("user_id", v))
	}
	if v := FromTag(ctx); v != "" {
		fields = append(fields, zap.String("tag", v))
	}
	if v := FromStack(ctx); v != "" {
		fields = append(fields, zap.String("stack", v))
	}
	return FromLogger(ctx).With(fields...) // 创建子日志记录器，沿用父日志记录器的配置
}

func ExtendLoggerContext(newCtx context.Context, oldCtx context.Context) context.Context {
	// 复制 Logger
	if logger := FromLogger(oldCtx); logger != nil {
		newCtx = NewLogger(newCtx, logger)
	}

	// 复制 TraceID
	if traceID := FromTraceID(oldCtx); traceID != "" {
		newCtx = NewTraceID(newCtx, traceID)
	}

	// 复制 Tag
	if tag := FromTag(oldCtx); tag != "" {
		newCtx = NewTag(newCtx, tag)
	}

	// 复制 UserID
	if userID := FromUserID(oldCtx); userID != 0 {
		newCtx = NewUserID(newCtx, userID)
	}

	return newCtx
}
