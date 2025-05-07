package ai

import (
	"bufio"
	"bytes"
	"context"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

// openAIClient 实现 OpenAI API 的客户端
type openAIClient struct {
	config           *clientConfig
	httpClient       *http.Client // 非流式响应客户端
	httpStreamClient *http.Client // 流式响应客户端
}

// openAIRequest 表示 OpenAI API 请求
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	Stream      bool            `json:"stream"`
}

// openAIMessage 表示 OpenAI API 消息
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse 表示 OpenAI API 同步调用响应
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// openAIStreamResponse 表示 OpenAI API 流式调用响应
type openAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// newOpenAIClient 创建新的 OpenAI 客户端
func newOpenAIClient(config *clientConfig) (*openAIClient, error) {
	return &openAIClient{
		config: config,
		// 非流式请求响应内容少，需要设置超时
		httpClient: &http.Client{
			Timeout: time.Duration(config.timeout) * time.Second,
		},
		// 流式请求响应内容多，不设置超时
		httpStreamClient: &http.Client{},
	}, nil
}

// Call 实现 LLMClient.Call 接口
func (c *openAIClient) Call(ctx context.Context, prompt string) (string, error) {
	// 构建请求体
	reqBody := openAIRequest{
		Model:       c.config.model,
		Messages:    []openAIMessage{{Role: "user", Content: prompt}},
		Temperature: c.config.temperature,
		Stream:      false,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf("API请求失败，状态码: %d, 响应: %s",
			resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var respBody openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	// 验证响应有效性
	if len(respBody.Choices) == 0 {
		return "", errors.Errorf("API返回空响应")
	}

	return respBody.Choices[0].Message.Content, nil
}

func (c *openAIClient) StreamCall(ctx context.Context, prompt string) (<-chan string, error) {
	// 构建请求
	reqBody := openAIRequest{
		Model:       c.config.model,
		Messages:    []openAIMessage{{Role: "user", Content: prompt}},
		Temperature: c.config.temperature,
		Stream:      true,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", c.config.endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpStreamClient.Do(req)
	if err != nil {
		return nil, err
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, errors.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 构建流式消息管道
	streamChan := make(chan string)

	// 读取 API 流式响应
	go func(ctx context.Context) {
		defer close(streamChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(c.splitSSE)

		for scanner.Scan() {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				logging.Context(ctx).Info("客户端连接已断开，取消流式响应")
				return
			default:
				// 继续处理
			}

			// 获取一个完整的 SSE 事件
			eventData := scanner.Text()

			// 解析 SSE 事件数据
			content, done, err := c.parseSSE(eventData)
			if err != nil {
				logging.Context(ctx).Error("解析 SSE 数据失败", zap.Error(err), zap.String("data", eventData))
				continue
			}
			if done {
				break
			}
			if content != "" {
				streamChan <- content
			}
		}

		if err := scanner.Err(); err != nil {
			logging.Context(ctx).Error("读取流式响应失败", zap.Error(err))
			streamChan <- "ERROR: AI助手发生异常，请重试"
		}
	}(ctx)

	return streamChan, nil
}

// splitSSE 从 OpenAI API 响应的流式数据中解析出事件
// 补充，OpenAI API 响应的流式数据格式如下：
// data: {"id": "chatcmpl-123", "object": "chat.completion.chunk", "created": 1700000000, "model": "gpt-4", "choices": [{"index": 0, "delta": {"content": "Hello"}, "finish_reason": null}]}\n\n
// data: {"id": "chatcmpl-123", "object": "chat.completion.chunk", "created": 1700000001, "model": "gpt-4", "choices": [{"index": 0, "delta": {"content": " world!"}, "finish_reason": null}]}\n\n
// data: {"id": "chatcmpl-123", "object": "chat.completion.chunk", "created": 1700000002, "model": "gpt-4", "choices": [{"index": 0, "delta": {}, "finish_reason": "stop"}]}\n\n
// data: [DONE]
func (c *openAIClient) splitSSE(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// 查找事件流中的数据边界："data: "开头，以"\n\n"结尾
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, c.dropCR(data[:i]), nil
	}

	// 如果我们到达文件末尾，但没有找到分隔符，返回剩余的数据
	if atEOF {
		return len(data), c.dropCR(data), nil
	}

	// 请求更多数据
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func (c *openAIClient) dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// parseSSE 解析 OpenAI API 响应的流式数据
func (c *openAIClient) parseSSE(data string) (string, bool, error) {
	if !strings.HasPrefix(data, "data: ") {
		return "", false, nil
	}

	// 移除 “data: ” 前缀
	content := strings.TrimPrefix(data, "data: ")
	if content == "[DONE]" {
		return "", true, nil // 流结束
	}

	// 解析 JSON
	var streamResp openAIStreamResponse
	if err := json.Unmarshal([]byte(content), &streamResp); err != nil {
		return "", false, err
	}

	// 无响应内容
	if len(streamResp.Choices) == 0 {
		return "", false, nil
	}

	// debug
	//f, err := os.OpenFile("AI_Output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//if err != nil {
	//	fmt.Printf("写入文件错误: %v\n", err)
	//} else {
	//	defer f.Close()
	//	if _, err := f.WriteString(streamResp.Choices[0].Delta.Content); err != nil {
	//		fmt.Printf("写入文件错误: %v\n", err)
	//	}
	//}
	//
	//fmt.Printf("%q\n", streamResp.Choices[0].Delta.Content)

	return streamResp.Choices[0].Delta.Content, false, nil
}
