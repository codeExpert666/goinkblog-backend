package biz

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/schema"
)

// StatService 统计业务逻辑层
type StatService struct {
	StatRepository *dal.StatRepository
}

// GetArticleStatistic 获取文章统计信息
func (s *StatService) GetArticleStatistic(ctx context.Context) *schema.ArticleStatisticResponse {
	return s.StatRepository.GetArticleStatistic(ctx)
}

// GetVisitTrend 获取访问趋势数据
func (s *StatService) GetVisitTrend(ctx context.Context, days int) ([]schema.APIAccessTrendItem, error) {
	if days <= 0 {
		days = 7
	}
	return s.StatRepository.GetVisitTrend(ctx, days)
}

// GetUserActivityTrend 获取用户活跃度数据
func (s *StatService) GetUserActivityTrend(ctx context.Context, days int) ([]schema.UserActivityTrendItem, error) {
	if days <= 0 {
		days = 7
	}
	return s.StatRepository.GetUserActivityTrend(ctx, days)
}

// GetCategoryDistribution 获取文章分类分布
func (s *StatService) GetCategoryDistribution(ctx context.Context) ([]schema.CategoryDistItem, error) {
	return s.StatRepository.GetCategoryDistribution(ctx)
}

// GetLogList 获取日志列表
func (s *StatService) GetLogger(ctx context.Context, params *schema.LoggerQueryParams) (*schema.LoggerPaginationResult, error) {
	return s.StatRepository.GetLogList(ctx, params)
}
