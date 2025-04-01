package dal

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	userSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetCommentDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.Comment{})
}

// CommentRepository 评论数据访问层
type CommentRepository struct {
	DB *gorm.DB
}

// Create 创建评论
func (r *CommentRepository) Create(ctx context.Context, comment *schema.Comment) error {
	result := GetCommentDB(ctx, r.DB).Create(comment)
	return errors.WithStack(result.Error)
}

// DeleteByID 通过 ID 删除评论
func (r *CommentRepository) DeleteByID(ctx context.Context, id uint) error {
	result := GetCommentDB(ctx, r.DB).Where("id = ?", id).Delete(&schema.Comment{})
	return errors.WithStack(result.Error)
}

// DeleteByParentID 通过父评论 ID 删除评论
func (r *CommentRepository) DeleteByParentID(ctx context.Context, parentID uint) error {
	result := GetCommentDB(ctx, r.DB).Where("parent_id = ?", parentID).Delete(&schema.Comment{})
	return errors.WithStack(result.Error)
}

// GetByID 通过ID获取评论
func (r *CommentRepository) GetByID(ctx context.Context, id uint) (*schema.Comment, error) {
	var comment schema.Comment
	err := GetCommentDB(ctx, r.DB).Where("id = ?", id).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("评论不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &comment, nil
}

// GetArticleComments 获取文章的评论
func (r *CommentRepository) GetArticleComments(ctx context.Context, articleID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 只获取顶级评论
	db := GetCommentDB(ctx, r.DB).
		Where("article_id = ? AND parent_id IS NULL", articleID)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (page - 1) * pageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		var replies []schema.Comment
		if err := GetCommentDB(ctx, r.DB).Where("parent_id = ?", comment.ID).
			Order("created_at ASC").
			Find(&replies).Error; err != nil {
			logging.Context(ctx).Error("获取评论回复失败", zap.Uint("comment_id", comment.ID), zap.Error(err))
			replies = []schema.Comment{} // 确保 replies 为空切片
		}

		commentResp := schema.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			AuthorID:  comment.AuthorID,
			ArticleID: comment.ArticleID,
			ParentID:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}

		// 构造回复列表
		if len(replies) > 0 {
			var replyResps []schema.CommentResponse
			for _, reply := range replies {
				replyResp := schema.CommentResponse{
					ID:        reply.ID,
					Content:   reply.Content,
					AuthorID:  reply.AuthorID,
					ArticleID: reply.ArticleID,
					ParentID:  reply.ParentID,
					CreatedAt: reply.CreatedAt,
					UpdatedAt: reply.UpdatedAt,
				}
				r.FillAuthorInfo(ctx, &replyResp)
				replyResps = append(replyResps, replyResp)
			}
			commentResp.Replies = replyResps
		}

		r.FillAuthorInfo(ctx, &commentResp)
		items = append(items, commentResp)
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}

// GetCommentReplies 获取评论的回复
func (r *CommentRepository) GetCommentReplies(ctx context.Context, commentID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	db := GetCommentDB(ctx, r.DB).
		Where("parent_id = ?", commentID)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (page - 1) * pageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(pageSize).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		item := schema.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			AuthorID:  comment.AuthorID,
			ArticleID: comment.ArticleID,
			ParentID:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}
		r.FillAuthorInfo(ctx, &item)
		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}

// GetUserComments 获取用户的评论
func (r *CommentRepository) GetUserComments(ctx context.Context, userID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	db := GetCommentDB(ctx, r.DB).
		Where("author_id = ?", userID)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (page - 1) * pageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		item := schema.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			AuthorID:  comment.AuthorID,
			ArticleID: comment.ArticleID,
			ParentID:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}
		r.FillAuthorInfo(ctx, &item)
		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}

// FillAuthorInfo 填充评论作者信息（名称、头像）
func (r *CommentRepository) FillAuthorInfo(ctx context.Context, commentResp *schema.CommentResponse) {
	commentTableName := new(schema.Comment).TableName()
	userTableName := new(userSchema.User).TableName()

	var authorInfo struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	// 关联查询评论作者信息
	err := GetCommentDB(ctx, r.DB).
		Select("u.username", "u.avatar").
		Joins(fmt.Sprintf("JOIN %s u ON %s.author_id = u.id", userTableName, commentTableName)).
		Where(fmt.Sprintf("%s.author_id = ?", commentTableName), commentResp.AuthorID).
		First(&authorInfo).Error

	// 将信息填入传入的评论响应结构体中
	if err == nil {
		commentResp.Author = authorInfo.Username
		commentResp.Avatar = authorInfo.Avatar
	} else { // 出错记录日志
		logging.Context(ctx).Error("获取评论作者信息失败",
			zap.Uint("comment_id", commentResp.ID),
			zap.Uint("author_id", commentResp.AuthorID),
			zap.Error(errors.WithStack(err)))
	}
}
