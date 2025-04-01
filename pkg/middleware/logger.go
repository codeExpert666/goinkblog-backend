package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoggerConfig struct {
	AllowedPathPrefixes      []string
	SkippedPathPrefixes      []string
	MaxOutputRequestBodyLen  int
	MaxOutputResponseBodyLen int
}

var DefaultLoggerConfig = LoggerConfig{
	MaxOutputRequestBodyLen:  1024 * 1024, // 默认最大请求体记录长度为1MB
	MaxOutputResponseBodyLen: 1024 * 1024, // 默认最大响应体记录长度为1MB
}

// 记录详细的请求日志，用于快速排查问题。
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		start := time.Now()
		contentType := c.Request.Header.Get("Content-Type")

		// 记录请求日志
		fields := []zap.Field{
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()), // 请求来源
			zap.String("uri", c.Request.RequestURI),
			zap.String("host", c.Request.Host),
			zap.String("remote_addr", c.Request.RemoteAddr),
			zap.String("proto", c.Request.Proto),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.String("content_type", contentType),
			zap.String("pragma", c.Request.Header.Get("Pragma")), // pragma 头部
		}

		c.Next()

		// 记录请求用户信息
		ctx := c.Request.Context()
		fields = append(fields, zap.Uint("user_id", util.FromUserID(ctx)))

		// 对于POST或PUT请求，记录请求体内容（如果是JSON格式）
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			// mime.ParseMediaType 解析请求头中的 Content-Type 字段
			// 返回值有三部分：主媒体类型（如 "application/json"）、媒体类型参数（如 {"charset": "utf-8"}）、错误
			mediaType, _, _ := mime.ParseMediaType(contentType)
			if mediaType == "application/json" {
				if v, ok := c.Get(util.ReqBodyKey); ok {
					if b, ok := v.([]byte); ok && len(b) <= config.MaxOutputRequestBodyLen { // 超出最大长度则不记录
						fields = append(fields, zap.String("body", string(b)))
					}
				}
			}
		}

		// 记录响应日志
		cost := time.Since(start).Nanoseconds() / 1e6
		fields = append(fields, zap.Int64("cost", cost))
		fields = append(fields, zap.Int("status", c.Writer.Status()))
		fields = append(fields, zap.String("res_time", time.Now().Format("2006-01-02 15:04:05.999")))
		fields = append(fields, zap.Int("res_size", c.Writer.Size()))

		if v, ok := c.Get(util.ResBodyKey); ok {
			if b, ok := v.([]byte); ok && len(b) <= config.MaxOutputResponseBodyLen { // 超出最大长度则不记录
				fields = append(fields, zap.String("res_body", string(b)))
			}
		}

		ctx = logging.NewTag(ctx, logging.TagKeyRequest)
		logging.Context(ctx).Info(fmt.Sprintf("[HTTP] %s-%s-%d (%dms)",
			c.Request.URL.Path, c.Request.Method, c.Writer.Status(), cost), fields...)
	}
}
