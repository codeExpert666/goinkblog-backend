package middleware

import (
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	AdminID             uint
	Skipper             func(c *gin.Context) bool
	ParseUserID         func(c *gin.Context) (uint, error)
}

func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		userID, err := config.ParseUserID(c)
		if err != nil {
			util.ResError(c, err)
			return
		}

		ctx := util.NewUserID(c.Request.Context(), userID)
		ctx = logging.NewUserID(ctx, userID)
		if userID == config.AdminID {
			ctx = util.NewIsAdminUser(ctx)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
