package biz

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	userDal "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ArticleService 文章业务逻辑层
type ArticleService struct {
	ArticleRepository     *dal.ArticleRepository
	CategoryRepository    *dal.CategoryRepository
	TagRepository         *dal.TagRepository
	ArticleTagRepository  *dal.ArticleTagRepository
	InteractionRepository *dal.InteractionRepository
	UserRepository        *userDal.UserRepository
	Trans                 util.Trans
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(ctx context.Context, userID uint, req *schema.CreateArticleRequest) (*schema.ArticleResponse, error) {
	// 检查分类是否存在
	if req.CategoryID != nil && *req.CategoryID > 0 {
		_, err := s.CategoryRepository.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, err
		}
	}

	// 创建文章
	article := &schema.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		AuthorID:   userID,
		CategoryID: req.CategoryID,
		Cover:      req.Cover,
		Status:     req.Status,
	}

	if err := s.ArticleRepository.Create(ctx, article); err != nil {
		return nil, err
	}

	// 添加标签
	if len(req.TagIDs) > 0 {
		s.Trans.Exec(ctx, func(ctx context.Context) error {
			return s.addArticleTags(ctx, article.ID, req.TagIDs)
		})
	}

	// 获取文章详情
	return s.GetArticleByID(ctx, article.ID, userID)
}

func (s *ArticleService) addArticleTags(ctx context.Context, articleID uint, tagIDs []uint) error {
	// 删除现有标签
	err := s.ArticleTagRepository.DeleteByArticleID(ctx, articleID)
	if err != nil {
		return err
	}

	// 添加新标签
	for _, tagID := range tagIDs {
		articleTag := &schema.ArticleTag{
			ArticleID: articleID,
			TagID:     tagID,
		}
		if err := s.ArticleTagRepository.Create(ctx, articleTag); err != nil {
			return err
		}
	}

	return nil
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(ctx context.Context, userID uint, id uint, req *schema.UpdateArticleRequest) (*schema.ArticleResponse, error) {
	// 获取文章
	article, err := s.ArticleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if article.AuthorID != userID {
		return nil, errors.Forbidden("无权限修改此文章")
	}

	// 检查分类是否存在
	if req.CategoryID != nil && *req.CategoryID > 0 {
		_, err := s.CategoryRepository.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, err
		}
	}

	// 更新文章字段
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Cover != "" {
		article.Cover = req.Cover
	}
	if req.Status != "" {
		article.Status = req.Status
	}

	// 保存更新
	if err := s.ArticleRepository.Update(ctx, article); err != nil {
		return nil, err
	}

	// 更新标签
	if len(req.TagIDs) > 0 {
		s.Trans.Exec(ctx, func(ctx context.Context) error {
			return s.addArticleTags(ctx, article.ID, req.TagIDs)
		})
	}

	// 获取文章详情
	return s.GetArticleByID(ctx, article.ID, userID)
}

// DeleteArticle 删除文章
func (s *ArticleService) DeleteArticle(ctx context.Context, userID uint, id uint) error {
	// 获取文章
	article, err := s.ArticleRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if article.AuthorID != userID {
		return errors.Forbidden("无权限删除此文章")
	}

	// 删除文章
	err = s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 删除文章标签关联
		if err := s.ArticleTagRepository.DeleteByArticleID(ctx, id); err != nil {
			return err
		}
		// 删除文章
		return s.ArticleRepository.Delete(ctx, id)
	})

	return err
}

func (s *ArticleService) UploadCover(c *gin.Context) (*schema.CoverResponse, error) {
	ctx := c.Request.Context()

	// 获取封面文件
	file, err := c.FormFile("cover")
	if err != nil {
		return nil, errors.BadRequest("获取封面文件失败: %s", err.Error())
	}

	// 验证文件类型
	if !util.IsImageFile(file.Filename) {
		return nil, errors.BadRequest("支持的文件格式为: %s", config.SupportedImageFormats)
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// 设置保存路径
	coverURL := filepath.Join("/pic/covers", newFilename)
	dst := filepath.Join(config.C.Middleware.Static.Dir, coverURL)

	// 保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		logging.Context(ctx).Error("保存图片文件失败", zap.String("filePath", dst), zap.Int64("fileSize", file.Size), zap.Error(err))
		return nil, errors.InternalServerError("保存图片文件失败: %s", err.Error())
	}
	logging.Context(ctx).Info("保存封面图片文件成功", zap.String("filePath", dst), zap.Int64("fileSize", file.Size))

	return &schema.CoverResponse{
		URL: coverURL,
	}, nil
}

// GetArticleByID 通过ID获取文章
func (s *ArticleService) GetArticleByID(ctx context.Context, articleID uint, userID uint) (*schema.ArticleResponse, error) {
	// 获取文章
	article, err := s.ArticleRepository.GetByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.ArticleResponse{
		ID:            article.ID,
		Title:         article.Title,
		Content:       article.Content,
		Summary:       article.Summary,
		AuthorID:      article.AuthorID,
		CategoryID:    article.CategoryID,
		Cover:         article.Cover,
		Status:        article.Status,
		ViewCount:     article.ViewCount,
		LikeCount:     article.LikeCount,
		CommentCount:  article.CommentCount,
		FavoriteCount: article.FavoriteCount,
		CreatedAt:     article.CreatedAt,
		UpdatedAt:     article.UpdatedAt,
	}

	// 获取标签
	s.FillTags(ctx, response)

	// 获取作者信息
	s.FillAuthor(ctx, response)

	// 如果用户已登录，获取用户与文章的交互状态
	if userID > 0 {
		interactions, err := s.GetArticleInteractions(ctx, userID, articleID)
		if err != nil {
			logging.Context(ctx).Error("获取用户与文章的交互状态失败", zap.Uint("article_id", articleID), zap.Uint("user_id", userID), zap.Error(err))
		} else {
			response.Interactions = interactions
		}
	}

	return response, nil
}

// GetArticleList 获取文章列表
func (s *ArticleService) GetArticleList(ctx context.Context, params *schema.ArticleQueryParams) (*schema.ArticlePaginationResult, error) {
	// 获取文章列表
	result, err := s.ArticleRepository.GetList(ctx, params)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result.Items {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// 获取作者信息
func (s *ArticleService) FillAuthor(ctx context.Context, item interface{}) {
	// 使用类型断言获取文章 ID 与作者 ID
	var articleID, userID uint
	var authorSetter func(string, string)

	// 根据不同类型获取 ID 并设置回调函数
	switch v := item.(type) {
	case *schema.ArticleListItem:
		articleID, userID = v.ID, v.AuthorID
		authorSetter = func(username, avatar string) {
			v.Author = username
			v.AuthorAvatar = avatar
		}
	case *schema.ArticleResponse:
		articleID, userID = v.ID, v.AuthorID
		authorSetter = func(username, avatar string) {
			v.Author = username
			v.AuthorAvatar = avatar
		}
	default:
		logging.Context(ctx).Error("不支持的文章类型", zap.String("type", fmt.Sprintf("%T", item)))
		return
	}

	// 获取文章作者信息
	if userID > 0 {
		user, err := s.UserRepository.GetByID(ctx, userID)
		if err != nil {
			logging.Context(ctx).Error("获取文章作者信息失败", zap.Uint("article_id", articleID), zap.Uint("author_id", userID), zap.Error(err))
		} else {
			authorSetter(user.Username, user.Avatar)
		}
	}
}

// 获取标签
func (s *ArticleService) FillTags(ctx context.Context, item interface{}) {
	// 使用类型断言获取文章ID
	var articleID uint
	var tagSetter func([]uint)

	// 根据不同类型获取ID并设置回调函数
	switch v := item.(type) {
	case *schema.ArticleListItem:
		articleID = v.ID
		tagSetter = func(tagIDs []uint) {
			v.Tags = tagIDs
		}
	case *schema.ArticleResponse:
		articleID = v.ID
		tagSetter = func(tagIDs []uint) {
			v.Tags = tagIDs
		}
	default:
		logging.Context(ctx).Error("不支持的文章类型", zap.String("type", fmt.Sprintf("%T", item)))
		return
	}

	// 获取文章标签
	tags, err := s.ArticleRepository.GetArticleTags(ctx, articleID)
	if err != nil {
		logging.Context(ctx).Error("获取文章对应标签失败", zap.Uint("article_id", articleID), zap.Error(err))
	} else if len(tags) > 0 {
		// 提取标签名称
		tagIDs := make([]uint, 0, len(tags))
		for _, tag := range tags {
			tagIDs = append(tagIDs, tag.ID)
		}
		// 使用回调函数设置标签
		tagSetter(tagIDs)
	}
}

// ViewArticle 浏览文章
func (s *ArticleService) ViewArticle(ctx context.Context, userID, articleID uint) error {
	// 增加浏览次数
	if err := s.ArticleRepository.IncrementViewCount(ctx, articleID); err != nil {
		return err
	}

	// 记录用户交互
	if userID > 0 {
		return s.InteractionRepository.RecordView(ctx, userID, articleID)
	}

	return nil
}

// LikeArticle 点赞/取消点赞文章
func (s *ArticleService) LikeArticle(ctx context.Context, userID, articleID uint) (*schema.ArticleInteractionResponse, error) {
	// 获取用户当前交互状态
	_, err := s.InteractionRepository.Get(ctx, userID, articleID, "like")
	if err == nil { // 已点赞，则取消点赞
		// 删除交互记录
		if err := s.InteractionRepository.Delete(ctx, userID, articleID, "like"); err != nil {
			return nil, err
		}
		// 减少点赞数
		if err := s.ArticleRepository.IncrementLikeCount(ctx, articleID, -1); err != nil {
			return nil, err
		}
	} else if errors.IsNotFound(err) { // 未点赞，则添加点赞
		// 添加点赞记录
		newInteraction := &schema.UserInteraction{
			UserID:    userID,
			ArticleID: articleID,
			Type:      "like",
		}
		if err := s.InteractionRepository.CreateOrUpdate(ctx, newInteraction); err != nil {
			return nil, err
		}
		// 增加点赞数
		if err := s.ArticleRepository.IncrementLikeCount(ctx, articleID, 1); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 获取最新的交互状态
	interaction, err := s.InteractionRepository.GetUserInteractions(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	return &schema.ArticleInteractionResponse{
		Interacted: interaction.Liked,
	}, nil
}

// FavoriteArticle 收藏/取消收藏文章
func (s *ArticleService) FavoriteArticle(ctx context.Context, userID, articleID uint) (*schema.ArticleInteractionResponse, error) {
	// 获取用户当前交互状态
	_, err := s.InteractionRepository.Get(ctx, userID, articleID, "favorite")
	if err == nil { // 已收藏，则取消收藏
		// 删除交互记录
		if err := s.InteractionRepository.Delete(ctx, userID, articleID, "favorite"); err != nil {
			return nil, err
		}
		// 减少收藏数
		if err := s.ArticleRepository.IncrementFavoriteCount(ctx, articleID, -1); err != nil {
			return nil, err
		}
	} else if errors.IsNotFound(err) { // 未收藏，则添加收藏
		// 添加收藏记录
		newInteraction := &schema.UserInteraction{
			UserID:    userID,
			ArticleID: articleID,
			Type:      "favorite",
		}
		if err := s.InteractionRepository.CreateOrUpdate(ctx, newInteraction); err != nil {
			return nil, err
		}
		// 增加收藏数
		if err := s.ArticleRepository.IncrementFavoriteCount(ctx, articleID, 1); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 获取最新的交互状态
	interaction, err := s.InteractionRepository.GetUserInteractions(ctx, userID, articleID)
	if err != nil {
		return nil, err
	}

	return &schema.ArticleInteractionResponse{
		Interacted: interaction.Favorited,
	}, nil
}

// GetUserLikedArticles 获取用户点赞的文章
func (s *ArticleService) GetUserLikedArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	logging.Context(ctx).Debug("获取用户点赞的文章", zap.Uint("user_id", userID), zap.Int("page", page), zap.Int("page_size", pageSize))
	result, err := s.ArticleRepository.GetUserLikedArticles(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result.Items {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetUserFavoriteArticles 获取用户收藏的文章
func (s *ArticleService) GetUserFavoriteArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	result, err := s.ArticleRepository.GetUserFavoriteArticles(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result.Items {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetUserViewHistory 获取用户浏览历史
func (s *ArticleService) GetUserViewHistory(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	result, err := s.InteractionRepository.GetUserViewHistory(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result.Items {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetUserCommentedArticles 获取用户评论过的文章
func (s *ArticleService) GetUserCommentedArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	result, err := s.ArticleRepository.GetUserCommentedArticles(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result.Items {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetHotArticles 获取热门文章
func (s *ArticleService) GetHotArticles(ctx context.Context, limit int) ([]*schema.ArticleListItem, error) {
	result, err := s.ArticleRepository.GetHotArticles(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetLatestArticles 获取最新文章
func (s *ArticleService) GetLatestArticles(ctx context.Context, limit int) ([]*schema.ArticleListItem, error) {
	result, err := s.ArticleRepository.GetLatestArticles(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 补充文章信息
	for _, item := range result {
		s.FillAuthor(ctx, item)
		s.FillTags(ctx, item)
	}

	return result, nil
}

// GetArticleInteractions 获取用户与文章的交互状态
func (s *ArticleService) GetArticleInteractions(ctx context.Context, userID, articleID uint) (*schema.InteractionResponse, error) {
	return s.InteractionRepository.GetUserInteractions(ctx, userID, articleID)
}
