package biz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
)

// AIService AI 业务逻辑层
type AIService struct {
	Config *schema.AIConfig `wire:"-"`
	Mutex  sync.RWMutex     `wire:"-"`
}

// GetConfig 获取当前 AI 配置
func (s *AIService) GetConfig(ctx context.Context) *schema.AIConfig {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.Config
}

// UpdateConfig 更新 AI 配置
func (s *AIService) UpdateConfig(ctx context.Context, req *schema.AIConfigUpdateRequest) *schema.AIConfig {
	ctx = logging.NewTag(ctx, logging.TagKeyAI)

	// 获取写锁保护配置更新
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// 更新配置
	fields := make([]zap.Field, 0)
	if req.Provider != "" && req.Provider != s.Config.Provider {
		s.Config.Provider = req.Provider
		fields = append(fields, zap.String("provider", s.Config.Provider))
	}
	if req.APIKey != "" && req.APIKey != s.Config.APIKey {
		s.Config.APIKey = req.APIKey
		fields = append(fields, zap.String("api_key", s.Config.APIKey))
	}
	if req.Endpoint != "" && req.Endpoint != s.Config.Endpoint {
		s.Config.Endpoint = req.Endpoint
		fields = append(fields, zap.String("endpoint", s.Config.Endpoint))
	}
	if req.Model != "" && req.Model != s.Config.Model {
		s.Config.Model = req.Model
		fields = append(fields, zap.String("model", s.Config.Model))
	}
	if req.Temperature != nil && *req.Temperature != s.Config.Temperature {
		s.Config.Temperature = *req.Temperature
		fields = append(fields, zap.Float64("temperature", s.Config.Temperature))
	}

	// 记录配置更新
	if len(fields) > 0 {
		logging.Context(ctx).Info("AI 配置已更新", fields...)
	}

	return s.Config
}

// 定义 AI 服务不可用错误
var ErrAIServiceUnavailable = errors.InternalServerError("AI 服务暂时不可用，请稍后重试")

// ProcessContent 处理内容
func (s *AIService) ProcessContent(ctx context.Context, req *schema.AIRequest) (*schema.AIResponse, error) {
	ctx = logging.NewTag(ctx, logging.TagKeyAI)

	// 获取读锁读取配置
	s.Mutex.RLock()
	provider := s.Config.Provider
	s.Mutex.RUnlock()

	switch provider {
	case "openai":
		return s.processWithOpenAI(ctx, req)
	case "local":
		// 本地 AI 模型处理逻辑
		return s.processWithLocalModel(ctx, req)
	default:
		logging.Context(ctx).Error("不支持的 AI 提供商", zap.String("provider", provider))
		return nil, ErrAIServiceUnavailable
	}
}

// processWithOpenAI 使用 OpenAI 处理内容
func (s *AIService) processWithOpenAI(ctx context.Context, req *schema.AIRequest) (*schema.AIResponse, error) {
	// 获取读锁读取配置
	s.Mutex.RLock()
	config := *s.Config // 复制配置以避免长时间持有锁
	s.Mutex.RUnlock()

	var prompt string

	switch req.Type {
	case "polish":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你对这篇文章的内容进行润色，提高其可读性和表达效果，但保留原意:\n\n%s", req.Content)
	case "title":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章生成5个吸引人的标题建议，确保它们简洁、有吸引力并且反映文章核心内容:\n\n%s\n\n 生成的标题按照以下格式返回:\n\n1. 标题1\n2. 标题2\n3. 标题3\n4. 标题4\n5. 标题5", req.Content)
	case "summary":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章生成一个简洁的摘要（不超过150字），突出文章的主要观点和价值:\n\n%s", req.Content)
	case "tags":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章推荐5个合适的标签，这些标签应该反映文章的主题、类别和关键内容:\n\n%s\n\n 生成的标签按照以下格式返回:\n\n1. 标签1\n2. 标签2\n3. 标签3\n4. 标签4\n5. 标签5", req.Content)
	default:
		logging.Context(ctx).Error("不支持的文章处理类型", zap.String("type", req.Type))
		return nil, ErrAIServiceUnavailable
	}

	// 构造请求体
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": config.Temperature,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logging.Context(ctx).Error("JSON 序列化请求体失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", config.Endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		logging.Context(ctx).Error("创建HTTP请求失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.APIKey)

	// 发送请求，设置15秒超时
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		logging.Context(ctx).Error("发送 OpenAI 请求失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}
	defer resp.Body.Close()

	// 检查状态码
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logging.Context(ctx).Error("OpenAI API 返回非成功状态码", zap.Int("status_code", resp.StatusCode), zap.String("body", string(bodyBytes)))
		return nil, ErrAIServiceUnavailable
	}

	// 解析响应
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logging.Context(ctx).Error("解析 OpenAI 响应失败", zap.String("body", string(bodyBytes)), zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	// 提取内容
	var content string
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if contentVal, ok := message["content"]; ok {
					if contentStr, ok := contentVal.(string); ok {
						content = contentStr
					}
				}
			}
		}
	}

	if content == "" {
		logging.Context(ctx).Error("AI 响应内容为空或格式异常", zap.String("body", string(bodyBytes)))
		return nil, ErrAIServiceUnavailable
	}

	// 根据不同类型处理响应
	result := &schema.AIResponse{}

	switch req.Type {
	case "polish":
		result.Result = content
	case "summary":
		result.Result = content
	case "title", "tags":
		// 分割多行内容为建议列表
		lines := strings.Split(content, "\n")
		var suggestions []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				// 去除可能的序号前缀，如 "1. "、"- " 等
				line = strings.TrimLeft(line, "0123456789.- ")
				line = strings.TrimSpace(line)
				suggestions = append(suggestions, line)
			}
		}

		if len(suggestions) > 0 {
			result.Result = suggestions[0]
			result.Suggestions = suggestions
		} else {
			result.Result = content
		}
	}

	return result, nil
}

// processWithLocalModel 使用本地模型处理内容
func (s *AIService) processWithLocalModel(ctx context.Context, req *schema.AIRequest) (*schema.AIResponse, error) {
	// 获取读锁读取配置
	s.Mutex.RLock()
	config := *s.Config // 复制配置以避免长时间持有锁
	s.Mutex.RUnlock()

	var prompt string

	switch req.Type {
	case "polish":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你对这篇文章的内容进行润色，提高其可读性和表达效果，但保留原意:\n\n%s", req.Content)
	case "title":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章生成5个吸引人的标题建议，确保它们简洁、有吸引力并且反映文章核心内容:\n\n%s\n\n 生成的标题按照以下格式返回:\n\n1. 标题1\n2. 标题2\n3. 标题3\n4. 标题4\n5. 标题5", req.Content)
	case "summary":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章生成一个简洁的摘要（不超过150字），突出文章的主要观点和价值:\n\n%s", req.Content)
	case "tags":
		prompt = fmt.Sprintf("以下是一篇 markdown 格式的文章，请你根据其内容为这篇文章推荐5个合适的标签，这些标签应该反映文章的主题、类别和关键内容:\n\n%s\n\n 生成的标签按照以下格式返回:\n\n1. 标签1\n2. 标签2\n3. 标签3\n4. 标签4\n5. 标签5", req.Content)
	default:
		logging.Context(ctx).Error("不支持的文章处理类型", zap.String("type", req.Type))
		return nil, ErrAIServiceUnavailable
	}

	// 构造Ollama请求体
	requestBody := map[string]interface{}{
		"model":  config.Model,
		"prompt": prompt,
		"options": map[string]interface{}{
			"temperature": config.Temperature,
		},
		"stream": false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logging.Context(ctx).Error("JSON 序列化请求体失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	// 获取Ollama API地址
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate" // Ollama默认地址
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		logging.Context(ctx).Error("创建HTTP请求失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second, // 本地模型可能需要更长的处理时间
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		logging.Context(ctx).Error("发送请求到本地 Ollama 失败", zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}
	defer resp.Body.Close()

	// 检查状态码
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logging.Context(ctx).Error("Ollama API 返回非成功状态码", zap.Int("status_code", resp.StatusCode), zap.String("body", string(bodyBytes)))
		return nil, ErrAIServiceUnavailable
	}

	// 解析Ollama响应
	var response struct {
		Model         string `json:"model"`
		Created       int64  `json:"created_at"`
		Response      string `json:"response"`
		Done          bool   `json:"done"`
		TotalDuration int64  `json:"total_duration"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logging.Context(ctx).Error("解析 Ollama 响应失败", zap.String("body", string(bodyBytes)), zap.Error(errors.WithStack(err)))
		return nil, ErrAIServiceUnavailable
	}

	if response.Response == "" {
		logging.Context(ctx).Error("Ollama 响应内容为空", zap.String("body", string(bodyBytes)))
		return nil, ErrAIServiceUnavailable
	}

	// 根据不同类型处理响应
	result := &schema.AIResponse{}

	switch req.Type {
	case "polish":
		result.Result = response.Response
	case "summary":
		result.Result = response.Response
	case "title", "tags":
		// 分割多行内容为建议列表
		lines := strings.Split(response.Response, "\n")
		var suggestions []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				// 去除可能的序号前缀，如 "1. "、"- " 等
				line = strings.TrimLeft(line, "0123456789.- ")
				line = strings.TrimSpace(line)
				suggestions = append(suggestions, line)
			}
		}

		if len(suggestions) > 0 {
			result.Result = suggestions[0]
			result.Suggestions = suggestions
		} else {
			result.Result = response.Response
		}
	}

	return result, nil
}
