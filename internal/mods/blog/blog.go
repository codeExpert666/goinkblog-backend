package blog

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/api"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
)

// Blog 博客模块
type Blog struct {
	DB              *gorm.DB
	ArticleHandler  *api.ArticleHandler
	CategoryHandler *api.CategoryHandler
	TagHandler      *api.TagHandler
}

// Set 注入博客模块
var Set = wire.NewSet(
	wire.Struct(new(Blog), "*"),

	// 文章相关结构体
	wire.Struct(new(api.ArticleHandler), "*"),
	wire.Struct(new(biz.ArticleService), "*"),
	wire.Struct(new(dal.ArticleRepository), "*"),

	// 分类相关结构体
	wire.Struct(new(api.CategoryHandler), "*"),
	wire.Struct(new(biz.CategoryService), "*"),
	wire.Struct(new(dal.CategoryRepository), "*"),

	// 标签相关结构体
	wire.Struct(new(api.TagHandler), "*"),
	wire.Struct(new(biz.TagService), "*"),
	wire.Struct(new(dal.TagRepository), "*"),

	// 文章标签关联相关结构体
	wire.Struct(new(dal.ArticleTagRepository), "*"),

	// 用户交互相关结构体
	wire.Struct(new(dal.InteractionRepository), "*"),
)

// AutoMigrate 自动迁移数据库
func (b *Blog) AutoMigrate(ctx context.Context) error {
	return b.DB.AutoMigrate(
		&schema.Article{},
		&schema.Category{},
		&schema.Tag{},
		&schema.ArticleTag{},
		&schema.UserInteraction{},
	)
}

// Init 初始化博客模块
func (b *Blog) Init(ctx context.Context) error {
	if config.C.Storage.DB.AutoMigrate {
		if err := b.AutoMigrate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// RegisterRouters 注册路由
func (b *Blog) RegisterRouters(ctx context.Context, blog *gin.RouterGroup) error {
	// 文章接口
	articles := blog.Group("/articles")
	{
		articles.GET("", b.ArticleHandler.GetArticleList)
		articles.GET("/:id", b.ArticleHandler.GetArticle)
		articles.POST("", b.ArticleHandler.CreateArticle)
		articles.PUT("/:id", b.ArticleHandler.UpdateArticle)
		articles.DELETE("/:id", b.ArticleHandler.DeleteArticle)
		articles.POST("/upload-cover", b.ArticleHandler.UploadCover)
		articles.POST("/:id/like", b.ArticleHandler.LikeArticle)
		articles.POST("/:id/favorite", b.ArticleHandler.FavoriteArticle)
		articles.GET("/liked", b.ArticleHandler.GetUserLikedArticles)
		articles.GET("/favorites", b.ArticleHandler.GetUserFavoriteArticles)
		articles.GET("/history", b.ArticleHandler.GetUserViewHistory)
		articles.GET("/commented", b.ArticleHandler.GetUserCommentedArticles)
		articles.GET("/hot", b.ArticleHandler.GetHotArticles)
		articles.GET("/latest", b.ArticleHandler.GetLatestArticles)
	}

	// 分类接口
	categories := blog.Group("/categories")
	{
		categories.GET("", b.CategoryHandler.GetAllCategories)
		categories.GET("/paginate", b.CategoryHandler.GetCategoryList)
		categories.GET("/:id", b.CategoryHandler.GetCategory)
		categories.POST("", b.CategoryHandler.CreateCategory)
		categories.PUT("/:id", b.CategoryHandler.UpdateCategory)
		categories.DELETE("/:id", b.CategoryHandler.DeleteCategory)
	}

	// 标签接口
	tags := blog.Group("/tags")
	{
		tags.GET("", b.TagHandler.GetAllTags)
		tags.GET("/paginate", b.TagHandler.GetTagList)
		tags.GET("/:id", b.TagHandler.GetTag)
		tags.POST("", b.TagHandler.CreateTag)
		tags.PUT("/:id", b.TagHandler.UpdateTag)
		tags.DELETE("/:id", b.TagHandler.DeleteTag)
		tags.GET("/hot", b.TagHandler.GetHotTags)
	}

	return nil
}

// Release 释放资源
func (b *Blog) Release(ctx context.Context) error {
	return nil
}
