package dal

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"gorm.io/gorm"
)

func GetArticleTagDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.ArticleTag{})
}

type ArticleTagRepository struct {
	DB *gorm.DB
}

// Create 创建文章标签关联
func (r *ArticleTagRepository) Create(ctx context.Context, articleTag *schema.ArticleTag) error {
	result := GetArticleTagDB(ctx, r.DB).Create(articleTag)
	return errors.WithStack(result.Error)
}

// DeleteByArticleID 根据文章 ID 删除文章标签关联
func (r *ArticleTagRepository) DeleteByArticleID(ctx context.Context, articleID uint) error {
	result := GetArticleTagDB(ctx, r.DB).Where("article_id = ?", articleID).Delete(&schema.ArticleTag{})
	return errors.WithStack(result.Error)
}

// DeleteByTagID 根据标签 ID 删除文章标签关联
func (r *ArticleTagRepository) DeleteByTagID(ctx context.Context, tagID uint) error {
	result := GetArticleTagDB(ctx, r.DB).Where("tag_id = ?", tagID).Delete(&schema.ArticleTag{})
	return errors.WithStack(result.Error)
}
