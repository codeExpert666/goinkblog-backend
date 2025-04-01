package biz

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/jwtx"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证业务逻辑层
type AuthService struct {
	UserRepository *dal.UserRepository
	Auth           jwtx.Auther
	Cache          cachex.Cacher
}

// ParseUserID 解析用户ID（中间件使用）
func (s *AuthService) ParseUserID(c *gin.Context) (uint, error) {
	adminID := config.C.General.Admin.ID
	// 如果禁用了认证中间件，直接返回管理员用户名
	if config.C.Middleware.Auth.Disable {
		return adminID, nil
	}

	invalidToken := errors.Unauthorized("无效的访问令牌")
	// 从请求中获取token
	token := util.GetToken(c)
	if token == "" {
		return 0, invalidToken
	}

	ctx := c.Request.Context()
	ctx = util.NewUserToken(ctx, token)

	// 解析token中的用户名
	userID, err := s.Auth.ParseSubject(ctx, token)
	if err != nil {
		if errors.Is(err, jwtx.ErrInvalidToken) {
			return 0, invalidToken
		}
		return 0, err
	} else if userID == adminID {
		// 如果是管理员用户，设置特殊标记
		c.Request = c.Request.WithContext(util.NewIsAdminUser(ctx))
		return userID, nil
	}

	// 用户名有效性校验
	// 从缓存中获取用户信息
	userIDStr := fmt.Sprintf("%d", userID)
	userCacheVal, ok, err := s.Cache.Get(ctx, config.CacheNSForUser, userIDStr)
	if err != nil {
		return 0, err
	} else if ok {
		userCache := util.ParseUserCache(userCacheVal)
		c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
		return userID, nil
	}

	// 缓存中未找到用户信息，则从数据库检查用户状态，若用户不存在则强制登出
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, invalidToken
		}
		return 0, err
	}

	// 将用户信息（角色）存入缓存
	userCache := util.UserCache{
		Role: user.Role,
	}
	err = s.Cache.Set(ctx, config.CacheNSForUser, userIDStr, userCache.String())
	if err != nil {
		return 0, err
	}

	c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
	return userID, nil
}

// GetCaptcha 生成新的验证码
// 返回验证码ID，验证码长度由配置决定
func (s *AuthService) GetCaptcha(ctx context.Context) (*schema.Captcha, error) {
	return &schema.Captcha{
		CaptchaID: captcha.NewLen(config.C.Util.Captcha.Length),
	}, nil
}

// ResponseCaptcha 响应验证码图片
// 生成验证码图片并设置相应的HTTP头
func (s *AuthService) ResponseCaptcha(ctx context.Context, w http.ResponseWriter, id string, reload bool) error {
	if reload && !captcha.Reload(id) {
		return errors.NotFound("找不到对应的验证码ID")
	}

	err := captcha.WriteImage(w, id, config.C.Util.Captcha.Width, config.C.Util.Captcha.Height)
	if err != nil {
		if errors.Is(err, captcha.ErrNotFound) {
			return errors.NotFound("找不到对应的验证码ID")
		}
		return err
	}

	// 设置HTTP响应头，确保验证码图片不被缓存，确保每次请求都会从服务器获取新的验证码图片
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache") // 向后兼容
	w.Header().Set("Expires", "0")       // 双重保证且向后兼容
	w.Header().Set("Content-Type", "image/png")
	return nil
}

func (s *AuthService) genUserToken(ctx context.Context, userID uint) (*schema.LoginToken, error) {
	// 生成JWT令牌
	token, err := s.Auth.GenerateToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokenBuf, err := token.EncodeToJSON()
	if err != nil {
		return nil, err
	}
	logging.Context(ctx).Info("生成用户访问令牌", zap.String("token", string(tokenBuf)))

	return &schema.LoginToken{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
		ExpiresAt:   token.GetExpiresAt(),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *schema.RegisterRequest) (*schema.LoginResponse, error) {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &schema.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // 默认角色
	}

	if err := s.UserRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	// 生成用户访问令牌
	token, err := s.genUserToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 构造响应
	response := &schema.LoginResponse{
		Token: *token,
		User: schema.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Bio:       user.Bio,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *schema.LoginRequest) (*schema.LoginResponse, error) {
	// 验证验证码
	if !captcha.VerifyString(req.CaptchaID, req.Captcha) {
		return nil, errors.BadRequest("验证码错误")
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogin)

	// 获取用户
	user, err := s.UserRepository.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.BadRequest("用户名或密码错误")
		}
		return nil, err
	}

	// 验证密码
	if !s.UserRepository.CheckPassword(ctx, user, req.Password) {
		return nil, errors.BadRequest("用户名或密码错误")
	}

	// 生成JWT令牌
	token, err := s.genUserToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	ctx = logging.NewUserID(ctx, user.ID)

	// 处理管理员用户登录
	if user.Role == "admin" {
		logging.Context(ctx).Info("管理员用户登录")
	} else { // 设置用户缓存和角色信息
		userIDStr := fmt.Sprintf("%d", user.ID)
		userCache := util.UserCache{
			Role: user.Role,
		}
		err = s.Cache.Set(ctx, config.CacheNSForUser, userIDStr, userCache.String(),
			time.Duration(config.C.Dictionary.UserCacheExp))
		if err != nil {
			logging.Context(ctx).Error("设置用户缓存失败", zap.Error(err))
		}
		logging.Context(ctx).Info("登录成功", zap.Uint("userID", user.ID), zap.String("username", user.Username))
	}

	// 构造响应
	response := &schema.LoginResponse{
		Token: *token,
		User: schema.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Bio:       user.Bio,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(ctx context.Context) (*schema.UserResponse, error) {
	userID := util.FromUserID(ctx)
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &schema.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return response, nil
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(ctx context.Context, req *schema.UpdateProfileRequest) (*schema.UserResponse, error) {
	if util.FromIsAdminUser(ctx) {
		return nil, errors.BadRequest("管理员资料不允许更新")
	}

	// 获取用户
	userID := util.FromUserID(ctx)
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 如果更新密码，验证旧密码并更新
	if req.Password != "" {
		if req.OldPassword == "" {
			return nil, errors.BadRequest("需要提供旧密码")
		}

		if !s.UserRepository.CheckPassword(ctx, user, req.OldPassword) {
			return nil, errors.BadRequest("旧密码错误")
		}

		// 更新密码
		if err := s.UserRepository.UpdatePassword(ctx, user.ID, req.Password); err != nil {
			return nil, err
		}
	}

	// 更新其他资料
	if req.Username != "" && req.Username != user.Username {
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		user.Email = req.Email
	}

	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// 保存更新
	if err := s.UserRepository.Update(ctx, user); err != nil {
		return nil, err
	}

	// 构造响应
	response := &schema.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return response, nil
}

// UploadAvatar 上传头像
func (s *AuthService) UploadAvatar(c *gin.Context) (string, error) {
	ctx := c.Request.Context()

	if util.FromIsAdminUser(ctx) {
		return "", errors.BadRequest("管理员资料不允许更新")
	}

	// 获取头像文件
	file, err := c.FormFile("avatar")
	if err != nil {
		return "", errors.BadRequest("获取图片文件失败: %s", err.Error())
	}

	// 验证文件类型
	if !util.IsImageFile(file.Filename) {
		return "", errors.BadRequest("支持的文件格式为: %s", config.SupportedImageFormats)
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// 设置保存路径
	avatarURL := filepath.Join("/pic/avatars", newFilename)
	dst := filepath.Join(config.C.Middleware.Static.Dir, avatarURL)

	// 保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		logging.Context(ctx).Error("保存图片文件失败", zap.String("filePath", dst), zap.Int64("fileSize", file.Size), zap.Error(err))
		return "", errors.InternalServerError("保存图片文件失败: %s", err.Error())
	}
	logging.Context(ctx).Info("保存图片文件成功", zap.String("filePath", dst), zap.Int64("fileSize", file.Size))

	return avatarURL, nil
}

// Logout 处理用户登出请求
func (s *AuthService) Logout(ctx context.Context) error {
	userToken := util.FromUserToken(ctx)
	if userToken == "" {
		return nil
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogout)
	if err := s.Auth.DestroyToken(ctx, userToken); err != nil {
		return err
	}

	userID := util.FromUserID(ctx)
	userIDStr := fmt.Sprintf("%d", userID)
	err := s.Cache.Delete(ctx, config.CacheNSForUser, userIDStr)
	if err != nil {
		logging.Context(ctx).Error("删除用户信息缓存失败", zap.Error(err))
	}
	logging.Context(ctx).Info("用户登出成功")

	return nil
}
