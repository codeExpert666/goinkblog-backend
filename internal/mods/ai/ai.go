package ai

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
)

// AI AI模块
type AI struct {
	AIHandler *api.AIHandler
}

// Set 注入 AI 模块
var Set = wire.NewSet(
	wire.Struct(new(AI), "*"),

	// AI 相关结构体
	wire.Struct(new(api.AIHandler), "*"),
	wire.Struct(new(biz.AIService), "*"),
)

// AutoMigrate 自动迁移数据库
func (a *AI) AutoMigrate(ctx context.Context) error {
	return nil
}

// Init 初始化 AI 模块
func (a *AI) Init(ctx context.Context) error {
	// 初始化 AI 配置
	a.AIHandler.AIService.Config = &schema.AIConfig{
		Provider:    config.C.AI.Provider,
		APIKey:      config.C.AI.APIKey,
		Endpoint:    config.C.AI.Endpoint,
		Model:       config.C.AI.Model,
		Temperature: config.C.AI.Temperature,
	}
	return nil
}

// RegisterRouters 注册路由
func (a *AI) RegisterRouters(ctx context.Context, ai *gin.RouterGroup) error {
	// AI 接口
	{
		ai.POST("/polish", a.AIHandler.Polish)
		ai.POST("/title", a.AIHandler.GenerateTitle)
		ai.POST("/summary", a.AIHandler.GenerateSummary)
		ai.POST("/tags", a.AIHandler.RecommendTags)

		// 配置管理接口
		ai.GET("/config", a.AIHandler.GetConfig)    // 获取 AI 配置
		ai.PUT("/config", a.AIHandler.UpdateConfig) // 更新 AI 配置
	}
	return nil
}

// Release 释放资源
func (a *AI) Release(ctx context.Context) error {
	return nil
}
