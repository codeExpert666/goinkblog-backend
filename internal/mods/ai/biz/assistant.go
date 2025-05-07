package biz

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/ai"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
	"strings"
	"time"
)

// TaskType 表示AI任务类型
type TaskType string

const (
	TaskTypePolish  TaskType = "polish"  // 润色文章
	TaskTypeTitle   TaskType = "title"   // 生成标题
	TaskTypeTag     TaskType = "tag"     // 生成标签
	TaskTypeSummary TaskType = "summary" // 生成摘要
)

var ErrAIAssistant = errors.InternalServerError("AI助手出错了，请稍后再试")

// AssistantService AI助手
type AssistantService struct {
	Selector *Selector
}

// buildPrompt 根据任务类型构建提示词
func (a *AssistantService) buildPrompt(ctx context.Context, taskType TaskType, content string) string {
	switch taskType {
	case TaskTypePolish:
		return `你是一个专业的文章润色助手，请对以下文章内容进行润色，使其更加通顺、专业、有吸引力，但保持原意不变。润色后的文章应当语言流畅，逻辑清晰，表达准确，同时增强文章的可读性和吸引力。以下是原文内容：

"""
` + content + `
"""

请直接给出润色后的文章全文，不需要额外的解释。`

	case TaskTypeTitle:
		return `你是一个专业的标题生成助手，请基于以下文章内容，生成5个吸引人的标题。这些标题应当:
1. 能够准确反映文章的核心内容
2. 具有吸引力，能激发读者的阅读兴趣
3. 简洁明了，长度适中
4. 风格多样，以供选择

以下是文章内容：
"""
` + content + `
"""

请直接给出5个标题，用逗号分隔，不需要额外的解释。`

	case TaskTypeTag:
		return `你是一个专业的标签生成助手，请基于以下文章内容，生成6个相关标签。这些标签应当:
1. 能够准确反映文章的主题、领域和关键概念
2. 包含读者可能会搜索的关键词
3. 有助于文章的分类和检索
4. 既包含广泛的分类，也包含特定的关键词

以下是文章内容：
"""
` + content + `
"""

请直接给出6个标签，每个标签用一个单词或短语表示，用逗号分隔，不需要额外的解释。`

	case TaskTypeSummary:
		return `你是一个专业的文章摘要生成助手，请基于以下文章内容，生成一段简洁的摘要。这个摘要应当:
1. 概括文章的核心内容和主要观点
2. 突出文章的价值和意义
3. 简明扼要，一般不超过200字
4. 能够吸引读者进一步阅读全文

以下是文章内容：
"""
` + content + `
"""

请直接给出摘要内容，不需要额外的解释。`

	default:
		logging.Context(ctx).Warn("意外的 AI 任务类型", zap.String("task_type", string(taskType)))
		return content
	}
}

// Process 处理AI任务
func (a *AssistantService) Process(ctx context.Context, taskType TaskType, req *schema.AssistantRequest) (*schema.AssistantResponse, func(), error) {
	// 选择模型
	model, callback, err := a.Selector.SelectModel(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 注册对应的LLM客户端
	client, err := ai.NewClient(
		ai.SetProvider(model.Provider),
		ai.SetAPIKey(model.APIKey),
		ai.SetEndpoint(model.Endpoint),
		ai.SetModel(model.ModelName),
		ai.SetTemperature(model.Temperature),
		ai.SetTimeout(model.Timeout))
	if err != nil {
		callback(false, 0)
		logging.Context(ctx).Error("注册模型对应客户端失败", zap.Error(err), zap.Uint("model_id", model.ID))
		return nil, nil, ErrAIAssistant
	}

	// 构建提示词
	prompt := a.buildPrompt(ctx, taskType, req.ArticleContent)

	// 调用LLM
	if req.Stream {
		// 发起请求并记录用时
		startTime := time.Now()
		streamChan, err := client.StreamCall(ctx, prompt)
		latency := time.Since(startTime)

		if err != nil {
			callback(false, latency)
			logging.Context(ctx).Error("流式访问大模型 API 失败", zap.Error(err), zap.Uint("model_id", model.ID))
			return nil, nil, ErrAIAssistant
		}
		return &schema.AssistantResponse{StreamChan: streamChan}, func() {
			callback(true, latency)
		}, nil
	} else {
		// 发起请求并记录用时
		startTime := time.Now()
		resp, err := client.Call(ctx, prompt)
		latency := time.Since(startTime)

		if err != nil {
			callback(false, latency)
			logging.Context(ctx).Error("同步访问大模型 API 失败", zap.Error(err), zap.Uint("model_id", model.ID))
			return nil, nil, ErrAIAssistant
		}

		return &schema.AssistantResponse{Contents: a.parseResp(ctx, taskType, resp)}, func() {
			callback(true, latency)
		}, nil
	}
}

// parseResp 解析同步调用得到的字符串结果（标题、标签）
func (a *AssistantService) parseResp(ctx context.Context, taskType TaskType, resp string) []string {
	respList := strings.Split(resp, ",")
	for i, r := range respList {
		respList[i] = strings.TrimSpace(r)
	}

	if taskType == TaskTypeTitle && len(respList) < 5 {
		logging.Context(ctx).Error("大模型生成的标题数量不足", zap.Int("title_len", len(respList)), zap.String("resp", resp))
		for range 5 - len(respList) {
			respList = append(respList, "抱歉，AI助手出错了")
		}
	}
	if taskType == TaskTypeTag && len(respList) < 6 {
		logging.Context(ctx).Error("大模型生成的标签数量不足", zap.Int("tag_len", len(respList)), zap.String("resp", resp))
		for range 6 - len(respList) {
			respList = append(respList, "抱歉，AI助手出错了")
		}
	}

	return respList
}
