package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// TagHandler 标签API处理器
type TagHandler struct {
	TagService *biz.TagService
}

// @Tags TagAPI
// @Summary 获取所有标签
// @Success 200 {object} util.ResponseResult{data=[]schema.TagResponse}
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags [get]
func (h *TagHandler) GetAllTags(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := h.TagService.GetAllTags(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags TagAPI
// @Summary 获取标签列表（带分页）
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.TagPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags/paginate [get]
func (h *TagHandler) GetTagList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页数参数"))
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.TagService.GetTagList(ctx, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags TagAPI
// @Summary 获取标签详情
// @Param id path uint true "标签ID" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.TagResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags/{id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的标签ID"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.TagService.GetTagByID(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags TagAPI
// @Security ApiKeyAuth
// @Summary 创建标签
// @Param name body string true "标签名称"
// @Success 200 {object} util.ResponseResult{data=schema.TagResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req schema.CreateTagRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.TagService.CreateTag(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags TagAPI
// @Security ApiKeyAuth
// @Summary 更新标签
// @Param id path uint true "标签ID" minimum(1)
// @Param name body string true "标签名称"
// @Success 200 {object} util.ResponseResult{data=schema.TagResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的标签ID"))
		return
	}

	var req schema.UpdateTagRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.TagService.UpdateTag(ctx, uint(id), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags TagAPI
// @Security ApiKeyAuth
// @Summary 删除标签
// @Param id path uint true "标签ID" minimum(1)
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的标签ID"))
		return
	}

	ctx := c.Request.Context()
	err = h.TagService.DeleteTag(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags TagAPI
// @Summary 获取热门标签
// @Param limit query int false "限制" minimum(1) maximum(50) default(10)
// @Success 200 {object} util.ResponseResult{data=[]schema.TagResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/tags/hot [get]
func (h *TagHandler) GetHotTags(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的限制参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.TagService.GetHotTags(ctx, limit)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
