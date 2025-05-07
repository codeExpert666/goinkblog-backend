package schema

// AssistantRequest AI助手请求
type AssistantRequest struct {
	ArticleContent string `json:"article_content" binding:"required"`
	Stream         bool   `json:"-" binding:"-"`
}

// AssistantResponse AI助手响应
type AssistantResponse struct {
	// 非流式传输
	Contents []string `json:"contents,omitempty"` // 标签、标题

	// 流式传输
	IsError    *bool         `json:"is_error,omitempty"` // 是否为错误数据流
	Content    string        `json:"content,omitempty"`  // 数据流内容
	StreamChan <-chan string `json:"-"`
}
