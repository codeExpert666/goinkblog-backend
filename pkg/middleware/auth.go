package middleware

import (
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	AdminID             uint
	Auth                func(c *gin.Context) bool // 判断当前请求是否需要认证
	ParseUserID         func(c *gin.Context) (uint, error)
}

func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		userID, err := config.ParseUserID(c)

		if err == nil { // 用户已认证
			ctx := util.NewUserID(c.Request.Context(), userID)
			ctx = logging.NewUserID(ctx, userID)
			if userID == config.AdminID {
				ctx = util.NewIsAdminUser(ctx)
			}
			c.Request = c.Request.WithContext(ctx)
		} else if !errors.IsUnauthorized(err) || config.Auth(c) { // 内部错误、用户未认证且请求需要认证
			util.ResError(c, err)
			return
		}

		c.Next()
	}
}
