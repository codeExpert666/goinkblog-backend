package middleware

import (
	"context"
	"fmt"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimiterConfig struct {
	Enable              bool
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	IPLimit             int
	UserLimit           int
	RedisConfig         RateLimiterRedisConfig
}

func RateLimiterWithConfig(cfg RateLimiterConfig) gin.HandlerFunc {
	if !cfg.Enable {
		return Empty()
	}

	redisRateLimiter := NewRatelimiterRedis(cfg.RedisConfig)

	return func(c *gin.Context) {
		logging.Context(c.Request.Context()).Debug("进入限流器中间件")
		if !AllowedPathPrefixes(c, cfg.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, cfg.SkippedPathPrefixes...) {
			logging.Context(c.Request.Context()).Debug("跳过限流器中间件")
			c.Next()
			return
		}

		var (
			allowed bool
			err     error
		)

		ctx := c.Request.Context()
		if userID := util.FromUserID(ctx); userID != 0 { // 已登录用户
			if !util.FromIsAdminUser(ctx) { // 普通用户
				allowed, err = redisRateLimiter.Allow(ctx, fmt.Sprintf("user_%d", userID),
					redis_rate.PerSecond(cfg.UserLimit))
			} else { // 管理员
				allowed = true
			}
		} else { // 未登录用户
			allowed, err = redisRateLimiter.Allow(ctx, c.ClientIP(),
				redis_rate.PerSecond(cfg.IPLimit))
		}

		if err != nil {
			logging.Context(ctx).Error("限流器中间件出错", zap.Error(err))
			util.ResError(c, errors.InternalServerError(""))
		} else if allowed {
			logging.Context(c.Request.Context()).Debug("限流器是否允许请求", zap.Bool("allowed", allowed))
			c.Next()
		} else {
			util.ResError(c, errors.TooManyRequests(""))
		}
	}
}

type RateLimiterRedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRatelimiterRedis(config RateLimiterRedisConfig) *RateLimiterRedis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RateLimiterRedis{
		limiter: redis_rate.NewLimiter(rdb),
	}
}

type RateLimiterRedis struct {
	limiter *redis_rate.Limiter
}

func (r *RateLimiterRedis) Allow(ctx context.Context, identifier string, RequestLimit redis_rate.Limit) (bool, error) {
	result, err := r.limiter.Allow(ctx, identifier, RequestLimit)
	if err != nil {
		return false, err
	}

	return result.Allowed > 0, nil
}
