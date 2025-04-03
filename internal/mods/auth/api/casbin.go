package api

import (
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

// CasbinHandler Casbin API处理器
type CasbinHandler struct {
	CasbinService *biz.CasbinService
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 获取策略列表（仅管理员可用）
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页数量" minimum(1) maximum(100)
// @Param type query string false "策略类型(p或g)" Enums(p, g)
// @Param subject query string false "主体(用户或角色)"
// @Param object query string false "对象(资源)"
// @Param action query string false "动作(操作)"
// @Success 200 {object} util.ResponseResult{data=schema.PolicyListResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policies [get]
func (h *CasbinHandler) GetPolicies(c *gin.Context) {
	var req schema.PolicyListRequest
	if err := util.ParseQuery(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.CasbinService.GetPolicies(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 添加策略（仅管理员可用）
// @Param subject body string true "主体(用户或角色)"
// @Param object body string true "对象(资源)"
// @Param action body string true "动作(操作)"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult "策略已存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policy [post]
func (h *CasbinHandler) AddPolicy(c *gin.Context) {
	var req schema.PolicyRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := h.CasbinService.AddPolicy(c.Request.Context(), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags RBACAPI	
// @Security ApiKeyAuth
// @Summary 移除策略（仅管理员可用）
// @Param subject body string true "主体(用户或角色)"
// @Param object body string true "对象(资源)"
// @Param action body string true "动作(操作)"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult "策略不存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/policy [delete]
func (h *CasbinHandler) RemovePolicy(c *gin.Context) {
	var req schema.PolicyRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := h.CasbinService.RemovePolicy(c.Request.Context(), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 为用户添加角色（仅管理员可用）
// @Param role body string true "角色"
// @Param user body string true "用户"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult "用户角色关系已存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/role [post]
func (h *CasbinHandler) AddRoleForUser(c *gin.Context) {
	var req schema.RoleRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := h.CasbinService.AddRoleForUser(c.Request.Context(), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 为用户移除角色（仅管理员可用）
// @Param role body string true "角色"
// @Param user body string true "用户"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult "用户角色关系不存在"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/role [delete]
func (h *CasbinHandler) RemoveRoleForUser(c *gin.Context) {
	var req schema.RoleRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	err := h.CasbinService.RemoveRoleForUser(c.Request.Context(), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 权限验证测试（仅管理员可用）
// @Param subject body string true "主体(用户或角色)"
// @Param object body string true "对象(资源)"
// @Param action body string true "动作(操作)"
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

	data, err := h.CasbinService.Enforce(c.Request.Context(), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 获取所有角色（仅管理员可用）
// @Success 200 {object} util.ResponseResult{data=map[string][]string{roles=[]string}}
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/roles [get]
func (h *CasbinHandler) GetAllRoles(c *gin.Context) {
	ctx := c.Request.Context()
	roles, err := h.CasbinService.GetAllRoles(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, gin.H{"roles": roles})
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 获取用户的所有角色（仅管理员可用）
// @Param user path string true "用户名"
// @Success 200 {object} util.ResponseResult{data=map[string][]string{roles=[]string}}
// @Failure 400 {object} util.ResponseResult "用户名不能为空"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/user/{user}/roles [get]
func (h *CasbinHandler) GetRolesForUser(c *gin.Context) {
	user := c.Param("user")
	if user == "" {
		util.ResError(c, errors.BadRequest("用户名不能为空"))
		return
	}

	ctx := c.Request.Context()
	roles, err := h.CasbinService.GetRolesForUser(ctx, user)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, gin.H{"roles": roles})
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 获取具有指定角色的所有用户（仅管理员可用）
// @Param role path string true "角色名"
// @Success 200 {object} util.ResponseResult{data=map[string][]string{users=[]string}}
// @Failure 400 {object} util.ResponseResult "角色名不能为空"
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/role/{role}/users [get]
func (h *CasbinHandler) GetUsersForRole(c *gin.Context) {
	role := c.Param("role")
	if role == "" {
		util.ResError(c, errors.BadRequest("角色名不能为空"))
		return
	}

	ctx := c.Request.Context()
	users, err := h.CasbinService.GetUsersForRole(ctx, role)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, gin.H{"users": users})
}

// @Tags RBACAPI
// @Security ApiKeyAuth
// @Summary 重新加载策略（仅管理员可用）
// @Success 200 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/rbac/reload [post]
func (h *CasbinHandler) ReloadPolicy(c *gin.Context) {
	ctx := c.Request.Context()
	err := h.CasbinService.ReloadPolicy(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}
