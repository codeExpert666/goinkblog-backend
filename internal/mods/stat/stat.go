package stat

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/dal"
)

// Stat 统计模块
type Stat struct {
	DB          *gorm.DB
	StatHandler *api.StatHandler
}

// Set 注入 Stat 模块
var Set = wire.NewSet(
	wire.Struct(new(Stat), "*"),

	// 统计相关结构体
	wire.Struct(new(api.StatHandler), "*"),
	wire.Struct(new(biz.StatService), "*"),
	wire.Struct(new(dal.StatRepository), "*"),
)

// AutoMigrate 自动迁移数据库
func (s *Stat) AutoMigrate(ctx context.Context) error {
	return nil
}

// Init 初始化统计模块
func (s *Stat) Init(ctx context.Context) error {
	return nil
}

// RegisterRouters 注册路由
func (s *Stat) RegisterRouters(ctx context.Context, stat *gin.RouterGroup) error {
	// 统计相关接口
	{
		stat.GET("/user/articles", s.StatHandler.GetUserArticleStatistic)
		stat.GET("/overview", s.StatHandler.GetSiteOverview)
		stat.GET("/visits", s.StatHandler.GetVisitTrend)
		stat.GET("/activity", s.StatHandler.GetUserActivityTrend)
		stat.GET("/user/categories", s.StatHandler.GetUserCategoryDistribution)
		stat.GET("/categories", s.StatHandler.GetCategoryDistribution)
		stat.GET("/logger", s.StatHandler.GetLogger)
		stat.GET("/user/articles/visits", s.StatHandler.GetUserArticleVisitTrend)
		stat.GET("/comments", s.StatHandler.GetCommentStatistic)
		stat.GET("/system", s.StatHandler.GetSystemInfo)
		stat.GET("/cpu", s.StatHandler.GetCPUInfo)
		stat.GET("/memory", s.StatHandler.GetMemoryInfo)
		stat.GET("/disk", s.StatHandler.GetDiskInfo)
		stat.GET("/go", s.StatHandler.GetGoInfo)
		stat.GET("/db", s.StatHandler.GetDBInfo)
		stat.GET("/cache", s.StatHandler.GetCacheInfo)
		stat.GET("/articles/creation", s.StatHandler.GetArticleCreationTimeStats)
	}
	return nil
}

// Release 释放模块资源
func (s *Stat) Release(ctx context.Context) error {
	return nil
}
