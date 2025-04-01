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

// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取文章统计信息
// @Success 200 {object} util.ResponseResult{data=schema.ArticleStatisticResponse}
// @Router /api/stat/articles [get]
func (h *StatHandler) GetArticleStatistic(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.StatService.GetArticleStatistic(ctx)

	util.ResSuccess(c, data)
}

// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取访问趋势数据
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

// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取用户活跃度数据
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

// @Tags StatAPI
// @Security ApiKeyAuth
// @Summary 获取日志列表
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(30) default(10)
// @Param level query string false "日志级别"
// @Param trace_id query string false "追踪ID"
// @Param username query string false "用户名关键词"
// @Param tag query string false "日志标签"
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
