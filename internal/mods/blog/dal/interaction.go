package dal

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetInteractionDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.UserInteraction{})
}

// InteractionRepository 用户交互数据访问层
type InteractionRepository struct {
	DB *gorm.DB
}

// TODO: 检查更新操作的是否正确（Omit 没有忽略 ID）
// CreateOrUpdate 创建或更新用户交互
func (r *InteractionRepository) CreateOrUpdate(ctx context.Context, interaction *schema.UserInteraction) error {
	// 先查询是否存在
	var count int64
	err := GetInteractionDB(ctx, r.DB).
		Where("user_id = ? AND article_id = ? AND type = ?",
			interaction.UserID, interaction.ArticleID, interaction.Type).
		Count(&count).Error

	if err != nil {
		return errors.WithStack(err)
	}

	// 如果不存在则创建
	if count == 0 {
		result := GetInteractionDB(ctx, r.DB).Create(interaction)
		return errors.WithStack(result.Error)
	}

	// 如果存在则更新
	result := GetInteractionDB(ctx, r.DB).
		Where("user_id = ? AND article_id = ? AND type = ?",
			interaction.UserID, interaction.ArticleID, interaction.Type).
		Select("*").Omit("created_at").
		Updates(interaction)
	return errors.WithStack(result.Error)
}

// Delete 删除用户交互
func (r *InteractionRepository) Delete(ctx context.Context, userID, articleID uint, interactionType string) error {
	result := GetInteractionDB(ctx, r.DB).
		Where("user_id = ? AND article_id = ? AND type = ?",
			userID, articleID, interactionType).
		Delete(&schema.UserInteraction{})
	return errors.WithStack(result.Error)
}

// Get 获取用户交互
func (r *InteractionRepository) Get(ctx context.Context, userID, articleID uint, interactionType string) (*schema.UserInteraction, error) {
	var interaction schema.UserInteraction
	err := GetInteractionDB(ctx, r.DB).Where("user_id = ? AND article_id = ? AND type = ?",
		userID, articleID, interactionType).
		First(&interaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户 %d 与文章 %d 的 %s 交互不存在", userID, articleID, interactionType)
		}
		return nil, errors.WithStack(err)
	}

	return &interaction, nil
}

// GetUserInteractions 获取用户与文章的所有交互
func (r *InteractionRepository) GetUserInteractions(ctx context.Context, userID, articleID uint) (*schema.InteractionResponse, error) {
	var response schema.InteractionResponse

	// 查询是否点赞
	var likeCount int64
	err := GetInteractionDB(ctx, r.DB).
		Where("user_id = ? AND article_id = ? AND type = ?", userID, articleID, "like").
		Count(&likeCount).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 查询是否收藏
	var favoriteCount int64
	err = GetInteractionDB(ctx, r.DB).
		Where("user_id = ? AND article_id = ? AND type = ?", userID, articleID, "favorite").
		Count(&favoriteCount).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response.Liked = likeCount > 0
	response.Favorited = favoriteCount > 0

	return &response, nil
}

// RecordView 记录文章浏览
func (r *InteractionRepository) RecordView(ctx context.Context, userID, articleID uint) error {
	interaction := schema.UserInteraction{
		UserID:    userID,
		ArticleID: articleID,
		Type:      "view",
	}

	return r.CreateOrUpdate(ctx, &interaction)
}

// GetUserViewHistory 获取用户浏览历史
func (r *InteractionRepository) GetUserViewHistory(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	var result schema.ArticlePaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	articleName := new(schema.Article).TableName()
	interactionName := new(schema.UserInteraction).TableName()
	db := GetArticleDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.*, u.created_at as interaction_time", articleName)).
		Joins(fmt.Sprintf("JOIN %s u ON %s.id = u.article_id", interactionName, articleName)).
		Where("u.user_id = ? AND u.type = ?", userID, "view")

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var articles []struct {
		schema.Article
		InteractionTime time.Time `json:"interaction_time"`
	}

	if err := db.Offset(offset).Limit(pageSize).
		Order("u.created_at DESC").
		Scan(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.ArticleListItem
	for _, article := range articles {
		item := schema.ArticleListItem{
			ID:            article.ID,
			Title:         article.Title,
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
			InteractionTime: article.InteractionTime,
		}
		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}
