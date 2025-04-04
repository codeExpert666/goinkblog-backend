package api

import (
	"github.com/gin-gonic/gin"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// AuthHandler 认证API处理器
type AuthHandler struct {
	AuthService *biz.AuthService
}

// @Tags AuthAPI
// @Summary 获取验证码ID
// @Success 200 {object} util.ResponseResult{data=schema.Captcha}
// @Router /api/auth/captcha/id [get]
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.AuthService.GetCaptcha(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags AuthAPI
// @Summary 响应验证码图片（宽400px x 高160px）
// @Param id query string true "验证码ID"
// @Param reload query number false "重新生成验证码（reload=1）"
// @Produce image/png
// @Success 200 {string} string "验证码图片 - 尺寸: 400x160 像素"
// @Failure 404 {object} util.ResponseResult
// @Router /api/auth/captcha/image [get]
func (h *AuthHandler) ResponseCaptcha(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.AuthService.ResponseCaptcha(ctx, c.Writer, c.Query("id"), c.Query("reload") == "1")
	if err != nil {
		util.ResError(c, err)
		return
	}
}

// @Tags AuthAPI
// @Summary 用户以邮箱、用户名和密码注册
// @Param username body string true "用户名" minLength(3) maxLength(20)
// @Param email body string true "邮箱" format(email)
// @Param password body string true "密码" minLength(6)
// @Success 200 {object} util.ResponseResult{data=schema.LoginResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 409 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req schema.RegisterRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := h.AuthService.Register(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags AuthAPI
// @Summary 用户以用户名、密码和验证码登录
// @Param username body string true "用户名"
// @Param password body string true "密码"
// @Param captcha body string true "验证码"
// @Param captcha_id body string true "验证码ID"
// @Success 200 {object} util.ResponseResult{data=schema.LoginResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req schema.LoginRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	loginResponse, err := h.AuthService.Login(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, loginResponse)
}

// @Tags AuthAPI
// @Security ApiKeyAuth
// @Summary 用户退出登录
// @Success 200 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	err := h.AuthService.Logout(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags AuthAPI
// @Security ApiKeyAuth
// @Summary 获取当前用户信息
// @Success 200 {object} util.ResponseResult{data=schema.UserResponse}
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/currentUser [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := h.AuthService.GetCurrentUser(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, user)
}

// @Tags AuthAPI
// @Security ApiKeyAuth
// @Summary 更新用户资料
// @Param username body string false "用户名" minLength(3) maxLength(20)
// @Param email body string false "邮箱" format(email)
// @Param bio body string false "个人简介"
// @Param avatar body string false "头像URL"
// @Param password body string false "密码"
// @Param old_password body string false "旧密码"
// @Success 200 {object} util.ResponseResult{data=schema.UserResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	var req schema.UpdateProfileRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	user, err := h.AuthService.UpdateProfile(ctx, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, user)
}

// @Tags AuthAPI
// @Security ApiKeyAuth
// @Summary 上传用户头像
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像图片文件"
// @Success 200 {object} util.ResponseResult{data=schema.AvatarResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/auth/avatar [post]
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	data, err := h.AuthService.UploadAvatar(c)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
