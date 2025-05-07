package api

import (
	"fmt"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

// AssistantHandler AI助手控制器
type AssistantHandler struct {
	AssistantService *biz.AssistantService
}

// PolishArticle 润色文章
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 使用AI助手润色文章
// @Accept json
// @Produce text/event-stream
// @Param article_content body string true "文章内容"
// @Success 200 {string} string "流式返回润色后的文章内容"
// @Failure 400 {object} util.ResponseResult
// @Failure 429 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Failure 503 {object} util.ResponseResult
// @Router /api/ai/polish [post]
func (h *AssistantHandler) PolishArticle(c *gin.Context) {
	var req schema.AssistantRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}
	req.Stream = true

	// 流式处理
	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	resp, callback, err := h.AssistantService.Process(ctx, biz.TaskTypePolish, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	// 成功回调
	defer callback()

	// 将管道中的数据流式传输到客户端
	h.streamTransfer(c, resp.StreamChan)
}

// GenerateSummary 生成摘要
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 使用AI助手为文章生成摘要
// @Accept json
// @Produce text/event-stream
// @Param article_content body string true "文章内容"
// @Success 200 {string} string "流式返回生成的文章摘要"
// @Failure 400 {object} util.ResponseResult
// @Failure 429 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Failure 503 {object} util.ResponseResult
// @Router /api/ai/summary [post]
func (h *AssistantHandler) GenerateSummary(c *gin.Context) {
	var req schema.AssistantRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}
	req.Stream = true

	// 流式处理
	ctx := logging.NewTag(c.Request.Context(), logging.TagKeyAI)
	resp, callback, err := h.AssistantService.Process(ctx, biz.TaskTypeSummary, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	// 成功回调
	defer callback()

	// 将管道中的数据流式传输到客户端
	h.streamTransfer(c, resp.StreamChan)
}

// GenerateTitle 生成标题
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 使用AI助手为文章生成5个标题
// @Accept json
// @Produce json
// @Param article_content body string true "文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AssistantResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 429 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Failure 503 {object} util.ResponseResult
// @Router /api/ai/title [post]
func (h *AssistantHandler) GenerateTitle(c *gin.Context) {
	var req schema.AssistantRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	data, callback, err := h.AssistantService.Process(c, biz.TaskTypeTitle, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	defer callback()

	util.ResSuccess(c, data)
}

// GenerateTag 生成标签
// @Tags AIAPI
// @Security ApiKeyAuth
// @Summary 使用AI助手为文章生成6个标签
// @Accept json
// @Produce json
// @Param article_content body string true "文章内容"
// @Success 200 {object} util.ResponseResult{data=schema.AssistantResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 429 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Failure 503 {object} util.ResponseResult
// @Router /api/ai/tag [post]
func (h *AssistantHandler) GenerateTag(c *gin.Context) {
	var req schema.AssistantRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	data, callback, err := h.AssistantService.Process(c, biz.TaskTypeTag, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	defer callback()

	util.ResSuccess(c, data)
}

func (h *AssistantHandler) streamTransfer(c *gin.Context, streamChan <-chan string) {
	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 创建一个心跳计时器，每15秒发送一次
	// 目的是保证连接一直 alive，这在流式传输中很重要
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	// Go的HTTP服务器默认使用缓冲写入，创建辅助函数用于及时刷新缓冲
	flush := func(w io.Writer) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// 流式响应
	c.Stream(func(w io.Writer) bool {
		select {
		case content, ok := <-streamChan:
			if !ok {
				// 通道已关闭，发送完成事件并终止流
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flush(w) // 刷新[DONE]消息
				return false
			}

			// 检查是否为错误消息
			if strings.HasPrefix(content, "ERROR: ") {
				// 处理错误消息，发送带有错误标记的JSON
				errorContent := strings.TrimPrefix(content, "ERROR: ")
				isError := true
				response := schema.AssistantResponse{
					IsError: &isError,
					Content: errorContent,
				}
				responseBytes, err := json.Marshal(response)
				if err != nil {
					fmt.Fprintf(w, "data: {\"is_error\":true,\"content\":\"json序列化错误消息失败\"}\n\n")
					return false
				}
				fmt.Fprintf(w, "data: %s\n\n", responseBytes)
				flush(w) // 刷新错误消息
				return false
			}

			// 发送正常数据
			isError := false
			response := schema.AssistantResponse{
				IsError: &isError,
				Content: content,
			}
			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Fprintf(w, "data: {\"is_error\":true,\"content\":\"json序列化正常消息失败\"}\n\n")
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", responseBytes)
			flush(w) // 刷新正常数据
			return true

		case <-heartbeatTicker.C:
			// 发送心跳消息（SSE注释格式）
			fmt.Fprintf(w, ": heartbeat\n\n")
			flush(w) // 刷新心跳消息
			return true

		case <-c.Request.Context().Done(): // 监听客户端关闭连接
			return false
		}
	})
}
