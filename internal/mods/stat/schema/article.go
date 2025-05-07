package schema

// ArticleVisitTrendItem 文章访问趋势数据项
type ArticleVisitTrendItem struct {
	Date       string `json:"date" example:"2023-01-01"`
	VisitCount int64  `json:"visit_count"`
}

// UserArticleVisitTrendResponse 用户文章访问趋势响应
type UserArticleVisitTrendResponse struct {
	Items      []ArticleVisitTrendItem `json:"items"`
	TotalVisit int64                   `json:"total_visit"`
}

// SiteOverviewResponse 统计数据响应结构
type SiteOverviewResponse struct {
	TotalUsers     int64 `json:"total_users"`
	TotalArticles  int64 `json:"total_articles"`
	TotalViews     int64 `json:"total_views"`
	TotalLikes     int64 `json:"total_comments"`
	TotalComments  int64 `json:"total_likes"`
	TotalFavorites int64 `json:"total_favorites"`
}

// CategoryDistItem 分类分布数据项
type CategoryDistItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// ArticleCreationTimeStatsItem 文章创作时间统计数据项
type ArticleCreationTimeStatsItem struct {
	Date       string           `json:"date" example:"2025-05-01"`
	Count      int64            `json:"count"`
	Categories map[string]int64 `json:"categories"`
}
