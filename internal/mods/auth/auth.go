package auth

import (
	"context"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
)

// Auth 认证模块
type Auth struct {
	DB            *gorm.DB
	AuthHandler   *api.AuthHandler
	CasbinHandler *api.CasbinHandler
}

// Set 注入 Auth 模块
var Set = wire.NewSet(
	wire.Struct(new(Auth), "*"),

	// 登录认证相关结构体
	wire.Struct(new(api.AuthHandler), "*"),
	wire.Struct(new(biz.AuthService), "*"),
	wire.Struct(new(dal.UserRepository), "*"),

	// Casbin相关结构体
	wire.Struct(new(api.CasbinHandler), "*"),
	wire.Struct(new(biz.CasbinService), "*"),
	wire.Struct(new(dal.CasbinRepository), "*"),
)

// AutoMigrate 自动迁移数据库表结构
func (a *Auth) AutoMigrate(ctx context.Context) error {
	return a.DB.AutoMigrate(
		new(schema.User),
		new(schema.CasbinRule),
	)
}

// Init 初始化认证模块
func (a *Auth) Init(ctx context.Context) error {
	// 根据配置自动迁移数据库表结构
	if config.C.Storage.DB.AutoMigrate {
		if err := a.AutoMigrate(ctx); err != nil {
			return err
		}
	}

	// 添加管理员信息
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(config.C.General.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = a.AuthHandler.AuthService.UserRepository.Create(ctx, &schema.User{
		Username: config.C.General.Admin.Username,
		Email:    config.C.General.Admin.Email,
		Password: string(pwdHash),
		Avatar:   config.C.General.Admin.Avatar,
		Bio:      config.C.General.Admin.Bio,
		Role:     config.C.General.Admin.Role,
	})
	if err == nil {
		logging.Context(ctx).Info("管理员账户创建成功")
	} else if errors.IsConflict(err) {
		logging.Context(ctx).Info("管理员账户已存在，跳过创建")
	} else {
		logging.Context(ctx).Error("添加管理员信息失败", zap.Error(err))
		return err
	}

	// 将 rbac_policy.csv 中的策略导入到数据库
	if path := config.C.Middleware.Casbin.PolicyFile; path != "" {
		policyFile := filepath.Join(config.C.General.WorkDir, path)
		err = a.CasbinHandler.CasbinService.ImportPolicyFromFile(ctx, policyFile)
		if err != nil {
			logging.Context(ctx).Error("导入 Casbin 策略到数据库失败", zap.Error(err))
			return err
		}
	}

	// 初始化 Casbin 执行器
	err = a.CasbinHandler.CasbinService.SetEnforcer(ctx)
	if err != nil {
		logging.Context(ctx).Error("初始化 Casbin 执行器失败", zap.Error(err))
		return err
	}

	return nil
}

// RegisterRouters 注册路由
func (a *Auth) RegisterRouters(ctx context.Context, auth *gin.RouterGroup) error {
	// 验证码接口
	captcha := auth.Group("/captcha")
	{
		captcha.GET("/id", a.AuthHandler.GetCaptcha)
		captcha.GET("/image", a.AuthHandler.ResponseCaptcha)
	}

	// 认证相关接口
	{
		auth.POST("/register", a.AuthHandler.Register)
		auth.POST("/login", a.AuthHandler.Login)
		auth.POST("/logout", a.AuthHandler.Logout)
		auth.GET("/currentUser", a.AuthHandler.GetCurrentUser)
		auth.PUT("/profile", a.AuthHandler.UpdateProfile)
		auth.POST("/avatar", a.AuthHandler.UploadAvatar)
	}

	// RBAC权限管理接口
	rbac := auth.Group("/rbac")
	{
		rbac.GET("/policies", a.CasbinHandler.GetPolicies)
		rbac.POST("/policy", a.CasbinHandler.AddPolicy)
		rbac.DELETE("/policy", a.CasbinHandler.RemovePolicy)
		rbac.POST("/role", a.CasbinHandler.AddRoleForUser)
		rbac.DELETE("/role", a.CasbinHandler.RemoveRoleForUser)
		rbac.POST("/enforce", a.CasbinHandler.Enforce)
		rbac.GET("/roles", a.CasbinHandler.GetAllRoles)
		rbac.GET("/user/:user/roles", a.CasbinHandler.GetRolesForUser)
		rbac.GET("/role/:role/users", a.CasbinHandler.GetUsersForRole)
		rbac.POST("/reload", a.CasbinHandler.ReloadPolicy)
	}

	return nil
}

// Release 释放模块资源
func (a *Auth) Release(ctx context.Context) error {
	return nil
}
