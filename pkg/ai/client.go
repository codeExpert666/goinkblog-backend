package ai

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
)

// StreamChunk 表示流式响应的一个数据块
type StreamChunk struct {
	Content      string // 内容
	FinishReason string // 结束原因
}

// LLMClient 定义了大语言模型客户端的接口
type LLMClient interface {
	// Call 同步调用模型，返回文本响应
	Call(ctx context.Context, prompt string) (string, error)
	// StreamCall 流式调用模型，返回响应流
	StreamCall(ctx context.Context, prompt string) (<-chan string, error)
}

// ClientConfig 包含所有客户端配置选项
type clientConfig struct {
	provider    string
	apiKey      string
	endpoint    string
	model       string
	temperature float64
	timeout     int
}

// ClientOption 是设置客户端选项的函数类型
type ClientOption func(*clientConfig)

// SetProvider 设置 LLM 提供商
func SetProvider(provider string) ClientOption {
	return func(o *clientConfig) {
		o.provider = provider
	}
}

// SetAPIKey 设置 API 密钥
func SetAPIKey(apiKey string) ClientOption {
	return func(o *clientConfig) {
		o.apiKey = apiKey
	}
}

// SetEndpoint 设置 API 端点
func SetEndpoint(endpoint string) ClientOption {
	return func(o *clientConfig) {
		o.endpoint = endpoint
	}
}

// SetModel 设置模型名称
func SetModel(model string) ClientOption {
	return func(o *clientConfig) {
		o.model = model
	}
}

// SetTemperature 设置温度参数
func SetTemperature(temperature float64) ClientOption {
	return func(o *clientConfig) {
		o.temperature = temperature
	}
}

// SetTimeout 设置超时时间 (秒)
func SetTimeout(timeout int) ClientOption {
	return func(o *clientConfig) {
		o.timeout = timeout
	}
}

// NewClient 创建新的 LLM 客户端
func NewClient(opts ...ClientOption) (LLMClient, error) {
	config := &clientConfig{}

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 根据提供商创建对应客户端
	switch config.provider {
	case "openai":
		return newOpenAIClient(config)
	case "local": // 目前 ollama 使用 OpenAI 兼容模式，若有误，则后续可单独为 ollama 写一个客户端实现
		return newOpenAIClient(config)
	default:
		return nil, errors.Errorf("不支持的提供商: %s", config.provider)
	}
}
