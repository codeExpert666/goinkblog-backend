package mods

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat"
)

// API 路由前缀常量
const (
	apiPrefix = "/api/"
	aiPrefix  = "/api/ai/"
)

// Mods 所有模块的集合
type Mods struct {
	Auth    *auth.Auth
	Blog    *blog.Blog
	Comment *comment.Comment
	Stat    *stat.Stat
	AI      *ai.AI
}

// Set 定义注入器集合
var Set = wire.NewSet(
	wire.Struct(new(Mods), "*"),
	auth.Set,
	blog.Set,
	comment.Set,
	stat.Set,
	ai.Set,
)

// Init 初始化所有模块
func (a *Mods) Init(ctx context.Context) error {
	// 初始化 Auth 模块
	if err := a.Auth.Init(ctx); err != nil {
		return err
	}

	// 初始化Blog模块
	if err := a.Blog.Init(ctx); err != nil {
		return err
	}

	// 初始化Comment模块
	if err := a.Comment.Init(ctx); err != nil {
		return err
	}

	// 初始化Stat模块
	if err := a.Stat.Init(ctx); err != nil {
		return err
	}

	// 初始化AI模块
	if err := a.AI.Init(ctx); err != nil {
		return err
	}

	return nil
}

// RouterPrefixes 路由前缀列表
func (a *Mods) RouterPrefixes() []string {
	return []string{apiPrefix}
}

// AIRouterPrefixes AI路由前缀列表
func (a *Mods) AIRouterPrefixes() []string {
	return []string{aiPrefix}
}

// RegisterRouters 注册路由
func (a *Mods) RegisterRouters(ctx context.Context, e *gin.Engine) error {
	// 注册API路由
	gAPI := e.Group(apiPrefix)

	// 注册Auth模块路由
	authApi := gAPI.Group("auth")
	if err := a.Auth.RegisterRouters(ctx, authApi); err != nil {
		return err
	}

	// 注册Blog模块路由
	blogApi := gAPI.Group("blog")
	if err := a.Blog.RegisterRouters(ctx, blogApi); err != nil {
		return err
	}

	// 注册Comment模块路由
	commentApi := gAPI.Group("comment")
	if err := a.Comment.RegisterRouters(ctx, commentApi); err != nil {
		return err
	}

	// 注册Stat模块路由
	statApi := gAPI.Group("stat")
	if err := a.Stat.RegisterRouters(ctx, statApi); err != nil {
		return err
	}

	// 注册AI模块路由
	aiApi := e.Group(aiPrefix)
	if err := a.AI.RegisterRouters(ctx, aiApi); err != nil {
		return err
	}

	return nil
}

// Release 释放模块资源
func (a *Mods) Release(ctx context.Context) error {
	// 释放Auth模块资源
	if err := a.Auth.Release(ctx); err != nil {
		return err
	}

	// 释放Blog模块资源
	if err := a.Blog.Release(ctx); err != nil {
		return err
	}

	// 释放Comment模块资源
	if err := a.Comment.Release(ctx); err != nil {
		return err
	}

	// 释放Stat模块资源
	if err := a.Stat.Release(ctx); err != nil {
		return err
	}

	// 释放AI模块资源
	if err := a.AI.Release(ctx); err != nil {
		return err
	}

	return nil
}
