package schema

// APIAccessTrendItem API访问趋势数据项
type APIAccessTrendItem struct {
	Date             string `json:"date"`
	TotalCount       int64  `json:"total_count"`
	SuccessCount     int64  `json:"success_count"`
	ClientErrorCount int64  `json:"client_error_count"`
	ServerErrorCount int64  `json:"server_error_count"`
}

// UserActivityTrendItem 用户活跃度趋势数据项
type UserActivityTrendItem struct {
	Date      string `json:"date"`
	UserCount int64  `json:"user_count"`
}
