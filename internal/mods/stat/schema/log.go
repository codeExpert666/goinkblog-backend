package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`         // 日志ID
	Level     string    `gorm:"size:20;index;" json:"level"`           // 日志级别
	TraceID   string    `gorm:"size:64;index;" json:"trace_id"`        // 链路ID
	UserID    uint      `gorm:"index;" json:"user_id"`                 // 用户ID
	Tag       string    `gorm:"size:32;index;" json:"tag"`             // 日志标签
	Message   string    `gorm:"size:1024;" json:"message"`             // 日志消息
	Stack     string    `gorm:"type:text;" json:"stack"`               // 错误堆栈
	Data      string    `gorm:"type:text;" json:"data"`                // 日志数据
	CreatedAt time.Time `gorm:"index;" json:"created_at"`              // 创建时间
	Username  string    `json:"username" gorm:"<-:false;-:migration;"` // 用户名称（不存储在数据库中）
}

// TableName 表名
func (a *Logger) TableName() string {
	return config.C.FormatTableName("logger")
}

// LoggerQueryParams 日志查询参数
type LoggerQueryParams struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=30"`
	Level        string `form:"level"`
	TraceID      string `form:"trace_id"`
	LikeUsername string `form:"username"`
	Tag          string `form:"tag"`
	LikeMessage  string `form:"message"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
}

// LoggerPaginationResult 日志分页结果
type LoggerPaginationResult struct {
	Items      []Logger `json:"items"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// ArticleStatisticResponse 统计数据响应结构
type ArticleStatisticResponse struct {
	TotalArticles  int64 `json:"totalArticles"`
	TotalViews     int64 `json:"totalViews"`
	TotalLikes     int64 `json:"totalComments"`
	TotalComments  int64 `json:"totalLikes"`
	TotalFavorites int64 `json:"totalFavorites"`
}

// APIAccessTrendItem API访问趋势数据项
type APIAccessTrendItem struct {
	Date             string `json:"date"`
	TotalCount       int64  `json:"total_count"`
	SuccessCount     int64  `json:"success_count"`
	ClientErrorCount int64  `json:"client_error_count"`
	ServerErrorCount int64  `json:"server_error_count"`
}

// UserActivityTrendItem 用户活跃度趋势数据项
type UserActivityTrendItem struct {
	Date      string `json:"date"`
	UserCount int64  `json:"user_count"`
}

// CategoryDistItem 分类分布数据项
type CategoryDistItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
