package biz

import (
	"context"
	"time"

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
	// 初始化评论对象
	comment := &schema.Comment{
		Content:   req.Content,
		AuthorID:  userID,
		ArticleID: req.ArticleID,
		Level:     1, // 默认为顶级评论
	}

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

		// 检查父评论是否已审核通过
		if parentComment.Status != schema.CommentStatusApproved {
			return nil, errors.BadRequest("不能回复未审核通过的评论")
		}

		// 设置父评论ID
		comment.ParentID = req.ParentID

		// 设置根评论ID和层级
		if parentComment.RootID != nil {
			// 如果父评论不是根评论，使用父评论的根评论ID
			comment.RootID = parentComment.RootID
			comment.Level = parentComment.Level + 1
		} else {
			// 如果父评论是根评论，则父评论ID即为根评论ID
			comment.RootID = req.ParentID
			comment.Level = 2
		}

		// 检查评论层级是否过深
		maxLevel := 20 // 最大支持20层评论
		if comment.Level > maxLevel {
			return nil, errors.BadRequest("评论层级过深，请直接回复根评论")
		}
	}

	// 管理员发表的评论自动通过审核
	if util.FromIsAdminUser(ctx) {
		now := time.Now()
		comment.Status = schema.CommentStatusApproved
		comment.ReviewedAt = &now
		comment.ReviewerID = &userID
		comment.ReviewRemark = "管理员自动通过"
	} else {
		comment.Status = schema.CommentStatusPending
	}

	err := s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 创建评论
		err := s.CommentRepository.Create(ctx, comment)
		if err != nil {
			return err
		}

		// 如果评论状态为已通过，则增加文章评论数
		if comment.Status == schema.CommentStatusApproved {
			return s.ArticleRepository.IncrementCommentCount(ctx, comment.ArticleID, 1)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.CommentResponse{
		ID:           comment.ID,
		Content:      comment.Content,
		AuthorID:     comment.AuthorID,
		ArticleID:    comment.ArticleID,
		ParentID:     comment.ParentID,
		RootID:       comment.RootID,
		Level:        comment.Level,
		Status:       comment.Status,
		ReviewedAt:   comment.ReviewedAt,
		ReviewerID:   comment.ReviewerID,
		ReviewRemark: comment.ReviewRemark,
		ReplyCount:   0, // 新评论没有回复
		CreatedAt:    comment.CreatedAt,
	}

	s.CommentRepository.FillAuthorInfo(ctx, response)
	s.CommentRepository.FillReviewerInfo(ctx, response)
	s.CommentRepository.FillArticleInfo(ctx, response)
	s.CommentRepository.FillParentCommentInfo(ctx, response)

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
		// 递归删除所有回复，replyNum 为删除的已审核通过的回复数量
		replyNum, err := s.CommentRepository.DeleteByParentID(ctx, id)
		if err != nil {
			return err
		}

		// 删除评论
		err = s.CommentRepository.DeleteByID(ctx, id)
		if err != nil {
			return err
		}

		// 减少文章评论数
		if comment.Status == schema.CommentStatusApproved {
			return s.ArticleRepository.IncrementCommentCount(ctx, comment.ArticleID, -int(replyNum+1))
		}
		// 评论未审核通过，不占用文章评论数，同时也不会有回复，故不需要减少文章评论数
		return nil
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
		ID:           comment.ID,
		Content:      comment.Content,
		AuthorID:     comment.AuthorID,
		ArticleID:    comment.ArticleID,
		ParentID:     comment.ParentID,
		RootID:       comment.RootID,
		Level:        comment.Level,
		Status:       comment.Status,
		ReviewedAt:   comment.ReviewedAt,
		ReviewerID:   comment.ReviewerID,
		ReviewRemark: comment.ReviewRemark,
		CreatedAt:    comment.CreatedAt,
	}

	s.CommentRepository.FillAuthorInfo(ctx, response)
	s.CommentRepository.FillReviewerInfo(ctx, response)
	s.CommentRepository.FillArticleInfo(ctx, response)
	s.CommentRepository.FillParentCommentInfo(ctx, response)
	response.ReplyCount, _ = s.CommentRepository.CountReplies(ctx, id)

	return response, nil
}

// GetArticleComments 获取文章的评论
func (s *CommentService) GetArticleComments(ctx context.Context, articleID uint, req *schema.ArticleCommentsRequest) (*schema.CommentPaginationResult, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.SortByCreate == "" {
		req.SortByCreate = "desc"
	}

	// 管理员可以查看待审核评论
	if util.FromIsAdminUser(ctx) {
		req.IncludePending = true
	}

	// 管理员可以查看待审核评论
	return s.CommentRepository.GetArticleComments(ctx, articleID, req)
}

// GetCommentReplies 获取评论的回复，扁平化列表
func (s *CommentService) GetCommentReplies(ctx context.Context, commentID uint, req *schema.CommentRepliesRequest) (*schema.CommentPaginationResult, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.MaxDepth < 0 {
		req.MaxDepth = 0
	}
	if req.SortByCreate == "" {
		req.SortByCreate = "asc"
	}

	// 管理员可以查看待审核评论
	if util.FromIsAdminUser(ctx) {
		req.IncludePending = true
	}

	return s.CommentRepository.GetCommentReplies(ctx, commentID, req)
}

// GetUserComments 获取用户的评论
func (s *CommentService) GetUserComments(ctx context.Context, userID uint, req *schema.UserCommentsRequest) (*schema.CommentPaginationResult, error) {
	return s.CommentRepository.GetUserComments(ctx, userID, req)
}

// GetCommentsForReview 获取评论审核列表
func (s *CommentService) GetCommentsForReview(ctx context.Context, req *schema.CommentReviewListRequest) (*schema.CommentPaginationResult, error) {
	return s.CommentRepository.GetCommentsForReview(ctx, req)
}

// ReviewComment 审核评论
func (s *CommentService) ReviewComment(ctx context.Context, reviewerID uint, req *schema.ReviewCommentRequest) error {
	// 获取评论
	comment, err := s.CommentRepository.GetByID(ctx, req.CommentID)
	if err != nil {
		return err
	}

	if comment.Status != schema.CommentStatusPending {
		return errors.BadRequest("评论已审核，不能重复审核")
	}

	// 更新评论审核状态
	now := time.Now()
	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 更新评论状态
		err := s.CommentRepository.UpdateCommentStatus(ctx, reviewerID, req, &now)
		if err != nil {
			return err
		}

		// 如果评论审核通过，需要增加文章评论数
		if req.Status == schema.CommentStatusApproved {
			// 增加文章评论数
			return s.ArticleRepository.IncrementCommentCount(ctx, comment.ArticleID, 1)
		}

		return nil
	})
}
