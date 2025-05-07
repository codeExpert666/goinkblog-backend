package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`                   // 日志ID
	Level     string    `gorm:"size:20;index;" json:"level"`                     // 日志级别
	TraceID   string    `gorm:"size:64;index;" json:"trace_id,omitempty"`        // 链路ID
	UserID    uint      `gorm:"index;" json:"user_id,omitempty"`                 // 用户ID
	Tag       string    `gorm:"size:32;index;" json:"tag,omitempty"`             // 日志标签
	Message   string    `gorm:"size:1024;" json:"message,omitempty"`             // 日志消息
	Stack     string    `gorm:"type:text;" json:"stack,omitempty"`               // 错误堆栈
	Data      string    `gorm:"type:text;" json:"data,omitempty"`                // 日志数据
	CreatedAt time.Time `gorm:"index;" json:"created_at"`                        // 创建时间
	Username  string    `json:"username,omitempty" gorm:"<-:false;-:migration;"` // 用户名称（不存储在数据库中）
}

// TableName 表名
func (a *Logger) TableName() string {
	return config.C.FormatTableName("logger")
}

// LoggerQueryParams 日志查询参数
type LoggerQueryParams struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Level        string `form:"level" binding:"omitempty,oneof=debug info warn error dpanic panic fatal"`
	TraceID      string `form:"trace_id"`
	LikeUsername string `form:"username"`
	Tag          string `form:"tag" binding:"omitempty,oneof=main recovery request login logout system operate ai"`
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
