package dal

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	userSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	blogDal "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	blogSchema "github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetLogDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.Logger{})
}

// StatRepository 统计数据访问层
type StatRepository struct {
	DB *gorm.DB
}

// GetArticleStatistic 获取文章统计信息
func (r *StatRepository) GetArticleStatistic(ctx context.Context) *schema.ArticleStatisticResponse {
	var result schema.ArticleStatisticResponse

	// 获取文章总数
	if err := blogDal.GetArticleDB(ctx, r.DB).Count(&result.TotalArticles).Error; err != nil {
		logging.Context(ctx).Error("获取文章总数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总浏览次数
	// COALESCE 函数是一个 SQL 函数，它接受多个参数，返回第一个非 NULL 的参数
	// 如果没有文章记录或所有 view_count 都为 NULL，SUM(view_count) 会返回 NULL
	// COALESCE(SUM(view_count), 0) 确保即使没有数据，也返回 0 而不是 NULL
	if err := blogDal.GetArticleDB(ctx, r.DB).Select("COALESCE(SUM(view_count), 0)").Row().Scan(&result.TotalViews); err != nil {
		logging.Context(ctx).Error("获取文章总浏览次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总点赞次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Select("COALESCE(SUM(like_count), 0)").Row().Scan(&result.TotalLikes); err != nil {
		logging.Context(ctx).Error("获取文章总点赞次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总评论次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Select("COALESCE(SUM(comment_count), 0)").Row().Scan(&result.TotalComments); err != nil {
		logging.Context(ctx).Error("获取文章总评论次数失败", zap.Error(errors.WithStack(err)))
	}

	// 获取总收藏次数
	if err := blogDal.GetArticleDB(ctx, r.DB).Select("COALESCE(SUM(favorite_count), 0)").Row().Scan(&result.TotalFavorites); err != nil {
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

	// 获取每天的访问次数
	// 注意：JSON_EXTRACT 函数只适用于 MySQL
	totalCountSQL := "COUNT(*) as total_count"
	successCountSQL := "SUM(CASE WHEN JSON_EXTRACT(data, '$.status') BETWEEN 200 AND 299 THEN 1 ELSE 0 END) as success_count"
	clientErrorCountSQL := "SUM(CASE WHEN JSON_EXTRACT(data, '$.status') BETWEEN 400 AND 499 THEN 1 ELSE 0 END) as client_error_count"
	serverErrorCountSQL := "SUM(CASE WHEN JSON_EXTRACT(data, '$.status') BETWEEN 500 AND 599 THEN 1 ELSE 0 END) as server_error_count"
	rows, err := GetLogDB(ctx, r.DB).
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
	rows, err := GetLogDB(ctx, r.DB).
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

// GetCategoryDistribution 获取文章分类分布
func (r *StatRepository) GetCategoryDistribution(ctx context.Context) ([]schema.CategoryDistItem, error) {
	var result []schema.CategoryDistItem

	categoryTableName := new(blogSchema.Category).TableName()
	articleTableName := new(blogSchema.Article).TableName()
	err := blogDal.GetCategoryDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.name, COUNT(a.id) as count", categoryTableName)).
		Joins(fmt.Sprintf("LEFT JOIN %s a ON %s.id = a.category_id", articleTableName, categoryTableName)).
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
	db := GetLogDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.*", logTableName), "u.username").
		Joins(fmt.Sprintf("LEFT JOIN %s u ON %s.user_id = u.id", userTableName, logTableName))

	// 应用过滤条件
	if params.Level != "" {
		db = db.Where(fmt.Sprintf("%s.level = ?", logTableName), params.Level)
	}
	if params.TraceID != "" {
		db = db.Where(fmt.Sprintf("%s.trace_id = ?", logTableName), params.TraceID)
	}
	if params.LikeUsername != "" {
		db = db.Where("u.username LIKE ?", "%"+params.LikeUsername+"%")
	}
	if params.Tag != "" {
		db = db.Where(fmt.Sprintf("%s.tag = ?", logTableName), params.Tag)
	}
	if params.LikeMessage != "" {
		db = db.Where(fmt.Sprintf("%s.message LIKE ?", logTableName), "%"+params.LikeMessage+"%")
	}
	if params.StartTime != "" && params.EndTime != "" {
		db = db.Where(fmt.Sprintf("%s.created_at BETWEEN ? AND ?", logTableName), params.StartTime, params.EndTime)
	}

	// 计算总记录数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	} else if total == 0 {
		return nil, errors.NotFound("没有符合条件的日志")
	}

	// 获取分页数据
	offset := (params.Page - 1) * params.PageSize
	db = db.Offset(offset).Limit(params.PageSize).Order(fmt.Sprintf("%s.created_at DESC", logTableName))

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
