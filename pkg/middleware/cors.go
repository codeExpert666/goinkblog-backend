package middleware

import (
	"context"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig 定义了跨域资源共享(CORS)的配置结构
// CORS 允许浏览器向跨源服务器，发出 XMLHttpRequest 请求
type CORSConfig struct {
	// Enable 表示是否启用 CORS 中间件
	Enable bool
	// AllowAllOrigins 表示是否允许所有来源的跨域请求
	AllowAllOrigins bool
	// AllowOrigins 定义允许的跨域请求来源列表
	// 如果包含 "*"，则允许所有来源
	// 默认为空列表 []
	AllowOrigins []string
	// AllowMethods 定义允许的 HTTP 请求方法列表
	// 默认值包括：GET, POST, PUT, PATCH, DELETE, HEAD 和 OPTIONS
	AllowMethods []string
	// AllowHeaders 定义允许的非简单请求头列表
	// 用于指定允许跨域请求时可以使用的 HTTP 请求头
	AllowHeaders []string
	// AllowCredentials 表示是否允许跨域请求携带认证信息
	// 例如 cookies、HTTP 认证信息或客户端 SSL 证书
	AllowCredentials bool
	// ExposeHeaders 定义可以暴露给客户端的响应头列表
	ExposeHeaders []string
	// MaxAge 定义预检请求的缓存时间（以秒为单位）
	// 在该时间内，浏览器无需再次发送预检请求
	MaxAge int
	// AllowWildcard 允许使用通配符来指定来源
	// 例如：http://some-domain/*, https://api.* 或 http://some.*.subdomain.com
	AllowWildcard bool
	// AllowBrowserExtensions 允许使用浏览器扩展的特殊协议架构
	AllowBrowserExtensions bool
	// AllowWebSockets 允许使用 WebSocket 协议
	AllowWebSockets bool
	// AllowFiles 允许使用 file:// 协议（危险！）
	// 仅在确实需要时使用
	AllowFiles bool
}

var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
}

func CORSWithConfig(cfg CORSConfig) gin.HandlerFunc {
	if !cfg.Enable {
		logging.Context(context.Background()).Debug("未启用 CORS 中间件")
		return Empty() // 如果未启用 CORS，返回一个空的中间件
	}

	return cors.New(cors.Config{
		AllowAllOrigins:        cfg.AllowAllOrigins,
		AllowOrigins:           cfg.AllowOrigins,
		AllowMethods:           cfg.AllowMethods,
		AllowHeaders:           cfg.AllowHeaders,
		AllowCredentials:       cfg.AllowCredentials,
		ExposeHeaders:          cfg.ExposeHeaders,
		MaxAge:                 time.Second * time.Duration(cfg.MaxAge),
		AllowWildcard:          cfg.AllowWildcard,
		AllowBrowserExtensions: cfg.AllowBrowserExtensions,
		AllowWebSockets:        cfg.AllowWebSockets,
		AllowFiles:             cfg.AllowFiles,
	})
}
