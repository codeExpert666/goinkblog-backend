package ai

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// AI AI模块
type AI struct {
	DB               *gorm.DB
	ModelHandler     *api.ModelHandler
	AssistantHandler *api.AssistantHandler
}

// Set 注入 AI 模块
var Set = wire.NewSet(
	wire.Struct(new(AI), "*"),

	// AI助手相关结构体
	wire.Struct(new(api.AssistantHandler), "*"),
	wire.Struct(new(biz.AssistantService), "*"),
	wire.Struct(new(biz.Selector), "*"),

	// AI模型管理相关结构体
	wire.Struct(new(api.ModelHandler), "*"),
	wire.Struct(new(biz.ModelService), "*"),

	// 公共结构体
	wire.Struct(new(dal.ModelRepository), "*"),
)

// AutoMigrate 自动迁移数据库
func (a *AI) AutoMigrate(ctx context.Context) error {
	return a.DB.AutoMigrate(new(schema.Model))
}

// Init 初始化 AI 模块
func (a *AI) Init(ctx context.Context) error {
	// 根据配置自动迁移数据库表结构
	if config.C.Storage.DB.AutoMigrate {
		if err := a.AutoMigrate(ctx); err != nil {
			return err
		}
	}

	// 检查数据库中是否存在模型配置
	var modelCount int64
	if err := a.DB.Model(&schema.Model{}).Count(&modelCount).Error; err != nil {
		return err
	}

	// 数据库中不存在模型配置，则从配置文件中加载
	if modelCount == 0 {
		for _, cfg := range config.C.AI.Models {
			if err := a.DB.Create(&schema.Model{
				Provider:    cfg.Provider,
				APIKey:      cfg.APIKey,
				Endpoint:    cfg.Endpoint,
				ModelName:   cfg.ModelName,
				Temperature: cfg.Temperature,
				Timeout:     cfg.Timeout,
				Active:      cfg.Active,
				Description: cfg.Description,
				RPM:         cfg.RPM,
				Weight:      cfg.Weight,
			}).Error; err != nil {
				return err
			}
		}

	}

	// 初始化模型选择器
	if err := a.AssistantHandler.AssistantService.Selector.Load(ctx); err != nil {
		return err
	}

	// 启动定时权重更新任务
	a.AssistantHandler.AssistantService.Selector.AutoWeightUpdate(ctx)

	return nil
}

// RegisterRouters 注册路由
func (a *AI) RegisterRouters(ctx context.Context, ai *gin.RouterGroup) error {
	// AI助手接口
	{
		ai.POST("/polish", a.AssistantHandler.PolishArticle)
		ai.POST("/title", a.AssistantHandler.GenerateTitle)
		ai.POST("/tag", a.AssistantHandler.GenerateTag)
		ai.POST("/summary", a.AssistantHandler.GenerateSummary)
	}

	// AI模型管理接口
	{
		ai.GET("/models/:id", a.ModelHandler.GetModel)
		ai.GET("/models", a.ModelHandler.ListModels)
		ai.POST("/models", a.ModelHandler.CreateModel)
		ai.PUT("/models/:id", a.ModelHandler.UpdateModel)
		ai.DELETE("/models/:id", a.ModelHandler.DeleteModel)
		ai.POST("/models/:id/reset", a.ModelHandler.ResetStats)
		ai.POST("/models/reset", a.ModelHandler.ResetAllStats)
		ai.GET("/models/overview", a.ModelHandler.GetOverviewStats)
	}

	return nil
}

// Release 释放资源
func (a *AI) Release(ctx context.Context) error {
	if err := a.AssistantHandler.AssistantService.Selector.Release(ctx); err != nil {
		return err
	}
	return nil
}
