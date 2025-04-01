package middleware

import (
	"fmt"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecoveryConfig struct {
	Skip int // default: 2
}

var DefaultRecoveryConfig = RecoveryConfig{
	Skip: 2,
}

// Recovery 从任何 panic 中恢复并写入500。
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rv := recover(); rv != nil {
				ctx := c.Request.Context()
				ctx = logging.NewTag(ctx, logging.TagKeyRecovery)

				var fields []zap.Field
				fields = append(fields, zap.Strings("error", []string{fmt.Sprintf("%v", rv)}))
				fields = append(fields, zap.StackSkip("stack", config.Skip))

				if gin.IsDebugging() {
					httpRequest, _ := httputil.DumpRequest(c.Request, false) // 将当前的请求信息（false 不包含请求体）转换为字符串
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *" // 将 token 替换为 *，避免在日志中泄露敏感信息
						}
					}
					fields = append(fields, zap.Strings("headers", headers))
				}

				logging.Context(ctx).Error(fmt.Sprintf("[Recovery] %s panic recovered", time.Now().Format("2006/01/02 - 15:04:05")), fields...)
				util.ResError(c, errors.InternalServerError(""))
			}
		}()

		c.Next()
	}
}
