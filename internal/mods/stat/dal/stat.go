package dal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	userDal "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	userSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	blogDal "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	blogSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	commentDal "github.com/codeExpert666/goinkblog-backend/internal/mods/comment/dal"
	commentSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetLogDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB)
}

// StatRepository 统计数据访问层
type StatRepository struct {
	DB *gorm.DB
}

// GetUserArticleVisitTrend 获取用户文章访问趋势数据
func (r *StatRepository) GetUserArticleVisitTrend(ctx context.Context, userID uint, days int) (*schema.UserArticleVisitTrendResponse, error) {
	var result schema.UserArticleVisitTrendResponse
	var articles []blogSchema.Article
	var totalVisit int64

	// 默认查询7天
	if days <= 0 {
		days = 7
	}

	// 1. 获取用户的所有文章ID
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).
		Where("author_id = ?", userID).
		Find(&articles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 如果用户没有文章，返回空结果
	if len(articles) == 0 {
		return &schema.UserArticleVisitTrendResponse{
			Items:      []schema.ArticleVisitTrendItem{},
			TotalVisit: 0,
		}, nil
	}

	// 提取文章ID
	var articleIDs []uint
	for _, article := range articles {
		articleIDs = append(articleIDs, article.ID)
		totalVisit += int64(article.ViewCount)
	}

	// 2. 生成日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 3. 从日志表中查询访问记录
	// 构建文章访问路径模式
	pathPatterns := make([]string, 0, len(articleIDs))
	for _, id := range articleIDs {
		pathPatterns = append(pathPatterns, fmt.Sprintf("data->>'$.path' = '/api/blog/articles/%d'", id))
	}

	// 查询条件：是请求日志、成功的GET请求、访问了用户的文章
	whereCondition := fmt.Sprintf(
		"tag = 'request' AND created_at >= ? AND data->>'$.method' = 'GET' AND data->>'$.status' + 0 >= 200 AND data->>'$.status' + 0 < 300 AND (%s)",
		strings.Join(pathPatterns, " OR "),
	)

	rows, err := GetLogDB(ctx, r.DB).Model(&schema.Logger{}).
		Select("DATE(created_at) as date", "COUNT(*) as count").
		Where(whereCondition, startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	// 构建日期到计数的映射
	dateCountMap := make(map[string]*schema.ArticleVisitTrendItem)
	for rows.Next() {
		var date string
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			return nil, errors.WithStack(err)
		}

		// 剔除日期中的时间时区部分
		date = strings.Split(date, "T")[0]
		logging.Context(ctx).Warn("用户文章访问量", zap.String("date", date), zap.Int64("count", count))

		dateCountMap[date] = &schema.ArticleVisitTrendItem{
			Date:       date,
			VisitCount: count,
		}
	}

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	// 生成连续的日期数据
	currentDate := startDate
	var items []schema.ArticleVisitTrendItem
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		item, exists := dateCountMap[dateStr]
		if !exists {
			item = &schema.ArticleVisitTrendItem{
				Date:       dateStr,
				VisitCount: 0,
			}
		}
		items = append(items, *item)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	result.Items = items
	result.TotalVisit = totalVisit

	return &result, nil
}

// GetUserArticleStatistic 获取用户文章统计信息
func (r *StatRepository) GetUserArticleStatistic(ctx context.Context, userID uint) *schema.SiteOverviewResponse {
	var result schema.SiteOverviewResponse

	// 获取用户的文章总数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Where("author_id = ?", userID).Count(&result.TotalArticles).Error; err != nil {
		logging.Context(ctx).Error("获取用户的文章总数失败", zap.Uint("user_id", userID), zap.Error(errors.WithStack(err)))
	}

	// 获取用户的文章总浏览次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Where("author_id = ?", userID).Select("COALESCE(SUM(view_count), 0)").Row().Scan(&result.TotalViews); err != nil {
		logging.Context(ctx).Error("获取用户的文章总浏览次数失败", zap.Uint("user_id", userID), zap.Error(errors.WithStack(err)))
	}

	// 获取用户的文章总点赞次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Where("author_id = ?", userID).Select("COALESCE(SUM(like_count), 0)").Row().Scan(&result.TotalLikes); err != nil {
		logging.Context(ctx).Error("获取用户的文章总点赞次数失败", zap.Uint("user_id", userID), zap.Error(errors.WithStack(err)))
	}

	// 获取用户的文章总评论次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Where("author_id = ?", userID).Select("COALESCE(SUM(comment_count), 0)").Row().Scan(&result.TotalComments); err != nil {
		logging.Context(ctx).Error("获取用户的文章总评论次数失败", zap.Uint("user_id", userID), zap.Error(errors.WithStack(err)))
	}

	// 获取用户的文章总收藏次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Where("author_id = ?", userID).Select("COALESCE(SUM(favorite_count), 0)").Row().Scan(&result.TotalFavorites); err != nil {
		logging.Context(ctx).Error("获取用户的文章总收藏次数失败", zap.Uint("user_id", userID), zap.Error(errors.WithStack(err)))
	}

	return &result
}

// GetOverview 获取站点概览统计信息
func (r *StatRepository) GetOverview(ctx context.Context) *schema.SiteOverviewResponse {
	var result schema.SiteOverviewResponse

	// 获取用户总数
	if err := userDal.GetUserDB(ctx, r.DB).Count(&result.TotalUsers).Error; err != nil {
		logging.Context(ctx).Error("获取用户数量失败", zap.Error(errors.WithStack(err)))
	}

	// 获取文章总数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Count(&result.TotalArticles).Error; err != nil {
		logging.Context(ctx).Error("获取文章总数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总浏览次数
	// COALESCE 函数是一个 SQL 函数，它接受多个参数，返回第一个非 NULL 的参数
	// 如果没有文章记录或所有 view_count 都为 NULL，SUM(view_count) 会返回 NULL
	// COALESCE(SUM(view_count), 0) 确保即使没有数据，也返回 0 而不是 NULL
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Select("COALESCE(SUM(view_count), 0)").Row().Scan(&result.TotalViews); err != nil {
		logging.Context(ctx).Error("获取文章总浏览次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总点赞次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Select("COALESCE(SUM(like_count), 0)").Row().Scan(&result.TotalLikes); err != nil {
		logging.Context(ctx).Error("获取文章总点赞次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总评论次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Select("COALESCE(SUM(comment_count), 0)").Row().Scan(&result.TotalComments); err != nil {
		logging.Context(ctx).Error("获取文章总评论次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总收藏次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Model(&blogSchema.Article{}).Select("COALESCE(SUM(favorite_count), 0)").Row().Scan(&result.TotalFavorites); err != nil {
		logging.Context(ctx).Error("获取文章总收藏次数失败", zap.Error(errors.WithStack(err)))
	}

	return &result
}

// GetVisitTrend 获取访问趋势数据
func (r *StatRepository) GetVisitTrend(ctx context.Context, days int) ([]schema.APIAccessTrendItem, error) {
	var result []schema.APIAccessTrendItem

	// 生成日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	logging.Context(ctx).Warn("查看日期范围", zap.Time("start_date", startDate), zap.Time("end_date", endDate))

	// 获取每天的访问次数
	totalCountSQL := "COUNT(*) as total_count"
	successCountSQL := "SUM(CASE WHEN data->>'$.status' + 0 BETWEEN 200 AND 299 THEN 1 ELSE 0 END) as success_count"
	clientErrorCountSQL := "SUM(CASE WHEN data->>'$.status' + 0 BETWEEN 400 AND 499 THEN 1 ELSE 0 END) as client_error_count"
	serverErrorCountSQL := "SUM(CASE WHEN data->>'$.status' + 0 BETWEEN 500 AND 599 THEN 1 ELSE 0 END) as server_error_count"
	rows, err := GetLogDB(ctx, r.DB).Model(&schema.Logger{}).
		Select("DATE(created_at) as date", totalCountSQL, successCountSQL, clientErrorCountSQL, serverErrorCountSQL).
		Where("tag = 'request' AND created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	// 构建日期到计数的映射
	dateCountMap := make(map[string]*schema.APIAccessTrendItem)
	for rows.Next() {
		var date string
		var totalCount, successCount, clientErrorCount, serverErrorCount int64
		if err := rows.Scan(&date, &totalCount, &successCount, &clientErrorCount, &serverErrorCount); err != nil {
			return nil, errors.WithStack(err)
		}

		// 剔除日期信息中的时间时区信息
		date = strings.Split(date, "T")[0]
		logging.Context(ctx).Warn("解析到的数据", zap.String("date", date), zap.Int64("total_count", totalCount), zap.Int64("success_count", successCount), zap.Int64("client_error_count", clientErrorCount), zap.Int64("server_error_count", serverErrorCount))

		dateCountMap[date] = &schema.APIAccessTrendItem{
			Date:             date,
			TotalCount:       totalCount,
			SuccessCount:     successCount,
			ClientErrorCount: clientErrorCount,
			ServerErrorCount: serverErrorCount,
		}
	}

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	// 生成连续的日期数据
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		item, exists := dateCountMap[dateStr]
		if !exists {
			item = &schema.APIAccessTrendItem{
				Date: dateStr,
			}
		}
		result = append(result, *item)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// GetUserActivityTrend 获取用户活跃度数据
func (r *StatRepository) GetUserActivityTrend(ctx context.Context, days int) ([]schema.UserActivityTrendItem, error) {
	var result []schema.UserActivityTrendItem

	// 生成日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 获取每天的用户活跃数
	rows, err := GetLogDB(ctx, r.DB).Model(&schema.Logger{}).
		Select("DATE(created_at) as date", "COUNT(DISTINCT user_id) as count").
		Where("tag = 'request' AND created_at >= ? AND user_id > 0", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	// 构建日期到计数的映射
	dateCountMap := make(map[string]*schema.UserActivityTrendItem)
	for rows.Next() {
		var date string
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			return nil, errors.WithStack(err)
		}

		// 剔除日期信息中的时间时区信息
		date = strings.Split(date, "T")[0]

		dateCountMap[date] = &schema.UserActivityTrendItem{
			Date:      date,
			UserCount: count,
		}
	}

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	// 生成连续的日期数据
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		item, exists := dateCountMap[dateStr]
		if !exists {
			item = &schema.UserActivityTrendItem{
				Date: dateStr,
			}
		}
		result = append(result, *item)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// GetUserCategoryDistribution 获取用户文章分类分布
func (r *StatRepository) GetUserCategoryDistribution(ctx context.Context, userID uint) ([]schema.CategoryDistItem, error) {
	var result []schema.CategoryDistItem

	categoryTableName := new(blogSchema.Category).TableName()
	articleTableName := new(blogSchema.Article).TableName()
	err := blogDal.GetCategoryDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.name, COUNT(a.id) as count", categoryTableName)).
		Joins(fmt.Sprintf("LEFT JOIN %s a ON %s.id = a.category_id", articleTableName, categoryTableName), userID).
		Where("a.author_id = ?", userID).
		Group(fmt.Sprintf("%s.id, %s.name", categoryTableName, categoryTableName)).
		Order("count DESC").
		Scan(&result).Error

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// GetCategoryDistribution 获取文章分类分布
func (r *StatRepository) GetCategoryDistribution(ctx context.Context) ([]schema.CategoryDistItem, error) {
	var result []schema.CategoryDistItem

	categoryTableName := new(blogSchema.Category).TableName()
	articleTableName := new(blogSchema.Article).TableName()
	err := blogDal.GetCategoryDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.id, %s.name, COUNT(a.id) as count", categoryTableName, categoryTableName)).
		Joins(fmt.Sprintf("LEFT JOIN %s a ON %s.id = a.category_id", articleTableName, categoryTableName)).
		Where("a.status LIKE 'published'").
		Group(fmt.Sprintf("%s.id, %s.name", categoryTableName, categoryTableName)).
		Order("count DESC").
		Scan(&result).Error

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// GetLogList 获取日志列表
func (r *StatRepository) GetLogList(ctx context.Context, params *schema.LoggerQueryParams) (*schema.LoggerPaginationResult, error) {
	var result schema.LoggerPaginationResult

	// 默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 关联用户表
	logTableName := new(schema.Logger).TableName()
	userTableName := new(userSchema.User).TableName()
	db := GetLogDB(ctx, r.DB).Table(fmt.Sprintf("%s AS l", logTableName)).
		Select("l.*", "u.username").
		Joins(fmt.Sprintf("LEFT JOIN %s AS u ON l.user_id = u.id", userTableName))

	// 应用过滤条件
	if params.Level != "" {
		db = db.Where("l.level = ?", params.Level)
	}
	if params.TraceID != "" {
		db = db.Where("l.trace_id = ?", params.TraceID)
	}
	if params.LikeUsername != "" {
		db = db.Where("u.username LIKE ?", "%"+params.LikeUsername+"%")
	}
	if params.Tag != "" {
		db = db.Where("l.tag = ?", params.Tag)
	}
	if params.LikeMessage != "" {
		db = db.Where("l.message LIKE ?", "%"+params.LikeMessage+"%")
	}
	if params.StartTime != "" && params.EndTime != "" {
		db = db.Where("l.created_at BETWEEN ? AND ?", params.StartTime, params.EndTime)
	}

	// 计算总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 获取分页数据
	offset := (params.Page - 1) * params.PageSize
	db = db.Offset(offset).Limit(params.PageSize).Order("l.created_at DESC")

	var logs []schema.Logger
	if err := db.Find(&logs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	result.Items = logs
	result.Total = total
	result.Page = params.Page
	result.PageSize = params.PageSize
	result.TotalPages = int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &result, nil
}

// GetArticleCreationTimeStats 获取文章创作时间统计数据
func (r *StatRepository) GetArticleCreationTimeStats(ctx context.Context, days int) ([]schema.ArticleCreationTimeStatsItem, error) {
	// 生成日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 查询文章表（关联分类表）
	articleTableName := new(blogSchema.Article).TableName()
	db := blogDal.GetArticleDB(ctx, r.DB).Table(fmt.Sprintf("%s AS a", articleTableName))

	categoryTableName := new(blogSchema.Category).TableName()
	rows, err := db.Select("DATE(a.created_at) as date", "c.name as category", "COUNT(a.id) as count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS c ON a.category_id = c.id", categoryTableName)).
		Where("a.created_at >= ?", startDate).
		Group("DATE(a.created_at), c.name").
		Order("date ASC").
		Rows()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer rows.Close()

	// 构建日期到统计数据的映射
	dateMap := make(map[string]*schema.ArticleCreationTimeStatsItem)

	// 遍历查询结果
	for rows.Next() {
		var date string
		var count int64
		var category sql.NullString // 应对文章没有类型名的情况

		// 扫描数据
		if err := rows.Scan(&date, &category, &count); err != nil {
			return nil, errors.WithStack(err)
		}

		// 剔除日期中的时间时区部分
		date = strings.Split(date, "T")[0]

		// 获取或创建当前日期的数据项
		item, exists := dateMap[date]
		if !exists {
			item = &schema.ArticleCreationTimeStatsItem{
				Date:       date,
				Count:      0,
				Categories: make(map[string]int64),
			}
			dateMap[date] = item
		}

		// 更新该日期的总数量
		item.Count += count

		// 更新该日期的分类统计
		categoryName := "未分类" // 为 NULL 值设置默认分类名
		if category.Valid {
			categoryName = category.String
		}
		item.Categories[categoryName] += count
	}

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	// 生成连续的日期数据
	currentDate := startDate
	var result []schema.ArticleCreationTimeStatsItem

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		item, exists := dateMap[dateStr]
		if !exists {
			// 如果该日期没有数据，创建一个空数据项
			item = &schema.ArticleCreationTimeStatsItem{
				Date:       dateStr,
				Count:      0,
				Categories: make(map[string]int64),
			}
		}
		result = append(result, *item)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// GetCommentStatistic 获取所有评论统计数据
func (r *StatRepository) GetCommentStatistic(ctx context.Context) *schema.CommentStatisticResponse {
	var result schema.CommentStatisticResponse

	// 获取评论总数
	if err := commentDal.GetCommentDB(ctx, r.DB).Count(&result.TotalComments).Error; err != nil {
		logging.Context(ctx).Error("获取评论总数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取通过的评论数量
	if err := commentDal.GetCommentDB(ctx, r.DB).Where("status = ?", commentSchema.CommentStatusApproved).Count(&result.PassedComments).Error; err != nil {
		logging.Context(ctx).Error("获取通过的评论数量失败", zap.Error(errors.WithStack(err)))
	}

	// 获取待审核的评论数量
	if err := commentDal.GetCommentDB(ctx, r.DB).Where("status = ?", commentSchema.CommentStatusPending).Count(&result.PendingComments).Error; err != nil {
		logging.Context(ctx).Error("获取待审核的评论数量失败", zap.Error(errors.WithStack(err)))
	}

	// 获取拒绝的评论数量
	if err := commentDal.GetCommentDB(ctx, r.DB).Where("status = ?", commentSchema.CommentStatusRejected).Count(&result.RejectedComments).Error; err != nil {
		logging.Context(ctx).Error("获取拒绝的评论数量失败", zap.Error(errors.WithStack(err)))
	}

	return &result
}
