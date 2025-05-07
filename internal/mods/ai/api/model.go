package api

import (
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"strconv"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/gin-gonic/gin"
)

// ModelHandler 模型控制器
type ModelHandler struct {
	ModelService *biz.ModelService
}

// GetModel 获取模型
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 根据ID获取模型配置（仅管理员可用）
// @Param id path uint true "模型ID"
// @Success 200 {object} util.ResponseResult{data=schema.Model}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models/{id} [get]
func (h *ModelHandler) GetModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的模型ID"))
		return
	}

	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	data, err := h.ModelService.GetModel(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// ListModels 列出所有模型
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 列出模型配置，支持分页和排序（仅管理员可用）
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(300) default(10)
// @Param provider query string false "提供商"
// @Param model query string false "模型名称"
// @Param active query string false "是否只显示激活的模型" Enums(true, false)
// @Param sort_by_weight query string false "如何按照模型权重排序" Enums(desc, asc)
// @Param sort_by_rpm query string false "如何按照模型的每分钟最大请求数排序" Enums(desc, asc)
// @Param sort_by_current_tokens query string false "如何按照模型当前可用令牌数排序" Enums(desc, asc)
// @Param sort_by_success_count query string false "如何按照模型成功请求数排序" Enums(desc, asc)
// @Param sort_by_failure_count query string false "如何按照模型失败请求数排序" Enums(desc, asc)
// @Param sort_by_avg_latency query string false "如何按照模型请求平均延迟排序（注意：平均延迟为0表示尚不存在有效的延迟数据，该操作会剔除这类数据）" Enums(desc, asc)
// @Success 200 {object} util.ResponseResult{data=schema.ListModelsResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models [get]
func (h *ModelHandler) ListModels(c *gin.Context) {
	// 解析请求
	var req schema.ListModelsRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 默认值设置
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 调用业务层获取分页数据
	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	data, err := h.ModelService.ListModels(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// CreateModel 创建模型
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 创建新的模型配置
// @Param provider body string true "提供商" Enums(openai, local)
// @Param api_key body string false "API密钥，当provider=openai时必需"
// @Param endpoint body string true "API端点，当provider=openai时必需，且须是有效的HTTP/HTTPS URL" format(url)
// @Param model body string true "模型名称"
// @Param temperature body float64 false "温度参数" minimum(0) maximum(2) default(0.7)
// @Param timeout body int false "访问超时时间（秒）" minimum(1) maximum(300) default(30)
// @Param active body bool false "是否激活" default(false)
// @Param description body string false "描述" maxLength(256)
// @Param rpm body int false "每分钟最大请求数" minimum(1) default(10)
// @Param weight body int false "权重" minimum(1) default(100)
// @Success 200 {object} util.ResponseResult{data=schema.Model}
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models [post]
func (h *ModelHandler) CreateModel(c *gin.Context) {
	var req schema.CreateModelRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 默认值设置
	if req.Provider == "local" {
		req.APIKey = "ollama"
		req.Endpoint = "http://localhost:11434/v1/chat/completions"
	}
	if req.Temperature == nil {
		*req.Temperature = 0.7
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	if req.RPM <= 0 {
		req.RPM = 10
	}
	if req.Weight <= 0 {
		req.Weight = 100
	}

	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	data, err := h.ModelService.CreateModel(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// UpdateModel 更新模型
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 更新现有的模型配置
// @Param id path uint true "模型ID"
// @Param provider body string false "提供商" Enums(openai, local)
// @Param api_key body string false "API密钥，当provider=openai时必需"
// @Param endpoint body string false "API端点，当provider=openai时必需，且须是有效的HTTP/HTTPS URL" format(url)
// @Param model body string false "模型名称"
// @Param temperature body float64 false "温度参数" minimum(0) maximum(2)
// @Param timeout body int false "访问超时时间（秒）" minimum(1) maximum(300)
// @Param active body bool false "是否激活"
// @Param description body string false "描述" maxLength(256)
// @Param rpm body int false "每分钟最大请求数" minimum(1)
// @Param weight body int false "权重" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.Model}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models/{id} [put]
func (h *ModelHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的模型ID"))
		return
	}

	var req schema.UpdateModelRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 请求字段有效性校验
	if req.Provider != nil && *req.Provider == "openai" {
		if req.APIKey == nil {
			util.ResError(c, errors.BadRequest("模型提供商为 OpenAI 时，API 密钥不能为空"))
			return
		}
		if req.Endpoint == nil {
			util.ResError(c, errors.BadRequest("模型提供商为 OpenAI 时，API 端点不能为空"))
			return
		}
	}
	if req.ModelName == nil {
		util.ResError(c, errors.BadRequest("模型名称不能为空"))
		return
	}

	// 默认值填充
	if req.Provider != nil && *req.Provider == "local" {
		*req.APIKey = "ollama"
		*req.Endpoint = "http://localhost:11434/v1/chat/completions"
	}
	if req.Temperature == nil {
		*req.Temperature = 0.7
	}
	if req.Timeout == nil {
		*req.Timeout = 30
	}
	if req.RPM == nil {
		*req.RPM = 10
	}
	if req.Weight == nil {
		*req.Weight = 100
	}

	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	data, err := h.ModelService.UpdateModel(ctx, uint(id), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// DeleteModel 删除模型
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 删除现有的模型配置
// @Param id path uint true "模型ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models/{id} [delete]
func (h *ModelHandler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		util.ResError(c, errors.BadRequest("无效的模型ID"))
		return
	}

	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	err = h.ModelService.DeleteModel(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// ResetStats 重置统计信息
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 重置特定模型的使用统计信息
// @Param id path int true "模型ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models/{id}/reset [post]
func (h *ModelHandler) ResetStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的模型ID"))
		return
	}

	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	err = h.ModelService.ResetStats(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// ResetAllStats 重置所有统计信息
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 重置所有模型的使用统计信息
// @Success 200 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/models/reset [post]
func (h *ModelHandler) ResetAllStats(c *gin.Context) {
	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	err := h.ModelService.ResetAllStats(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// GetOverviewStats 获取总体统计信息
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 获取所有AI模型的总体统计信息，适合仪表板展示
// @Success 200 {object} util.ResponseResult{data=schema.ModelOverallStatsResponse}
// @Router /api/ai/models/overview [get]
func (h *ModelHandler) GetOverviewStats(c *gin.Context) {
	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	data := h.ModelService.GetOverviewStats(ctx)
	util.ResSuccess(c, data)
}
