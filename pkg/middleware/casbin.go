package middleware

import (
	"github.com/casbin/casbin/v2"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var ErrCasbinDenied = errors.Forbidden("您的权限不足，访问已拒绝")

type CasbinConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	GetEnforcer         func(c *gin.Context) *casbin.Enforcer
	GetSubjects         func(c *gin.Context) []string
}

func CasbinWithConfig(config CasbinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logging.Context(c.Request.Context()).Debug("进入 Casbin 中间件")
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			logging.Context(c.Request.Context()).Debug("跳过 Casbin 中间件")
			c.Next()
			return
		}

		// 获取 Casbin 执行器
		enforcer := config.GetEnforcer(c)

		for _, sub := range config.GetSubjects(c) {
			logging.Context(c.Request.Context()).Debug("获取到用户角色", zap.String("user_role", sub))
			// 到这里说明 Auth 中间件认证通过，匿名用户只可能访问公共接口
			if sub == "anonymous" {
				logging.Context(c.Request.Context()).Debug("未登录用户通过 casbin")
				c.Next()
				return
			}
			if b, err := enforcer.Enforce(sub, c.Request.URL.Path, c.Request.Method); err != nil {
				util.ResError(c, err)
				return
			} else if b {
				logging.Context(c.Request.Context()).Debug("登录用户通过 casbin")
				c.Next()
				return
			}
		}
		util.ResError(c, ErrCasbinDenied)
	}
}
