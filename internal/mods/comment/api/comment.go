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
// @Summary 获取文章评论
// @Param articleId path uint true "文章ID" minimum(1)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(30) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/article/{article_id} [get]
func (h *CommentHandler) GetArticleComments(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("article_id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.CommentService.GetArticleComments(ctx, uint(articleID), page, pageSize)
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
// @Summary 创建评论
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
// @Summary 删除评论
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
// @Summary 获取评论的回复
// @Param id path uint true "评论ID" minimum(1)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(30) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.CommentPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/comment/{id}/replies [get]
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的评论ID"))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.CommentService.GetCommentReplies(ctx, uint(commentID), page, pageSize)
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
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.CommentService.GetUserComments(ctx, userID, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

