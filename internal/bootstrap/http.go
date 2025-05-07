package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/casbin/casbin/v2"
	"net/http"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/wirex"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/middleware"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"

	_ "github.com/codeExpert666/goinkblog-backend/internal/swagger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// 启动HTTP服务
func startHTTPServer(ctx context.Context, injector *wirex.Injector) (func(), error) {
	if config.C.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()

	e.GET("/health", func(c *gin.Context) {
		util.ResOK(c)
	})

	e.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		Skip: config.C.Middleware.Recovery.Skip,
	}))

	e.NoMethod(func(c *gin.Context) {
		util.ResError(c, errors.MethodNotAllowed(""))
	})

	e.NoRoute(func(c *gin.Context) {
		util.ResError(c, errors.NotFound(""))
	})

	allowedPrefixes := injector.M.RouterPrefixes()

	// 注册中间件
	if err := useHTTPMiddlewares(ctx, e, injector, allowedPrefixes); err != nil {
		return nil, err
	}

	// 注册路由
	if err := injector.M.RegisterRouters(ctx, e); err != nil {
		return nil, err
	}

	// 注册swagger
	if !config.C.General.DisableSwagger {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	if dir := config.C.Middleware.Static.Dir; dir != "" {
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:                dir,
			SkippedPathPrefixes: allowedPrefixes, // 非 api/ 路径
		}))
	}

	addr := config.C.General.HTTP.Addr
	logging.Context(ctx).Info(fmt.Sprintf("HTTP 服务已启动，监听地址：%s", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  time.Second * time.Duration(config.C.General.HTTP.ReadTimeout),  // 服务器读取请求头的最大允许时间
		WriteTimeout: time.Second * time.Duration(config.C.General.HTTP.WriteTimeout), // 服务器写入响应的最大允许时间
		IdleTimeout:  time.Second * time.Duration(config.C.General.HTTP.IdleTimeout),  // 服务器保持空闲连接的最大时间（需要启动 keep-alives）
	}

	go func() {
		var err error
		if config.C.General.HTTP.CertFile != "" && config.C.General.HTTP.KeyFile != "" {
			// TLS 是 HTTPS 使用的加密协议，TLS 1.2 是一个相对安全且广泛支持的 TLS 版本
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(config.C.General.HTTP.CertFile, config.C.General.HTTP.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Context(ctx).Error("监听 HTTP 服务失败", zap.Error(err))
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.C.General.HTTP.ShutdownTimeout))
		defer cancel()

		// 禁用 HTTP Keep-Alive 连接，加速关闭
		srv.SetKeepAlivesEnabled(false)
		// Shutdown 优雅关闭服务器
		if err := srv.Shutdown(ctx); err != nil {
			logging.Context(ctx).Error("关闭 HTTP 服务失败", zap.Error(err))
		}
	}, nil
}

// 使用HTTP中间件
func useHTTPMiddlewares(ctx context.Context, e *gin.Engine, injector *wirex.Injector, allowedPrefixes []string) error {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Enable:                 config.C.Middleware.CORS.Enable,
		AllowAllOrigins:        config.C.Middleware.CORS.AllowAllOrigins,
		AllowOrigins:           config.C.Middleware.CORS.AllowOrigins,
		AllowMethods:           config.C.Middleware.CORS.AllowMethods,
		AllowHeaders:           config.C.Middleware.CORS.AllowHeaders,
		AllowCredentials:       config.C.Middleware.CORS.AllowCredentials,
		ExposeHeaders:          config.C.Middleware.CORS.ExposeHeaders,
		MaxAge:                 config.C.Middleware.CORS.MaxAge,
		AllowWildcard:          config.C.Middleware.CORS.AllowWildcard,
		AllowBrowserExtensions: config.C.Middleware.CORS.AllowBrowserExtensions,
		AllowWebSockets:        config.C.Middleware.CORS.AllowWebSockets,
		AllowFiles:             config.C.Middleware.CORS.AllowFiles,
	}))

	e.Use(middleware.TraceWithConfig(middleware.TraceConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Trace.SkippedPathPrefixes,
		RequestHeaderKey:    config.C.Middleware.Trace.RequestHeaderKey,
		ResponseTraceKey:    config.C.Middleware.Trace.ResponseTraceKey,
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		AllowedPathPrefixes:      allowedPrefixes,
		SkippedPathPrefixes:      config.C.Middleware.Logger.SkippedPathPrefixes,
		MaxOutputRequestBodyLen:  config.C.Middleware.Logger.MaxOutputRequestBodyLen,
		MaxOutputResponseBodyLen: config.C.Middleware.Logger.MaxOutputResponseBodyLen,
	}))

	e.Use(middleware.CopyBodyWithConfig(middleware.CopyBodyConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.CopyBody.SkippedPathPrefixes,
		MaxContentLen:       config.C.Middleware.CopyBody.MaxContentLen,
	}))

	// 定义auth函数，通过RBAC策略来确定请求是否需要认证
	authFunc := func(c *gin.Context) bool {
		logging.Context(c.Request.Context()).Debug("进入 authFunc 函数")
		// 检查"anonymous"用户是否有权访问
		e := injector.M.Auth.CasbinHandler.CasbinService.Casbinx.GetEnforcer()
		allowed, err := e.Enforce(
			"anonymous",
			c.Request.URL.Path,
			c.Request.Method,
		)
		if err != nil {
			logging.Context(c.Request.Context()).Error("认证中间件检查请求是否需要认证失败", zap.Error(err))
			return true
		}
		logging.Context(c.Request.Context()).Debug("请求是否需要认证", zap.Bool("need_auth", !allowed))
		// 如果allowed为true，表示匿名用户可以访问，不需要认证
		return !allowed
	}

	e.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Auth.SkippedPathPrefixes,
		ParseUserID:         injector.M.Auth.AuthHandler.AuthService.ParseUserID,
		AdminID:             config.C.General.Admin.ID,
		Auth:                authFunc,
	}))

	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Enable:              config.C.Middleware.RateLimiter.Enable,
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.RateLimiter.SkippedPathPrefixes,
		IPLimit:             config.C.Middleware.RateLimiter.IPLimit,
		UserLimit:           config.C.Middleware.RateLimiter.UserLimit,
		RedisConfig: middleware.RateLimiterRedisConfig{
			Addr:     config.C.Middleware.RateLimiter.Redis.Addr,
			Password: config.C.Middleware.RateLimiter.Redis.Password,
			DB:       config.C.Middleware.RateLimiter.Redis.DB,
			Username: config.C.Middleware.RateLimiter.Redis.Username,
		},
	}))

	e.Use(middleware.CasbinWithConfig(middleware.CasbinConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Casbin.SkippedPathPrefixes,
		GetEnforcer: func(c *gin.Context) *casbin.Enforcer {
			return injector.M.Auth.CasbinHandler.CasbinService.Casbinx.GetEnforcer()
		},
		GetSubjects: func(c *gin.Context) []string {
			return []string{util.FromUserCache(c.Request.Context()).Role}
		},
	}))

	return nil
}
