package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// CommentHandler 评论API处理器
type CommentHandler struct {
	CommentService *biz.CommentService
}

// @Tags CommentAPI
// @Security ApiKeyAuth
// @Summary 获取评论列表（适用于审核和管理，仅管理员可用）
// @Param article_id query uint false "文章ID"
// @Param author_id query uint false "评论作者ID"
// @Param parent_id query uint false "父评论ID"
// @Param root_id query uint false "根评论ID"
// @Param level query int false "评论层级"
// @Param status query int false "评论状态：0-待审核，1-已通过，2-已拒绝" Enums(0, 1, 2)
// @Param keyword query string false "关键词"
// @Param create_start_time query string false "创建开始时间，格式：2006-01-02 15:04:05"
// @Param create_end_time query string false "创建结束时间，格式：2006-01-02 15:04:05"
// @Param review_start_time query string false "审核开始时间，格式：2006-01-02 15:04:05"
// @Param review_end_time query string false "审核结束时间，格式：2006-01-02 15:04:05"
// @Param reviewer_id query uint false "审核人员ID"
// @Param sort_by query string false "排序字段：create-创建时间，review-审核时间" Enums(create, review) default(create)
// @Param sort_order query string false "排序方式：desc-降序，asc-升序" Enums(desc, asc) default(desc)
// @Param page query int true "页码" minimum(1) default(1)
// @Param page_size query int true "页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/review [get]
func (h *CommentHandler) GetCommentsForReview(c *gin.Context) {
	var req schema.CommentReviewListRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CommentService.GetCommentsForReview(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CommentAPI
// @Security ApiKeyAuth
// @Summary 审核评论（仅管理员可用）
// @Param comment_id body uint true "评论ID" minimum(1)
// @Param status body int true "审核状态：1-通过，2-拒绝" Enums(1, 2)
// @Param review_remark body string false "审核备注"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/review [post]
func (h *CommentHandler) ReviewComment(c *gin.Context) {
	var req schema.ReviewCommentRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	err := h.CommentService.ReviewComment(ctx, userID, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags CommentAPI
// @Summary 获取文章的顶级评论
// @Param article_id path uint true "文章ID" minimum(1)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(30) default(10)
// @Param sort_by_create query string false "排序方式" Enums(asc, desc) default(desc)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/article/{article_id} [get]
func (h *CommentHandler) GetArticleComments(c *gin.Context) {
	// 解析文章ID（路径参数）
	articleID, err := strconv.ParseUint(c.Param("article_id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	// 绑定并验证查询参数
	var req schema.ArticleCommentsRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 查询数据
	ctx := c.Request.Context()
	data, err := h.CommentService.GetArticleComments(ctx, uint(articleID), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CommentAPI
// @Summary 获取评论详情
// @Param id path uint true "评论ID" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.CommentResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/{id} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的评论ID"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.CommentService.GetCommentByID(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CommentAPI
// @Security ApiKeyAuth
// @Summary 创建评论（会增加文章评论数、响应刚创建评论的全部信息）
// @Param article_id body uint true "文章ID" minimum(1)
// @Param content body string true "评论内容"
// @Param parent_id body integer false "父评论ID" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.CommentResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req schema.CreateCommentRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.CommentService.CreateComment(ctx, userID, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CommentAPI
// @Security ApiKeyAuth
// @Summary 删除评论（回复也会被删除、文章评论数也会减少）
// @Param id path uint true "评论ID" minimum(1)
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 403 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的评论ID"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	err = h.CommentService.DeleteComment(ctx, userID, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags CommentAPI
// @Summary 获取顶级评论的所有回复（扁平化列表，适合前端显示）
// @Param id path uint true "评论ID" minimum(1)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(30) default(10)
// @Param include_replies query bool false "是否包含回复的回复" default(false)
// @Param max_depth query int false "回复层数限制，i (i>0) 表示获取到第 i 层回复，0表示不限制" minimum(0) default(0)
// @Param sort_by_create query string false "排序方式" Enums(asc, desc) default(asc)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/{id}/replies [get]
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	// 解析评论ID（路径参数）
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的评论ID"))
		return
	}

	// 绑定并验证查询参数
	var req schema.CommentRepliesRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 查询数据
	ctx := c.Request.Context()
	data, err := h.CommentService.GetCommentReplies(ctx, uint(commentID), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CommentAPI
// @Security ApiKeyAuth
// @Summary 获取用户的评论
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(30) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/user [get]
func (h *CommentHandler) GetUserComments(c *gin.Context) {
	// 绑定并验证查询参数
	var req schema.UserCommentsRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.CommentService.GetUserComments(ctx, userID, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
