package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Model 单个AI模型配置
type Model struct {
	ID          uint    `json:"id" gorm:"primaryKey"`                             // 主键
	Provider    string  `json:"provider" gorm:"type:varchar(10);not null"`        // 提供商
	APIKey      string  `json:"api_key" gorm:"type:varchar(256);not null"`        // API密钥
	Endpoint    string  `json:"endpoint" gorm:"type:varchar(256);not null"`       // API端点
	ModelName   string  `json:"model_name" gorm:"type:varchar(64);not null"`      // 模型名称
	Temperature float64 `json:"temperature" gorm:"type:decimal(3,2);default:0.7"` // 温度参数
	Timeout     int     `json:"timeout" gorm:"type:int;default:30"`               // 访问超时时间（秒）
	Active      bool    `json:"active" gorm:"type:boolean;default:true"`          // 是否激活
	Description string  `json:"description" gorm:"type:varchar(256)"`             // 描述

	// 令牌桶与加权轮询相关
	Weight         int       `json:"weight" gorm:"type:int;default:100"`         // 权重
	RPM            int       `json:"rpm" gorm:"type:int;default:10"`             // 每分钟最大请求数
	CurrentTokens  int       `json:"current_tokens" gorm:"type:int;default:10"`  // 当前可用令牌数
	LastRefillTime time.Time `json:"last_refill_time" gorm:"autoCreateTime"`     // 上次令牌补充时间
	SuccessCount   int64     `json:"success_count" gorm:"type:bigint;default:0"` // 成功请求数
	FailureCount   int64     `json:"failure_count" gorm:"type:bigint;default:0"` // 失败请求数
	AvgLatency     float64   `json:"avg_latency" gorm:"type:double;default:0"`   // 请求平均延迟（毫秒）

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}

func (a *Model) TableName() string {
	return config.C.FormatTableName("model")
}

// ======================== 查询操作 ========================

// ListModelsRequest 列出AI模型配置请求
type ListModelsRequest struct {
	Page                int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize            int    `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=300"`
	Provider            string `json:"provider" form:"provider"`
	ModelName           string `json:"model_name" form:"model_name"`
	Active              string `json:"active" form:"active" binding:"omitempty,oneof=true false"`
	SortByWeight        string `json:"sort_by_weight" form:"sort_by_weight" binding:"omitempty,oneof=desc asc"`
	SortByRPM           string `json:"sort_by_rpm" form:"sort_by_rpm" binding:"omitempty,oneof=desc asc"`
	SortByCurrentTokens string `json:"sort_by_current_tokens" form:"sort_by_current_tokens" binding:"omitempty,oneof=desc asc"`
	SortBySuccessCount  string `json:"sort_by_success_count" form:"sort_by_success_count" binding:"omitempty,oneof=desc asc"`
	SortByFailureCount  string `json:"sort_by_failure_count" form:"sort_by_failure_count" binding:"omitempty,oneof=desc asc"`
	SortByAvgLatency    string `json:"sort_by_avg_latency" form:"sort_by_avg_latency" binding:"omitempty,oneof=desc asc"`
}

// ListModelsResponse 列出AI模型配置响应
type ListModelsResponse struct {
	Models     []*Model `json:"models"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// ModelOverallStatsResponse 所有模型的总体统计信息
type ModelOverallStatsResponse struct {
	TotalModels          int64   `json:"total_models"`
	ActiveModels         int64   `json:"active_models"`
	TotalRequests        int64   `json:"total_requests"`
	TotalSuccess         int64   `json:"total_success"`
	TotalFailure         int64   `json:"total_failure"`
	OverallSuccessRate   float64 `json:"overall_success_rate"` // 百分比形式
	OverallAvgLatency    float64 `json:"overall_avg_latency"`
	TotalAvailableTokens int64   `json:"total_available_tokens"`
	MostUsedModel        struct {
		ID           uint   `json:"id"`
		Provider     string `json:"provider"`
		ModelName    string `json:"model_name"`
		RequestCount int64  `json:"request_count"`
	} `json:"most_used_model"`
	MostSuccessfulModel struct {
		ID          uint    `json:"id"`
		Provider    string  `json:"provider"`
		ModelName   string  `json:"model_name"`
		SuccessRate float64 `json:"success_rate"` // 百分比形式
	} `json:"most_successful_model"`
	FastestModel struct {
		ID         uint    `json:"id"`
		Provider   string  `json:"provider"`
		ModelName  string  `json:"model_name"`
		AvgLatency float64 `json:"avg_latency"`
	} `json:"fastest_model"`
}

// ======================== 创建操作 ========================

// CreateModelRequest 创建AI模型配置请求
type CreateModelRequest struct {
	Provider    string   `json:"provider" binding:"required,oneof=openai local"`                     // 提供商
	APIKey      string   `json:"api_key" binding:"required_if=Provider openai"`                      // API密钥
	Endpoint    string   `json:"endpoint" binding:"required_if=Provider openai,url,startswith=http"` // API端点
	ModelName   string   `json:"model_name" binding:"required"`                                      // 模型名称
	Temperature *float64 `json:"temperature" binding:"omitempty,min=0,max=2"`                        // 温度参数
	Timeout     int      `json:"timeout" binding:"omitempty,min=1,max=300"`                          // 访问超时时间（秒）
	Active      bool     `json:"active"`                                                             // 是否激活
	Description string   `json:"description" binding:"omitempty,max=256"`                            // 描述
	RPM         int      `json:"rpm" binding:"omitempty,min=1"`                                      // 每分钟最大请求数
	Weight      int      `json:"weight" binding:"omitempty,min=1"`                                   // 权重
}

// ======================== 更新操作 ========================

// UpdateModelRequest 更新AI模型配置请求
type UpdateModelRequest struct {
	Provider    *string  `json:"provider" binding:"omitempty,oneof=openai local"`  // 提供商
	APIKey      *string  `json:"api_key"`                                          // API密钥
	Endpoint    *string  `json:"endpoint" binding:"omitempty,url,startswith=http"` // API端点
	ModelName   *string  `json:"model_name"`                                       // 模型名称
	Temperature *float64 `json:"temperature" binding:"omitempty,min=0,max=2"`      // 温度参数
	Timeout     *int     `json:"timeout" binding:"omitempty,min=1,max=300"`        // 访问超时时间（秒）
	Active      *bool    `json:"active"`                                           // 是否激活
	Description *string  `json:"description" binding:"omitempty,max=256"`          // 描述
	RPM         *int     `json:"rpm" binding:"omitempty,min=1"`                    // 每分钟最大请求数
	Weight      *int     `json:"weight" binding:"omitempty,min=1"`                 // 权重
}
