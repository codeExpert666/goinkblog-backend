package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Article 文章模型
type Article struct {
	ID            uint      `json:"id" gorm:"index;primaryKey"`
	Title         string    `json:"title" gorm:"size:255;not null;comment:文章标题"`
	Content       string    `json:"content" gorm:"type:text;not null;comment:文章内容"`
	Summary       string    `json:"summary" gorm:"type:text;comment:文章摘要"`
	AuthorID      uint      `json:"author_id" gorm:"index;not null;comment:作者ID"`
	CategoryID    *uint     `json:"category_id" gorm:"index;comment:分类ID"`
	Cover         string    `json:"cover" gorm:"size:255;comment:封面图URL"`
	Status        string    `json:"status" gorm:"size:20;not null;default:draft;comment:状态"`
	ViewCount     int       `json:"view_count" gorm:"default:0;comment:浏览次数"`
	LikeCount     int       `json:"like_count" gorm:"default:0;comment:点赞次数"`
	CommentCount  int       `json:"comment_count" gorm:"default:0;comment:评论次数"`
	FavoriteCount int       `json:"favorite_count" gorm:"default:0;comment:收藏次数"`
	CreatedAt     time.Time `json:"created_at" gorm:"index;comment:创建时间"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

func (a *Article) TableName() string {
	return config.C.FormatTableName("article")
}

// ArticleResponse 文章响应结构
type ArticleResponse struct {
	ID            uint                 `json:"id"`
	Title         string               `json:"title"`
	Content       string               `json:"content"`
	Summary       string               `json:"summary"`
	AuthorID      uint                 `json:"author_id"`
	Author        string               `json:"author,omitempty"`        // 作者名称
	AuthorAvatar  string               `json:"author_avatar,omitempty"` // 作者头像
	CategoryID    *uint                `json:"category_id"`
	CategoryName  string               `json:"category_name,omitempty"` // 分类名称
	Cover         string               `json:"cover"`
	Status        string               `json:"status"`
	ViewCount     int                  `json:"view_count"`
	LikeCount     int                  `json:"like_count"`
	CommentCount  int                  `json:"comment_count"`
	FavoriteCount int                  `json:"favorite_count"`
	Tags          []uint               `json:"tags,omitempty"` // 标签列表
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	Interactions  *InteractionResponse `json:"interactions,omitempty"` // 交互状态
}

// ArticleListItem 文章列表项
type ArticleListItem struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	AuthorID        uint      `json:"author_id"`
	Author          string    `json:"author,omitempty"`        // 作者名称
	AuthorAvatar    string    `json:"author_avatar,omitempty"` // 作者头像
	CategoryID      *uint     `json:"category_id"`
	CategoryName    string    `json:"category_name,omitempty"` // 分类名称
	Cover           string    `json:"cover"`
	Status          string    `json:"status"`
	ViewCount       int       `json:"view_count"`
	LikeCount       int       `json:"like_count"`
	CommentCount    int       `json:"comment_count"`
	FavoriteCount   int       `json:"favorite_count"`
	Tags            []uint     `json:"tags,omitempty"` // 标签列表
	CreatedAt       time.Time `json:"created_at"`
	InteractionTime time.Time `json:"interaction_time,omitempty"` // 交互时间（用于历史记录等）
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary"`
	CategoryID *uint  `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
	Cover      string `json:"cover"`
	Status     string `json:"status" binding:"required,oneof=published draft"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CategoryID *uint  `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
	Cover      string `json:"cover"`
	Status     string `json:"status" binding:"omitempty,oneof=published draft"`
}

// ArticleQueryParams 文章查询参数
type ArticleQueryParams struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	CategoryIDs []uint `form:"category_ids" binding:"omitempty,dive,min=1"`
	TagIDs      []uint `form:"tag_ids" binding:"omitempty,dive,min=1"`
	Author      string `form:"author"`
	Status      string `form:"status" binding:"omitempty,oneof=published draft"`
	SortBy      string `form:"sort_by" binding:"omitempty,oneof=newest views likes favorites comments"`
	Keyword     string `form:"keyword"`
	TimeRange   string `form:"time_range" binding:"omitempty,oneof=today week month year all"`
}

// ArticlePaginationResult 文章分页结果
type ArticlePaginationResult struct {
	Items      []*ArticleListItem `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// CoverResponse 封面图片响应
type CoverResponse struct {
	URL string `json:"url"`
}

// ArticleInteractionResponse 文章交互响应
type ArticleInteractionResponse struct {
	Interacted bool `json:"interacted"`
}
