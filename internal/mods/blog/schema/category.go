package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Category 分类模型
type Category struct {
	ID          uint      `json:"id" gorm:"index;primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;unique;comment:分类名称"`
	Description string    `json:"description" gorm:"type:text;comment:分类描述"`
	CreatedAt   time.Time `json:"created_at" gorm:"index;comment:创建时间"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;comment:更新时间"`
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

// CategoryQueryParams 分类查询请求
type CategoryQueryParams struct {
	Page               int    `form:"page" binding:"omitempty,min=1"`
	PageSize           int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortByID           string `form:"sort_by_id" binding:"omitempty,oneof=desc asc"`
	SortByArticleCount string `form:"sort_by_article_count" binding:"omitempty,oneof=desc asc"`
	SortByCreate       string `form:"sort_by_create" binding:"omitempty,oneof=desc asc"`
	SortByUpdate       string `form:"sort_by_update" binding:"omitempty,oneof=desc asc"`
}

// CategoryPaginationResult 分类分页结果
type CategoryPaginationResult struct {
	Items                       []CategoryResponse `json:"items"`
	TotalCategories             int64              `json:"total_categories"`                // 分类总数
	TotalArticles               int64              `json:"total_articles"`                  // 有分类的文章总数
	CategoriesWithArticle       int64              `json:"categories_with_article"`         // 有文章的分类数量
	CategoryNameWithMostArticle string             `json:"category_name_with_most_article"` // 文章数量最多的分类名称
	MostArticleCounts           int64              `json:"most_article_counts"`             // 分类中最多的文章数量
	Page                        int                `json:"page"`
	PageSize                    int                `json:"page_size"`
	TotalPages                  int                `json:"total_pages"`
}
