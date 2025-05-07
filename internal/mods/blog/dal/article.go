package dal

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	userSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	commentSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetArticleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB)
}

// ArticleRepository 文章数据访问层
type ArticleRepository struct {
	DB *gorm.DB
}

// Create 创建文章
func (r *ArticleRepository) Create(ctx context.Context, article *schema.Article) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Create(article)
	return errors.WithStack(result.Error)
}

// Update 更新文章
func (r *ArticleRepository) Update(ctx context.Context, article *schema.Article) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", article.ID).Select("*").Omit("created_at").Updates(article)
	return errors.WithStack(result.Error)
}

// Delete 删除文章
func (r *ArticleRepository) Delete(ctx context.Context, id uint) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).Delete(&schema.Article{})
	return errors.WithStack(result.Error)
}

// GetByID 通过ID获取文章
func (r *ArticleRepository) GetByID(ctx context.Context, id uint) (*schema.Article, error) {
	var article schema.Article
	err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).First(&article).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("文章不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &article, nil
}

// GetList 获取文章列表
func (r *ArticleRepository) GetList(ctx context.Context, params *schema.ArticleQueryParams) (*schema.ArticlePaginationResult, error) {
	var result schema.ArticlePaginationResult

	// 默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	articleTableName := new(schema.Article).TableName()
	db := GetArticleDB(ctx, r.DB).Table(fmt.Sprintf("%s AS a", articleTableName))

	// 应用过滤条件
	if len(params.TagIDs) > 0 {
		articleTagTableName := new(schema.ArticleTag).TableName()
		db = db.Joins(fmt.Sprintf("JOIN %s AS t ON a.id = t.article_id", articleTagTableName)).
			Where("t.tag_id IN ?", params.TagIDs)
		// 如果有多个标签，需要群组查询以确保文章包含所有指定的标签
		if len(params.TagIDs) > 1 {
			db = db.Group("a.id").
				Having("COUNT(DISTINCT CASE WHEN t.tag_id IN (?) THEN t.tag_id ELSE NULL END) = ?", params.TagIDs, len(params.TagIDs))
		}
	}
	if params.Author != "" {
		if params.Author == "current" {
			db = db.Where("a.author_id = ?", util.FromUserID(ctx))
		} else {
			userTableName := new(userSchema.User).TableName()
			db = db.Joins(fmt.Sprintf("JOIN %s AS u ON a.author_id = u.id", userTableName)).
				Where("u.username = ?", params.Author)
		}
	}
	if len(params.CategoryIDs) > 0 {
		db = db.Where("a.category_id IN ?", params.CategoryIDs)
	}
	if params.Status != "" {
		db = db.Where("a.status = ?", params.Status)
	}
	if params.Keyword != "" {
		db = db.Where("a.title LIKE ? OR a.summary LIKE ?", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}
	logging.Context(ctx).Debug("params.TimeRange", zap.String("params.TimeRange", params.TimeRange))
	if params.TimeRange != "" {
		switch params.TimeRange {
		case "today":
			// 今天创建的文章
			today := time.Now().Format("2006-01-02")
			logging.Context(ctx).Debug("today", zap.String("today", today))
			db = db.Where("a.created_at >= ?", today+" 00:00:00")
		case "week":
			// 本周创建的文章（从本周一开始）
			weekday := int(time.Now().Weekday())
			if weekday == 0 {
				weekday = 7 // 周日是0，我们把它当作7
			}
			weekStart := time.Now().AddDate(0, 0, -(weekday - 1)).Format("2006-01-02")
			db = db.Where("a.created_at >= ?", weekStart+" 00:00:00")
		case "month":
			// 本月创建的文章（过去30天）
			monthAgo := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location()).Format("2006-01-02")
			db = db.Where("a.created_at >= ?", monthAgo+" 00:00:00")
		case "year":
			// 本年创建的文章
			yearStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Now().Location()).Format("2006-01-02")
			db = db.Where("a.created_at >= ?", yearStart+" 00:00:00")
		case "all":
			// 所有时间，不需要额外条件
		}
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用排序
	switch params.SortBy {
	case "views":
		db = db.Order("a.view_count DESC")
	case "likes":
		db = db.Order("a.like_count DESC")
	case "favorites":
		db = db.Order("a.favorite_count DESC")
	case "comments":
		db = db.Order("a.comment_count DESC")
	}
	db = db.Order("a.created_at DESC")

	// 分页
	offset := (params.Page - 1) * params.PageSize
	db = db.Offset(offset).Limit(params.PageSize)

	var articles []schema.Article
	if err := db.Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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
		}
		items = append(items, item)
	}

	result.Items = items
	result.Total = total
	result.Page = params.Page
	result.PageSize = params.PageSize
	// 不足一页的，也当一页
	result.TotalPages = int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &result, nil
}

// IncrementViewCount 增加文章浏览次数
func (r *ArticleRepository) IncrementViewCount(ctx context.Context, id uint) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	return errors.WithStack(result.Error)
}

// IncrementLikeCount 增加文章点赞次数
func (r *ArticleRepository) IncrementLikeCount(ctx context.Context, id uint, value int) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", value))
	return errors.WithStack(result.Error)
}

// IncrementFavoriteCount 增加文章收藏次数
func (r *ArticleRepository) IncrementFavoriteCount(ctx context.Context, id uint, value int) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", value))
	return errors.WithStack(result.Error)
}

// IncrementCommentCount 增加文章评论次数
func (r *ArticleRepository) IncrementCommentCount(ctx context.Context, id uint, value int) error {
	result := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", value))
	return errors.WithStack(result.Error)
}

// GetArticleTags 获取文章标签
func (r *ArticleRepository) GetArticleTags(ctx context.Context, articleID uint) ([]schema.Tag, error) {
	var tags []schema.Tag

	tagName := new(schema.Tag).TableName()
	articleTagName := new(schema.ArticleTag).TableName()
	err := GetTagDB(ctx, r.DB).
		Joins(fmt.Sprintf("JOIN %s a ON %s.id = a.tag_id", articleTagName, tagName)).
		Where("a.article_id = ?", articleID).
		Find(&tags).Error

	return tags, errors.WithStack(err)
}

// GetUserLikedArticles 获取用户点赞的文章
func (r *ArticleRepository) GetUserLikedArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	var result schema.ArticlePaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	articleName := new(schema.Article).TableName()
	db := GetArticleDB(ctx, r.DB).Table(fmt.Sprintf("%s AS a", articleName))

	userInteractionName := new(schema.UserInteraction).TableName()
	db = db.Joins(fmt.Sprintf("JOIN %s AS u ON a.id = u.article_id", userInteractionName)).
		Where("u.user_id = ? AND u.type = ?", userID, "like")

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (page - 1) * pageSize
	var articles []schema.Article
	if err := db.Offset(offset).Limit(pageSize).Order("u.created_at DESC").Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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

// GetUserFavoriteArticles 获取用户收藏的文章
func (r *ArticleRepository) GetUserFavoriteArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	var result schema.ArticlePaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	articleName := new(schema.Article).TableName()
	db := GetArticleDB(ctx, r.DB).Table(fmt.Sprintf("%s AS a", articleName))

	userInteractionName := new(schema.UserInteraction).TableName()
	db = db.Joins(fmt.Sprintf("JOIN %s AS u ON a.id = u.article_id", userInteractionName)).
		Where("u.user_id = ? AND u.type = ?", userID, "favorite")

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页
	offset := (page - 1) * pageSize
	var articles []schema.Article
	if err := db.Offset(offset).Limit(pageSize).Order("u.created_at DESC").Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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

// GetUserCommentedArticles 获取用户评论过的文章
func (r *ArticleRepository) GetUserCommentedArticles(ctx context.Context, userID uint, page, pageSize int) (*schema.ArticlePaginationResult, error) {
	var result schema.ArticlePaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	articleName := new(schema.Article).TableName()
	db := GetArticleDB(ctx, r.DB).Table(fmt.Sprintf("%s AS a", articleName))

	// 定义子查询SQL，获取每篇文章的最新评论时间
	commentName := new(commentSchema.Comment).TableName()
	subQuerySQL := fmt.Sprintf(
		"(SELECT article_id, MAX(created_at) as latest_comment_time FROM %s WHERE author_id = ? GROUP BY article_id)",
		commentName,
	)

	// 主查询
	db = db.Select("a.*").Joins(fmt.Sprintf("JOIN %s AS latest_comments ON a.id = latest_comments.article_id", subQuerySQL), userID)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 分页并按最新评论时间降序排序
	offset := (page - 1) * pageSize
	var articles []schema.Article
	if err := db.Offset(offset).Limit(pageSize).Order("latest_comments.latest_comment_time DESC").Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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

// GetHotArticles 获取热门文章
func (r *ArticleRepository) GetHotArticles(ctx context.Context, limit int) ([]*schema.ArticleListItem, error) {
	if limit <= 0 {
		limit = 5
	}

	var articles []schema.Article
	if err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).
		Where("status = ?", "published").
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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
		}
		items = append(items, item)
	}

	return items, nil
}

// GetLatestArticles 获取最新文章
func (r *ArticleRepository) GetLatestArticles(ctx context.Context, limit int) ([]*schema.ArticleListItem, error) {
	if limit <= 0 {
		limit = 5
	}

	var articles []schema.Article
	if err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).
		Where("status = ?", "published").
		Order("created_at DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var items []*schema.ArticleListItem
	for _, article := range articles {
		item := &schema.ArticleListItem{
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
		}
		items = append(items, item)
	}

	return items, nil
}
