package schema

// 	CommentStatisticResponse 评论统计数据响应
type CommentStatisticResponse struct {
	TotalComments int64 `json:"total_comments"`
	PassedComments int64 `json:"passed_comments"`
	PendingComments int64 `json:"pending_comments"`
	RejectedComments int64 `json:"rejected_comments"`
}
