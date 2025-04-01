package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Comment 评论模型
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Content   string    `json:"content" gorm:"type:text;not null;comment:评论内容"`
	AuthorID  uint      `json:"author_id" gorm:"not null;comment:评论作者ID"`
	ArticleID uint      `json:"article_id" gorm:"not null;comment:评论文章ID"`
	ParentID  *uint     `json:"parent_id" gorm:"comment:父评论ID"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 表名
func (a *Comment) TableName() string {
	return config.C.FormatTableName("comment")
}

// CommentResponse 评论响应结构
type CommentResponse struct {
	ID        uint              `json:"id"`
	Content   string            `json:"content"`
	AuthorID  uint              `json:"author_id"`
	Author    string            `json:"author,omitempty"` // 作者名称
	Avatar    string            `json:"avatar,omitempty"` // 作者头像
	ArticleID uint              `json:"article_id"`
	ParentID  *uint             `json:"parent_id"`
	Replies   []CommentResponse `json:"replies,omitempty"` // 回复列表
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"`
}

// CommentPaginationResult 评论分页结果
type CommentPaginationResult struct {
	Items      []CommentResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}
