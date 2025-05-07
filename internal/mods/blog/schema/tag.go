package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Tag 标签模型
type Tag struct {
	ID        uint      `json:"id" gorm:"index;primaryKey"`
	Name      string    `json:"name" gorm:"size:50;not null;uniqueIndex;comment:标签名称"`
	CreatedAt time.Time `json:"created_at" gorm:"index;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;comment:更新时间"`
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

// TagQueryParams 标签查询请求
type TagQueryParams struct {
	Page               int    `form:"page" binding:"omitempty,min=1"`
	PageSize           int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortByID           string `form:"sort_by_id" binding:"omitempty,oneof=desc asc"`
	SortByArticleCount string `form:"sort_by_article_count" binding:"omitempty,oneof=desc asc"`
	SortByCreate       string `form:"sort_by_create" binding:"omitempty,oneof=desc asc"`
	SortByUpdate       string `form:"sort_by_update" binding:"omitempty,oneof=desc asc"`
}

// TagPaginationResult 标签分页结果
type TagPaginationResult struct {
	Items                  []TagResponse `json:"items"`
	TotalTags              int64         `json:"total_tags"`
	TotalArticles          int64         `json:"total_articles"`
	TagsWithArticle        int64         `json:"tags_with_article"`
	TagNameWithMostArticle string        `json:"tag_name_with_most_article"`
	MostArticleCounts      int64         `json:"most_article_counts"`
	Page                   int           `json:"page"`
	PageSize               int           `json:"page_size"`
	TotalPages             int           `json:"total_pages"`
}
