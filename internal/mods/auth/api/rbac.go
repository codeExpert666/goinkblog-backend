package api

import (
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
	"strconv"
)

// CasbinHandler Casbin API处理器
type CasbinHandler struct {
	CasbinService *biz.CasbinService
}

// ListPolicies 获取策略列表
// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 获取策略列表（仅管理员可用）
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页数量" minimum(1) maximum(100)
// @Param type query string false "策略类型(p或g)" Enums(p, g)
// @Param subject query string false "主体"
// @Param object query string false "资源或角色，若策略类型为p，则为资源；若策略类型为g，则为角色"
// @Param action query string false "动作，策略类型为g时，该请求参数无效"
// @Success 200 {object} util.ResponseResult{data=schema.ListPolicyResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policies [get]
func (h *CasbinHandler) ListPolicies(c *gin.Context) {
	var req schema.ListPolicyRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CasbinService.ListPolicies(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// AddPolicy 添加策略
// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 添加策略（仅管理员可用）
// @Param type body string true "策略类型(p或g)" Enums(p, g)
// @Param subject body string true "主体"
// @Param object body string true "资源或角色，若策略类型为p，则为资源；若策略类型为g，则为角色"
// @Param action body string false "动作，策略类型为p时必需"
// @Success 200 {object} util.ResponseResult{data=schema.PolicyItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult "策略已存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policy [post]
func (h *CasbinHandler) AddPolicy(c *gin.Context) {
	var req schema.CreatePolicyRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	// 当策略类型为g时，清洗 action 项
	if req.Type == "g" {
		req.Action = ""
	}

	ctx := c.Request.Context()
	data, err := h.CasbinService.AddPolicy(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// DeletePolicy 移除策略
// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 移除策略（仅管理员可用）
// @Param id path uint true "策略ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult "策略不存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policy/{id} [delete]
func (h *CasbinHandler) DeletePolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		util.ResError(c, errors.BadRequest("无效的策略ID"))
		return
	}

	ctx := c.Request.Context()
	err = h.CasbinService.DeletePolicy(ctx, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// Enforce 权限验证
// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 权限验证测试（仅管理员可用）
// @Param subject body string true "主体"
// @Param object body string true "资源"
// @Param action body string true "动作"
// @Success 200 {object} util.ResponseResult{data=schema.EnforcerResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/enforce [post]
func (h *CasbinHandler) Enforce(c *gin.Context) {
	var req schema.EnforcerRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CasbinService.Enforce(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
