package middleware

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

var ErrCasbinDenied = errors.Forbidden("您的权限不足，访问已拒绝")

type CasbinConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Enforce             func(ctx context.Context, sub, path, method string) (bool, error)
	GetSubjects         func(c *gin.Context) []string
}

func CasbinWithConfig(config CasbinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		for _, sub := range config.GetSubjects(c) {
			// 到这里说明 Auth 中间件认证通过，匿名用户只可能访问公共接口
			if sub == "anonymous" { 
				c.Next()
				return
			}
			if b, err := config.Enforce(c.Request.Context(), sub, c.Request.URL.Path, c.Request.Method); err != nil {
				util.ResError(c, err)
			} else if b {
				c.Next()
				return
			}
		}
		util.ResError(c, ErrCasbinDenied)
	}
}
