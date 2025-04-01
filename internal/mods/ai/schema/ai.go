package schema

// AIRequest AI 处理请求
type AIRequest struct {
	Content string `json:"content" binding:"required"`
	Type    string `json:"type"`
}

// AIResponse AI 处理响应
type AIResponse struct {
	Result      string   `json:"result"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// AIConfig AI 配置
type AIConfig struct {
	Provider    string  `json:"provider"` // openai, local
	APIKey      string  `json:"api_key"`
	Endpoint    string  `json:"endpoint"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

// AIConfigUpdateRequest AI 配置更新请求
type AIConfigUpdateRequest struct {
	Provider    string  `json:"provider" binding:"omitempty,oneof=openai local"`
	APIKey      string  `json:"api_key"`
	Endpoint    string  `json:"endpoint"`
	Model       string  `json:"model"`
	Temperature *float64 `json:"temperature" binding:"omitempty,min=0,max=2"`
}
