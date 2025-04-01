package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Category 分类模型
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;uniqueIndex;comment:分类名称"`
	Description string    `json:"description" gorm:"type:text;comment:分类描述"`
	CreatedAt   time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 表名
func (a *Category) TableName() string {
	return config.C.FormatTableName("category")
}

// CategoryResponse 分类响应结构
type CategoryResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ArticleCount int       `json:"article_count"` // 文章数量
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CategoryPaginationResult 分类分页结果
type CategoryPaginationResult struct {
	Items      []CategoryResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
