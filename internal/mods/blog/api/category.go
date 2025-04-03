package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// CategoryHandler 分类API处理器
type CategoryHandler struct {
	CategoryService *biz.CategoryService
}

// @Tags CategoryAPI
// @Summary 获取所有分类
// @Success 200 {object} util.ResponseResult{data=[]schema.CategoryResponse}
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := h.CategoryService.GetAllCategories(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CategoryAPI
// @Summary 获取分类列表（带分页）
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.CategoryPaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories/paginate [get]
func (h *CategoryHandler) GetCategoryList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
	}

	ctx := c.Request.Context()
	data, err := h.CategoryService.GetCategoryList(ctx, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CategoryAPI
// @Summary 获取分类详情
// @Param id path uint true "分类ID" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.CategoryResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的分类ID"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.CategoryService.GetCategoryByID(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CategoryAPI
// @Security ApiKeyAuth
// @Summary 创建分类（仅管理员可用）
// @Param name body string true "分类名称"
// @Param description body string false "分类描述"
// @Success 200 {object} util.ResponseResult{data=schema.CategoryResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req schema.CreateCategoryRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CategoryService.CreateCategory(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CategoryAPI
// @Security ApiKeyAuth
// @Summary 更新分类（仅管理员可用）
// @Param id path uint true "分类ID" minimum(1)
// @Param name body string false "分类名称"
// @Param description body string false "分类描述"
// @Success 200 {object} util.ResponseResult{data=schema.CategoryResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的分类ID"))
		return
	}

	var req schema.UpdateCategoryRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CategoryService.UpdateCategory(ctx, uint(id), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags CategoryAPI
// @Security ApiKeyAuth
// @Summary 删除分类（仅管理员可用）
// @Param id path uint true "分类ID" minimum(1)
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的分类ID"))
		return
	}

	ctx := c.Request.Context()
	err = h.CategoryService.DeleteCategory(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}
