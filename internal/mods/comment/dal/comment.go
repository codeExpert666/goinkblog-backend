package dal

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	userDal "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	articleDal "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	articleSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
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

// DeleteByParentID 通过父评论 ID 删除评论（递归删除），并返回删除的评论（审核通过）数量
func (r *CommentRepository) DeleteByParentID(ctx context.Context, parentID uint) (int, error) {
	// 先获取所有子评论
	var comments []schema.Comment
	if err := GetCommentDB(ctx, r.DB).Where("parent_id = ?", parentID).Find(&comments).Error; err != nil {
		return 0, errors.WithStack(err)
	}

	if len(comments) == 0 {
		return 0, nil
	}

	// 递归删除子评论
	var count int
	for _, comment := range comments {
		c, err := r.DeleteByParentID(ctx, comment.ID)
		if err != nil {
			return 0, err
		}
		if comment.Status == schema.CommentStatusApproved {
			c += 1
		}
		count += c
	}

	err := GetCommentDB(ctx, r.DB).Where("parent_id = ?", parentID).Delete(&schema.Comment{}).Error
	return count, errors.WithStack(err)
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

// GetArticleComments 获取文章的评论，支持多级嵌套
func (r *CommentRepository) GetArticleComments(ctx context.Context, articleID uint, req *schema.ArticleCommentsRequest) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 只获取文章的顶级评论
	db := GetCommentDB(ctx, r.DB).
		Where("article_id = ? AND level = 1", articleID)

	// 只获取已审核通过的评论，除非管理员指定包含未审核评论
	if !req.IncludePending {
		db = db.Where("status = ?", schema.CommentStatusApproved)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用排序
	if req.SortByCreate == "asc" {
		db = db.Order("created_at ASC")
	} else {
		db = db.Order("created_at DESC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(req.PageSize).Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		commentResp := schema.CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			AuthorID:  comment.AuthorID,
			ArticleID: comment.ArticleID,
			ParentID:  comment.ParentID,
			RootID:    comment.RootID,
			Level:     comment.Level,
			Status:    comment.Status,
			CreatedAt: comment.CreatedAt,
		}

		// 填充评论作者信息
		r.FillAuthorInfo(ctx, &commentResp)

		// 计算回复数量
		replyCount, _ := r.CountReplies(ctx, comment.ID)
		commentResp.ReplyCount = replyCount

		items = append(items, commentResp)
	}

	result.Items = items
	result.Total = total
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &result, nil
}

// CountReplies 计算评论的回复数量
func (r *CommentRepository) CountReplies(ctx context.Context, commentID uint) (int64, error) {
	var count int64
	err := GetCommentDB(ctx, r.DB).
		Where("parent_id = ? OR root_id = ?", commentID, commentID).
		Where("status = ?", schema.CommentStatusApproved).
		Where("id != ?", commentID). // 排除自身
		Count(&count).Error

	return count, errors.WithStack(err)
}

// GetCommentReplies 获取评论的回复，扁平化列表
func (r *CommentRepository) GetCommentReplies(ctx context.Context, commentID uint, req *schema.CommentRepliesRequest) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 构建查询条件
	db := GetCommentDB(ctx, r.DB)

	// 根据查询方式不同调整查询条件
	if req.IncludeReplies {
		// 多级评论查询：获取所有以该评论为根的回复
		db = db.Where("(parent_id = ? OR root_id = ?) AND id != ?", commentID, commentID, commentID)
		if req.MaxDepth > 0 { // 如果指定了最大递归深度，则限制查询结果的层级
			// 首先获取当前评论的层级
			currentComment, err := r.GetByID(ctx, commentID)
			if err != nil {
				return nil, err
			}
			// 根据当前评论的层级，计算最大层级限制
			db = db.Where("level <= ?", currentComment.Level+req.MaxDepth)
		}
	} else {
		// 传统方式：只获取直接回复该评论的评论
		db = db.Where("parent_id = ?", commentID)
	}

	// 只获取已审核通过的评论，除非管理员指定包含未审核评论
	if !req.IncludePending {
		db = db.Where("status = ?", schema.CommentStatusApproved)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用排序
	if req.SortByCreate == "desc" {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(req.PageSize).Find(&comments).Error; err != nil {
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
			RootID:    comment.RootID,
			Level:     comment.Level,
			Status:    comment.Status,
			CreatedAt: comment.CreatedAt,
		}

		// 填充基本信息
		r.FillAuthorInfo(ctx, &item)
		r.FillParentCommentInfo(ctx, &item)

		// 计算回复数量
		replyCount, _ := r.CountReplies(ctx, comment.ID)
		item.ReplyCount = replyCount

		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &result, nil
}

// GetUserComments 获取用户的评论
func (r *CommentRepository) GetUserComments(ctx context.Context, userID uint, req *schema.UserCommentsRequest) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 用户可以看到自己的所有评论（包括待审核的、通过的、拒绝的）
	db := GetCommentDB(ctx, r.DB).Where("author_id = ?", userID)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		item := schema.CommentResponse{
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
		r.FillArticleInfo(ctx, &item)
		r.FillParentCommentInfo(ctx, &item)
		r.FillReviewerInfo(ctx, &item)
		item.ReplyCount, _ = r.CountReplies(ctx, comment.ID)
		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &result, nil
}

// GetCommentsForReview 获取评论审核列表
func (r *CommentRepository) GetCommentsForReview(ctx context.Context, req *schema.CommentReviewListRequest) (*schema.CommentPaginationResult, error) {
	var result schema.CommentPaginationResult

	// 默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	db := GetCommentDB(ctx, r.DB)

	// 应用基本筛选条件
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	if req.ArticleID != nil {
		db = db.Where("article_id = ?", *req.ArticleID)
	}

	if req.AuthorID != nil {
		db = db.Where("author_id = ?", *req.AuthorID)
	}

	if req.ParentID != nil {
		db = db.Where("parent_id = ?", *req.ParentID)
	}

	if req.RootID != nil {
		db = db.Where("root_id = ?", *req.RootID)
	}

	if req.Level != nil {
		db = db.Where("level = ?", *req.Level)
	}

	if req.Keyword != "" {
		db = db.Where("content LIKE ?", "%"+req.Keyword+"%")
	}

	// 应用时间范围筛选
	if req.CreateStartTime != nil {
		db = db.Where("created_at >= ?", req.CreateStartTime)
	}

	if req.CreateEndTime != nil {
		db = db.Where("created_at <= ?", req.CreateEndTime)
	}

	if req.ReviewStartTime != nil {
		db = db.Where("reviewed_at >= ?", req.ReviewStartTime)
	}

	if req.ReviewEndTime != nil {
		db = db.Where("reviewed_at <= ?", req.ReviewEndTime)
	}

	// 审核人员筛选
	if req.ReviewerID != nil {
		db = db.Where("reviewer_id = ?", *req.ReviewerID)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用排序
	if req.SortBy == "review" {
		// 按审核时间排序
		if req.SortOrder == "asc" {
			db = db.Order("reviewed_at ASC")
		} else {
			db = db.Order("reviewed_at DESC")
		}
	} else {
		// 默认按创建时间排序
		if req.SortOrder == "asc" {
			db = db.Order("created_at ASC")
		} else {
			db = db.Order("created_at DESC")
		}
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	var comments []schema.Comment
	if err := db.Offset(offset).Limit(req.PageSize).Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CommentResponse
	for _, comment := range comments {
		item := schema.CommentResponse{
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

		// 填充关联信息
		r.FillAuthorInfo(ctx, &item)
		r.FillArticleInfo(ctx, &item)
		r.FillParentCommentInfo(ctx, &item)
		r.FillReviewerInfo(ctx, &item)

		// 计算回复数量
		replyCount, _ := r.CountReplies(ctx, comment.ID)
		item.ReplyCount = replyCount

		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &result, nil
}

// UpdateCommentStatus 更新评论状态
func (r *CommentRepository) UpdateCommentStatus(ctx context.Context, reviewerID uint, req *schema.ReviewCommentRequest, reviewedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":        req.Status,
		"reviewer_id":   reviewerID,
		"review_remark": req.ReviewRemark,
		"reviewed_at":   reviewedAt,
	}

	err := GetCommentDB(ctx, r.DB).Where("id = ?", req.CommentID).Updates(updates).Error

	return errors.WithStack(err)
}

// FillAuthorInfo 填充评论作者信息（名称、头像）
func (r *CommentRepository) FillAuthorInfo(ctx context.Context, commentResp *schema.CommentResponse) {
	var authorInfo struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	// 关联查询评论作者信息
	err := userDal.GetUserDB(ctx, r.DB).
		Select("username", "avatar").
		Where("id = ?", commentResp.AuthorID).
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

// FillReviewerInfo 填充评论审核人员信息
func (r *CommentRepository) FillReviewerInfo(ctx context.Context, commentResp *schema.CommentResponse) {
	if commentResp.ReviewerID == nil {
		return
	}

	var reviewerInfo struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	// 查询审核人员信息
	err := userDal.GetUserDB(ctx, r.DB).
		Select("username", "avatar").
		Where("id = ?", *commentResp.ReviewerID).
		First(&reviewerInfo).Error

	// 将信息填入传入的评论响应结构体中
	if err == nil {
		commentResp.ReviewerName = reviewerInfo.Username
		commentResp.ReviewerAvatar = reviewerInfo.Avatar
	} else { // 出错记录日志
		logging.Context(ctx).Error("获取审核人员信息失败",
			zap.Uint("comment_id", commentResp.ID),
			zap.Uint("reviewer_id", *commentResp.ReviewerID),
			zap.Error(errors.WithStack(err)))
	}
}

func (r *CommentRepository) FillArticleInfo(ctx context.Context, commentResp *schema.CommentResponse) {
	var articleInfo struct {
		Title string `json:"title"`
	}

	// 查询文章标题
	err := articleDal.GetArticleDB(ctx, r.DB).Model(&articleSchema.Article{}).
		Select("title").
		Where("id = ?", commentResp.ArticleID).
		First(&articleInfo).Error

	// 将信息填入传入的评论响应结构体中
	if err == nil {
		commentResp.ArticleTitle = articleInfo.Title
	} else { // 出错记录日志
		logging.Context(ctx).Error("获取文章标题失败",
			zap.Uint("comment_id", commentResp.ID),
			zap.Uint("article_id", commentResp.ArticleID),
			zap.Error(errors.WithStack(err)))
	}
}

func (r *CommentRepository) FillParentCommentInfo(ctx context.Context, commentResp *schema.CommentResponse) {
	if commentResp.ParentID == nil {
		return
	}

	var parentCommentInfo struct {
		Content  string `json:"content"`
		AuthorID uint   `json:"author_id"`
	}

	// 查询父评论内容和作者
	err := GetCommentDB(ctx, r.DB).
		Select("content", "author_id").
		Where("id = ?", *commentResp.ParentID).
		First(&parentCommentInfo).Error

	// 将信息填入传入的评论响应结构体中
	if err == nil {
		commentResp.ParentContent = parentCommentInfo.Content

		// 获取父评论作者名称
		var authorInfo struct {
			Username string `json:"username"`
		}

		err := userDal.GetUserDB(ctx, r.DB).
			Select("username").
			Where("id = ?", parentCommentInfo.AuthorID).
			First(&authorInfo).Error

		if err == nil {
			commentResp.ParentAuthor = authorInfo.Username
		} else { // 出错记录日志
			logging.Context(ctx).Error("获取父评论作者信息失败",
				zap.Uint("comment_id", commentResp.ID),
				zap.Uint("parent_id", *commentResp.ParentID),
				zap.Uint("author_id", parentCommentInfo.AuthorID),
				zap.Error(errors.WithStack(err)))
		}
	} else { // 出错记录日志
		logging.Context(ctx).Error("获取父评论信息失败",
			zap.Uint("comment_id", commentResp.ID),
			zap.Uint("parent_id", *commentResp.ParentID),
			zap.Error(errors.WithStack(err)))
	}
}
