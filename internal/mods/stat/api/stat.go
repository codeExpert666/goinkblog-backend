package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// StatHandler 统计API处理器
type StatHandler struct {
	StatService *biz.StatService
}

// GetUserArticleVisitTrend 获取用户文章访问趋势
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取用户文章访问趋势
// @Param days query int false "天数" minimum(1) default(7)
// @Success 200 {object} util.ResponseResult{data=schema.UserArticleVisitTrendResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/user/articles/visits [get]
func (h *StatHandler) GetUserArticleVisitTrend(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的天数参数"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.StatService.GetUserArticleVisitTrend(ctx, userID, days)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetUserArticleStatistic 获取用户文章统计信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取用户文章统计信息
// @Success 200 {object} util.ResponseResult{data=schema.SiteOverviewResponse}
// @Router /api/stat/user/articles [get]
func (h *StatHandler) GetUserArticleStatistic(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data := h.StatService.GetUserArticleStatistic(ctx, userID)

	util.ResSuccess(c, data)
}

// GetSiteOverview 获取站点概览统计信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取站点概览统计信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.SiteOverviewResponse}
// @Router /api/stat/overview [get]
func (h *StatHandler) GetSiteOverview(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetOverview(ctx)

	util.ResSuccess(c, data)
}

// GetVisitTrend 获取访问趋势数据
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取访问趋势数据（仅管理员可用）
// @Param days query int false "天数" minimum(1) default(7)
// @Success 200 {object} util.ResponseResult{data=[]schema.APIAccessTrendItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/visits [get]
func (h *StatHandler) GetVisitTrend(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的天数参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.StatService.GetVisitTrend(ctx, days)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetUserActivityTrend 获取用户活跃度数据
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取用户活跃度数据（仅管理员可用）
// @Param days query int false "天数" minimum(1) default(7)
// @Success 200 {object} util.ResponseResult{data=[]schema.UserActivityTrendItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/activity [get]
func (h *StatHandler) GetUserActivityTrend(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的天数参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.StatService.GetUserActivityTrend(ctx, days)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetUserCategoryDistribution 获取用户文章分类分布
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取用户文章分类分布
// @Success 200 {object} util.ResponseResult{data=[]schema.CategoryDistItem}
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/user/categories [get]
func (h *StatHandler) GetUserCategoryDistribution(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.StatService.GetUserCategoryDistribution(ctx, userID)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetCategoryDistribution 获取文章分类分布
// @Tags StatAPI
// @Summary 获取文章分类分布
// @Success 200 {object} util.ResponseResult{data=[]schema.CategoryDistItem}
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/categories [get]
func (h *StatHandler) GetCategoryDistribution(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := h.StatService.GetCategoryDistribution(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetLogger 日志查询
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取日志列表（仅管理员可用）
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(100) default(10)
// @Param level query string false "日志级别" Enums(debug, info, warn, error, dpanic, panic, fatal)
// @Param trace_id query string false "追踪ID"
// @Param username query string false "用户名关键词"
// @Param tag query string false "日志标签" Enums(main, recovery, request, login, logout, system, operate, ai)
// @Param message query string false "日志消息关键词"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} util.ResponseResult{data=schema.LoggerPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/logger [get]
func (h *StatHandler) GetLogger(c *gin.Context) {
	var params schema.LoggerQueryParams
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.StatService.GetLogger(ctx, &params)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// GetCommentStatistic 获取评论统计数据
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取所有评论的统计数据（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.CommentStatisticResponse}
// @Router /api/stat/comments [get]
func (h *StatHandler) GetCommentStatistic(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetCommentStatistic(ctx)

	util.ResSuccess(c, data)
}

// GetSystemInfo 获取系统信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取系统信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.SystemInfo}
// @Router /api/stat/system [get]
func (h *StatHandler) GetSystemInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetSystemInfo(ctx)
	util.ResSuccess(c, data)
}

// GetCPUInfo 获取 CPU 信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取 CPU 信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.CPUInfo}
// @Router /api/stat/cpu [get]
func (h *StatHandler) GetCPUInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetCPUInfo(ctx)
	util.ResSuccess(c, data)
}

// GetMemoryInfo 获取内存信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取内存信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.MemoryInfo}
// @Router /api/stat/memory [get]
func (h *StatHandler) GetMemoryInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetMemoryInfo(ctx)
	util.ResSuccess(c, data)
}

// GetDiskInfo 获取硬盘信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取硬盘信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.DiskInfo}
// @Router /api/stat/disk [get]
func (h *StatHandler) GetDiskInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetDiskInfo(ctx)
	util.ResSuccess(c, data)
}

// GetGoInfo 获取 GO 运行时信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取 GO 运行时信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.GoRuntimeInfo}
// @Router /api/stat/go [get]
func (h *StatHandler) GetGoInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetGoRuntimeInfo(ctx)
	util.ResSuccess(c, data)
}

// GetDBInfo 获取数据库信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取数据库信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.DatabaseInfo}
// @Router /api/stat/db [get]
func (h *StatHandler) GetDBInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetDatabaseInfo(ctx)
	util.ResSuccess(c, data)
}

// GetCacheInfo 获取缓存信息
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取缓存信息（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=schema.CacheInfo}
// @Router /api/stat/cache [get]
func (h *StatHandler) GetCacheInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetCacheInfo(ctx)

	util.ResSuccess(c, data)
}

// GetArticleCreationTimeStats 获取文章创作时间统计
// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取文章创作时间统计（仅管理员可用）
// @Param days query int false "天数" minimum(1) default(30)
// @Success 200 {object} util.ResponseResult{data=[]schema.ArticleCreationTimeStatsItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/stat/articles/creation [get]
func (h *StatHandler) GetArticleCreationTimeStats(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的天数参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.StatService.GetArticleCreationTimeStats(ctx, days)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
