package dal

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetTagDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.Tag{})
}

// TagRepository 标签数据访问层
type TagRepository struct {
	DB *gorm.DB
}

// Create 创建标签
func (r *TagRepository) Create(ctx context.Context, tag *schema.Tag) error {
	result := GetTagDB(ctx, r.DB).Create(tag)
	return errors.WithStack(result.Error)
}

// Update 更新标签
func (r *TagRepository) Update(ctx context.Context, tag *schema.Tag) error {
	result := GetTagDB(ctx, r.DB).Where("id = ?", tag.ID).Select("*").Omit("created_at").Updates(tag)
	return errors.WithStack(result.Error)
}

// Delete 删除标签
func (r *TagRepository) Delete(ctx context.Context, id uint) error {
	result := GetTagDB(ctx, r.DB).Where("id = ?", id).Delete(&schema.Tag{})
	return errors.WithStack(result.Error)
}

// GetByID 通过ID获取标签
func (r *TagRepository) GetByID(ctx context.Context, id uint) (*schema.Tag, error) {
	var tag schema.Tag
	err := GetTagDB(ctx, r.DB).Where("id = ?", id).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("标签不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &tag, nil
}

// GetByName 通过名称获取标签
func (r *TagRepository) GetByName(ctx context.Context, name string) (*schema.Tag, error) {
	var tag schema.Tag
	err := GetTagDB(ctx, r.DB).Where("name = ?", name).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("标签不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &tag, nil
}

// GetAll 获取所有标签
func (r *TagRepository) GetAll(ctx context.Context) ([]schema.TagResponse, error) {
	var tags []schema.Tag
	if err := GetTagDB(ctx, r.DB).Find(&tags).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	var response []schema.TagResponse
	for _, tag := range tags {
		response = append(response, schema.TagResponse{
			ID:           tag.ID,
			Name:         tag.Name,
			ArticleCount: r.GetTagArticleCount(ctx, tag.ID),
			CreatedAt:    tag.CreatedAt,
			UpdatedAt:    tag.UpdatedAt,
		})
	}
	return response, nil
}

// GetList 获取标签列表（带分页）
func (r *TagRepository) GetList(ctx context.Context, page, pageSize int) (*schema.TagPaginationResult, error) {
	var result schema.TagPaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算总数
	var total int64
	if err := GetTagDB(ctx, r.DB).Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 查询数据
	var tags []schema.Tag
	offset := (page - 1) * pageSize
	if err := GetTagDB(ctx, r.DB).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tags).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.TagResponse
	for _, tag := range tags {
		items = append(items, schema.TagResponse{
			ID:           tag.ID,
			Name:         tag.Name,
			ArticleCount: r.GetTagArticleCount(ctx, tag.ID),
			CreatedAt:    tag.CreatedAt,
			UpdatedAt:    tag.UpdatedAt,
		})
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}

// GetHotTags 获取热门标签
func (r *TagRepository) GetHotTags(ctx context.Context, limit int) ([]schema.TagResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	var result []schema.TagResponse

	tagName := new(schema.Tag).TableName()
	articleTagName := new(schema.ArticleTag).TableName()
	err := GetTagDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.id, %s.name, COUNT(a.article_id) as article_count, %s.created_at, %s.updated_at", tagName, tagName, tagName, tagName)).
		Joins(fmt.Sprintf("JOIN %s a ON %s.id = a.tag_id", articleTagName, tagName)).
		Group(fmt.Sprintf("%s.id, %s.name, %s.created_at, %s.updated_at", tagName, tagName, tagName, tagName)).
		Order("article_count DESC").
		Limit(limit).
		Scan(&result).Error

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// GetTagArticleCount 获取指定标签下的文章数量
func (r *TagRepository) GetTagArticleCount(ctx context.Context, tagID uint) int {
	var articleCount int64

	err := GetArticleTagDB(ctx, r.DB).Where("tag_id = ?", tagID).Count(&articleCount).Error
	if err != nil {
		logging.Context(ctx).Error("获取标签下的文章数量失败", zap.Uint("标签ID", tagID), zap.Error(err))
		return 0
	}
	return int(articleCount)
}
