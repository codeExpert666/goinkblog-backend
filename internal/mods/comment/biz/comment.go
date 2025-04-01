package biz

import (
	"context"

	articleDal "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// CommentService 评论业务逻辑层
type CommentService struct {
	CommentRepository *dal.CommentRepository
	ArticleRepository *articleDal.ArticleRepository
	Trans             util.Trans
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(ctx context.Context, userID uint, req *schema.CreateCommentRequest) (*schema.CommentResponse, error) {
	// 检查父评论是否存在
	if req.ParentID != nil && *req.ParentID > 0 {
		parentComment, err := s.CommentRepository.GetByID(ctx, *req.ParentID)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, errors.NotFound("父评论不存在")
			}
			return nil, errors.WithStack(err)
		}

		// 确保回复的是当前文章的评论
		if parentComment.ArticleID != req.ArticleID {
			return nil, errors.Conflict("当前评论与父评论不属于同一文章")
		}
	}

	// 创建评论
	comment := &schema.Comment{
		Content:   req.Content,
		AuthorID:  userID,
		ArticleID: req.ArticleID,
		ParentID:  req.ParentID,
	}

	err := s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 创建评论
		err := s.CommentRepository.Create(ctx, comment)
		if err != nil {
			return err
		}

		// 增加文章评论数
		return s.ArticleRepository.IncrementCommentCount(ctx, comment.ArticleID, 1)
	})

	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		AuthorID:  comment.AuthorID,
		ArticleID: comment.ArticleID,
		ParentID:  comment.ParentID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
	s.CommentRepository.FillAuthorInfo(ctx, response)

	return response, nil
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(ctx context.Context, userID uint, id uint) error {
	// 获取评论
	comment, err := s.CommentRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if comment.AuthorID != userID {
		return errors.Forbidden("无权限删除此评论")
	}

	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 获取回复的数量
		replies, err := s.CommentRepository.GetCommentReplies(ctx, id, 0, 0)
		if err != nil {
			return err
		}
		replyNum := replies.Total

		// 删除所有回复
		if replyNum > 0 {
			err := s.CommentRepository.DeleteByParentID(ctx, id)
			if err != nil {
				return err
			}
		}

		// 删除评论
		err = s.CommentRepository.DeleteByID(ctx, id)
		if err != nil {
			return err
		}

		// 减少文章评论数
		return s.ArticleRepository.IncrementCommentCount(ctx, comment.ArticleID, -int(replyNum+1))
	})
}

// GetCommentByID 通过ID获取评论
func (s *CommentService) GetCommentByID(ctx context.Context, id uint) (*schema.CommentResponse, error) {
	// 获取评论
	comment, err := s.CommentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		AuthorID:  comment.AuthorID,
		ArticleID: comment.ArticleID,
		ParentID:  comment.ParentID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	s.CommentRepository.FillAuthorInfo(ctx, response)

	return response, nil
}

// GetArticleComments 获取文章的评论
func (s *CommentService) GetArticleComments(ctx context.Context, articleID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	return s.CommentRepository.GetArticleComments(ctx, articleID, page, pageSize)
}

// GetCommentReplies 获取评论的回复
func (s *CommentService) GetCommentReplies(ctx context.Context, commentID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	return s.CommentRepository.GetCommentReplies(ctx, commentID, page, pageSize)
}

// GetUserComments 获取用户的评论
func (s *CommentService) GetUserComments(ctx context.Context, userID uint, page, pageSize int) (*schema.CommentPaginationResult, error) {
	return s.CommentRepository.GetUserComments(ctx, userID, page, pageSize)
}
