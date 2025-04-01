package dal

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetCategoryDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.Category{})
}

// CategoryRepository 分类数据访问层
type CategoryRepository struct {
	DB *gorm.DB
}

// Create 创建分类
func (r *CategoryRepository) Create(ctx context.Context, category *schema.Category) error {
	result := GetCategoryDB(ctx, r.DB).Create(category)
	return errors.WithStack(result.Error)
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, category *schema.Category) error {
	result := GetCategoryDB(ctx, r.DB).Where("id = ?", category.ID).Select("*").Omit("created_at").Updates(category)
	return errors.WithStack(result.Error)
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	// 检查是否有文章使用此分类
	var count int64
	if err := GetArticleDB(ctx, r.DB).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return errors.WithStack(err)
	}
	if count > 0 {
		return errors.Conflict("该分类下有文章，无法删除")
	}

	result := GetCategoryDB(ctx, r.DB).Where("id = ?", id).Delete(&schema.Category{})
	return errors.WithStack(result.Error)
}

// GetByID 通过ID获取分类
func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*schema.Category, error) {
	var category schema.Category
	err := GetCategoryDB(ctx, r.DB).Where("id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("分类不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &category, nil
}

// GetByName 通过名称获取分类
func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*schema.Category, error) {
	var category schema.Category
	err := GetCategoryDB(ctx, r.DB).Where("name = ?", name).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("分类不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &category, nil
}

// GetAll 获取所有分类
func (r *CategoryRepository) GetAll(ctx context.Context) ([]schema.CategoryResponse, error) {
	var categories []schema.Category
	if err := GetCategoryDB(ctx, r.DB).Find(&categories).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var responses []schema.CategoryResponse
	for _, category := range categories {
		responses = append(responses, schema.CategoryResponse{
			ID:           category.ID,
			Name:         category.Name,
			Description:  category.Description,
			ArticleCount: r.GetCategoryArticleCount(ctx, category.ID),
			CreatedAt:    category.CreatedAt,
			UpdatedAt:    category.UpdatedAt,
		})
	}

	return responses, nil
}

// GetList 获取分类列表（带分页）
func (r *CategoryRepository) GetList(ctx context.Context, page, pageSize int) (*schema.CategoryPaginationResult, error) {
	var result schema.CategoryPaginationResult

	// 默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算总数
	var total int64
	if err := GetCategoryDB(ctx, r.DB).Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 查询数据
	var categories []schema.Category
	offset := (page - 1) * pageSize
	if err := GetCategoryDB(ctx, r.DB).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&categories).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应数据
	var items []schema.CategoryResponse
	for _, category := range categories {
		items = append(items, schema.CategoryResponse{
			ID:           category.ID,
			Name:         category.Name,
			Description:  category.Description,
			ArticleCount: r.GetCategoryArticleCount(ctx, category.ID),
			CreatedAt:    category.CreatedAt,
			UpdatedAt:    category.UpdatedAt,
		})
	}

	result.Items = items
	result.Total = total
	result.Page = page
	result.PageSize = pageSize
	result.TotalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return &result, nil
}

// GetCategoryArticleCount 获取指定分类下的文章数量
func (r *CategoryRepository) GetCategoryArticleCount(ctx context.Context, categoryID uint) int {
	var articleCount int64

	err := GetArticleDB(ctx, r.DB).Where("category_id = ?", categoryID).Count(&articleCount).Error
	if err != nil {
		logging.Context(ctx).Error("获取分类下的文章数量失败", zap.Uint("分类ID", categoryID), zap.Error(errors.WithStack(err)))
		return 0
	}
	return int(articleCount)
}
