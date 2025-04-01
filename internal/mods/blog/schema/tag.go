package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Tag 标签模型
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:50;not null;uniqueIndex;comment:标签名称"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 表名
func (a *Tag) TableName() string {
	return config.C.FormatTableName("tag")
}

// TagResponse 标签响应结构
type TagResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	ArticleCount int       `json:"article_count"` // 文章数量
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// TagPaginationResult 标签分页结果
type TagPaginationResult struct {
	Items      []TagResponse `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}
