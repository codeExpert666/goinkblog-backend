package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// Comment 评论模型
type Comment struct {
	ID           uint       `json:"id" gorm:"index;primaryKey"`
	Content      string     `json:"content" gorm:"type:text;not null;comment:评论内容"`
	AuthorID     uint       `json:"author_id" gorm:"not null;index;comment:评论作者ID"`
	ArticleID    uint       `json:"article_id" gorm:"not null;index;comment:评论文章ID"`
	ParentID     *uint      `json:"parent_id" gorm:"index;comment:父评论ID"`
	RootID       *uint      `json:"root_id" gorm:"index;comment:根评论ID,用于快速查询评论树"`
	Level        int        `json:"level" gorm:"type:tinyint;default:1;comment:评论层级,1为顶级评论"`
	Status       int        `json:"status" gorm:"type:tinyint;default:0;comment:审核状态,0-待审核,1-已通过,2-已拒绝"`
	ReviewedAt   *time.Time `json:"reviewed_at" gorm:"index;comment:审核时间"`
	ReviewerID   *uint      `json:"reviewer_id" gorm:"comment:审核员ID"`
	ReviewRemark string     `json:"review_remark" gorm:"type:varchar(255);comment:审核备注"`
	CreatedAt    time.Time  `json:"created_at" gorm:"index;comment:创建时间"`
}

// 评论状态常量
const (
	CommentStatusPending  = 0 // 待审核
	CommentStatusApproved = 1 // 已通过
	CommentStatusRejected = 2 // 已拒绝
)

// TableName 表名
func (a *Comment) TableName() string {
	return config.C.FormatTableName("comment")
}

// CommentResponse 评论响应结构
type CommentResponse struct {
	ID             uint       `json:"id"`
	Content        string     `json:"content"`
	AuthorID       uint       `json:"author_id"`
	Author         string     `json:"author,omitempty"` // 作者名称
	Avatar         string     `json:"avatar,omitempty"` // 作者头像
	ArticleID      uint       `json:"article_id"`
	ArticleTitle   string     `json:"article_title,omitempty"` // 文章标题
	ParentID       *uint      `json:"parent_id"`
	RootID         *uint      `json:"root_id,omitempty"`         // 根评论ID
	Level          int        `json:"level"`                     // 评论层级
	ParentContent  string     `json:"parent_content,omitempty"`  // 父评论内容
	ParentAuthor   string     `json:"parent_author,omitempty"`   // 父评论作者
	Status         int        `json:"status"`                    // 审核状态
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`     // 审核时间
	ReviewerID     *uint      `json:"reviewer_id,omitempty"`     // 审核员ID
	ReviewerName   string     `json:"reviewer_name,omitempty"`   // 审核员名称
	ReviewerAvatar string     `json:"reviewer_avatar,omitempty"` // 审核员头像
	ReviewRemark   string     `json:"review_remark,omitempty"`   // 审核备注
	ReplyCount     int64      `json:"reply_count"`               // 回复数量
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"`
}

// ReviewCommentRequest 评论审核请求
type ReviewCommentRequest struct {
	CommentID    uint   `json:"comment_id" binding:"required"`       // 评论ID
	Status       int    `json:"status" binding:"required,oneof=1 2"` // 审核状态：1-通过，2-拒绝
	ReviewRemark string `json:"review_remark"`                       // 审核备注
}

// CommentReviewListRequest 评论审核列表查询请求
type CommentReviewListRequest struct {
	// 基本筛选
	ArticleID *uint  `json:"article_id" form:"article_id"`                         // 文章ID
	AuthorID  *uint  `json:"author_id" form:"author_id"`                           // 评论作者ID
	ParentID  *uint  `json:"parent_id" form:"parent_id"`                           // 父评论ID
	RootID    *uint  `json:"root_id" form:"root_id"`                               // 根评论ID
	Level     *int   `json:"level" form:"level"`                                   // 评论层级
	Keyword   string `json:"keyword" form:"keyword"`                               // 关键词搜索
	Status    *int   `json:"status" form:"status" binding:"omitempty,oneof=0 1 2"` // 评论状态：0-待审核，1-已通过，2-已拒绝，不传则查询所有状态

	// 时间范围筛选
	CreateStartTime *time.Time `json:"create_start_time" form:"create_start_time"` // 创建开始时间
	CreateEndTime   *time.Time `json:"create_end_time" form:"create_end_time"`     // 创建结束时间
	ReviewStartTime *time.Time `json:"review_start_time" form:"review_start_time"` // 审核开始时间
	ReviewEndTime   *time.Time `json:"review_end_time" form:"review_end_time"`     // 审核结束时间

	// 审核人员筛选
	ReviewerID *uint `json:"reviewer_id" form:"reviewer_id"` // 审核人员ID

	// 排序选项
	SortBy    string `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=create review"`  // 排序字段：create-创建时间，review-审核时间
	SortOrder string `json:"sort_order" form:"sort_order" binding:"omitempty,oneof=desc asc"` // 排序方式：desc-降序，asc-升序

	// 分页选项
	Page     int `json:"page" form:"page" binding:"required,min=1"`                   // 页码
	PageSize int `json:"page_size" form:"page_size" binding:"required,min=1,max=100"` // 页大小
}

// CommentPaginationResult 评论分页结果
type CommentPaginationResult struct {
	Items      []CommentResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// PaginationRequest 分页请求基础结构体
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"omitempty,min=1"`                  // 页码
	PageSize int `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=30"` // 页容量
}

// ArticleCommentsRequest 文章评论查询请求
type ArticleCommentsRequest struct {
	PaginationRequest        // 嵌入分页请求基础结构
	IncludePending    bool   `json:"-" form:"-"`                                                              // 是否包含待审核评论（仅管理员可用）
	SortByCreate      string `json:"sort_by_create" form:"sort_by_create" binding:"omitempty,oneof=asc desc"` // 排序方式
}

// CommentRepliesRequest 评论回复查询请求
type CommentRepliesRequest struct {
	PaginationRequest        // 嵌入分页请求基础结构
	IncludeReplies    bool   `json:"include_replies" form:"include_replies"`                                  // 是否包含回复
	MaxDepth          int    `json:"max_depth" form:"max_depth" binding:"omitempty,min=0"`                    // 最大递归深度，0表示不限制
	IncludePending    bool   `json:"-" form:"-"`                                                              // 是否包含待审核评论（仅管理员可用）
	SortByCreate      string `json:"sort_by_create" form:"sort_by_create" binding:"omitempty,oneof=asc desc"` // 排序方式
}

// CommentTreeRequest 评论树查询请求
type CommentTreeRequest struct {
	MaxDepth int `json:"max_depth" form:"max_depth" binding:"omitempty,min=0"` // 最大递归深度，0表示不限制
}

// UserCommentsRequest 用户评论查询请求
type UserCommentsRequest struct {
	PaginationRequest // 嵌入分页请求基础结构
}
