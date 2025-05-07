package comment

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/comment/schema"
)

// Comment 评论模块
type Comment struct {
	DB             *gorm.DB
	CommentHandler *api.CommentHandler
}

// Set 注入评论模块
var Set = wire.NewSet(
	wire.Struct(new(Comment), "*"),

	// 评论相关结构体
	wire.Struct(new(api.CommentHandler), "*"),
	wire.Struct(new(biz.CommentService), "*"),
	wire.Struct(new(dal.CommentRepository), "*"),
)

// AutoMigrate 自动迁移数据库
func (c *Comment) AutoMigrate(ctx context.Context) error {
	return c.DB.AutoMigrate(
		&schema.Comment{},
	)
}

// Init 初始化评论模块
func (c *Comment) Init(ctx context.Context) error {
	if config.C.Storage.DB.AutoMigrate {
		if err := c.AutoMigrate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// RegisterRouters 注册路由
func (c *Comment) RegisterRouters(ctx context.Context, comment *gin.RouterGroup) error {
	// 评论接口
	{
		comment.GET("/article/:article_id", c.CommentHandler.GetArticleComments)
		comment.GET("/:id", c.CommentHandler.GetComment)
		comment.POST("", c.CommentHandler.CreateComment)
		comment.DELETE("/:id", c.CommentHandler.DeleteComment)
		comment.GET("/:id/replies", c.CommentHandler.GetCommentReplies)
		comment.GET("/user", c.CommentHandler.GetUserComments)
		// 管理员接口
		comment.GET("/review", c.CommentHandler.GetCommentsForReview)
		comment.POST("/review", c.CommentHandler.ReviewComment)
	}
	return nil
}

// Release 释放资源
func (c *Comment) Release(ctx context.Context) error {
	return nil
}
