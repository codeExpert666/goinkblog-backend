package api

import (
	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// AIHandler AI API处理器
type AIHandler struct {
	AIService *biz.AIService
}

// processAI 通用 AI 处理函数
func (h *AIHandler) processAI(c *gin.Context, aiType string) {
	var req schema.AIRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// TODO 检查内容是否为空
	if charList := []rune(req.Content); len(charList) < 100 {
		util.ResError(c, errors.BadRequest("文章内容至少 100 字"))
		return
	}

	// 强制设置类型
	req.Type = aiType

	ctx := c.Request.Context()
	data, err := h.AIService.ProcessContent(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 文章润色
// @Param content body string true "原始文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AIResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/polish [post]
func (h *AIHandler) Polish(c *gin.Context) {
	h.processAI(c, "polish")
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 生成标题建议
// @Param content body string true "原始文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AIResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/title [post]
func (h *AIHandler) GenerateTitle(c *gin.Context) {
	h.processAI(c, "title")
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 生成摘要
// @Param content body string true "原始文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AIResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/summary [post]
func (h *AIHandler) GenerateSummary(c *gin.Context) {
	h.processAI(c, "summary")
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 推荐标签
// @Param content body string true "原始文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AIResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/ai/tags [post]
func (h *AIHandler) RecommendTags(c *gin.Context) {
	h.processAI(c, "tags")
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 获取 AI 配置
// @Success 200 {object} util.ResponseResult{data=schema.AIConfig}
// @Router /api/ai/config [get]
func (h *AIHandler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()
	data := h.AIService.GetConfig(ctx)

	util.ResSuccess(c, data)
}

// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 更新 AI 配置
// @Param provider body string false "AI 提供商" enum("openai", "local")
// @Param api_key body string false "AI API 密钥"
// @Param endpoint body string false "AI API 端点"
// @Param model body string false "AI 模型"
// @Param temperature body number false "AI 温度" minimum(0) maximum(2)
// @Success 200 {object} util.ResponseResult{data=schema.AIConfig}
// @Failure 400 {object} util.ResponseResult
// @Router /api/ai/config [put]
func (h *AIHandler) UpdateConfig(c *gin.Context) {
	var req schema.AIConfigUpdateRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data := h.AIService.UpdateConfig(ctx, &req)

	util.ResSuccess(c, data)
}
