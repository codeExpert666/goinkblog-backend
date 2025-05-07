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
	if err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
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

// GetAll 获取所有分类（未来改为联表查询）
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
func (r *CategoryRepository) GetList(ctx context.Context, params *schema.CategoryQueryParams) (*schema.CategoryPaginationResult, error) {
	var result schema.CategoryPaginationResult

	// 默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 计算总数
	var totalCategories int64
	if err := GetCategoryDB(ctx, r.DB).Count(&totalCategories).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 计算有分类的文章总数
	var totalArticles int64
	if err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("status = ? AND category_id IS NOT NULL", "published").Count(&totalArticles).Error; err != nil {
		logging.Context(ctx).Error("获取有分类的文章总数失败", zap.Error(errors.WithStack(err)))
		totalArticles = 0
	}

	// 查询有文章的分类数量
	var categoriesWithArticle int64
	if err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).
		Select("COUNT(DISTINCT category_id)").
		Where("status = ? AND category_id IS NOT NULL", "published").
		Count(&categoriesWithArticle).Error; err != nil {
		logging.Context(ctx).Error("获取有文章的分类数量失败", zap.Error(errors.WithStack(err)))
		categoriesWithArticle = 0
	}

	// 查询文章数量最多的分类及其文章数量
	categoryTableName := new(schema.Category).TableName()
	articleTableName := new(schema.Article).TableName()

	var mostArticleCategory struct {
		CategoryName string 
		ArticleCount int64  
	}

	err := GetCategoryDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.name as category_name, COUNT(a.id) as article_count", categoryTableName)).
		Joins(fmt.Sprintf("LEFT JOIN %s a ON %s.id = a.category_id AND a.status = 'published'", articleTableName, categoryTableName)).
		Group(fmt.Sprintf("%s.id", categoryTableName)).
		Order("article_count DESC").
		Limit(1).
		Scan(&mostArticleCategory).Error

	if err != nil {
		logging.Context(ctx).Error("获取文章数量最多的分类失败", zap.Error(errors.WithStack(err)))
		mostArticleCategory.CategoryName = ""
		mostArticleCategory.ArticleCount = 0
	}

	// 关联查询分类数据
	db := GetCategoryDB(ctx, r.DB).
		Select(fmt.Sprintf("%s.*, COUNT(a.id) as article_count", categoryTableName)).
		Joins(fmt.Sprintf("LEFT JOIN %s a ON %s.id = a.category_id AND a.status = 'published'", articleTableName, categoryTableName)).
		Group(fmt.Sprintf("%s.id", categoryTableName))

	// 应用排序
	if params.SortByID == "asc" {
		db = db.Order(fmt.Sprintf("%s.id ASC", categoryTableName))
	} else if params.SortByID == "desc" {
		db = db.Order(fmt.Sprintf("%s.id DESC", categoryTableName))
	}

	if params.SortByArticleCount == "asc" {
		db = db.Order("article_count ASC")
	} else if params.SortByArticleCount == "desc" {
		db = db.Order("article_count DESC")
	}

	if params.SortByCreate == "asc" {
		db = db.Order(fmt.Sprintf("%s.created_at ASC", categoryTableName))
	} else if params.SortByCreate == "desc" {
		db = db.Order(fmt.Sprintf("%s.created_at DESC", categoryTableName))
	}

	if params.SortByUpdate == "asc" {
		db = db.Order(fmt.Sprintf("%s.updated_at ASC", categoryTableName))
	} else if params.SortByUpdate == "desc" {
		db = db.Order(fmt.Sprintf("%s.updated_at DESC", categoryTableName))
	}

	// 查询数据
	var categories []schema.CategoryResponse
	offset := (params.Page - 1) * params.PageSize
	if err := db.Offset(offset).Limit(params.PageSize).Find(&categories).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	result.Items = categories
	result.TotalCategories = totalCategories
	result.TotalArticles = totalArticles
	result.CategoriesWithArticle = categoriesWithArticle
	result.CategoryNameWithMostArticle = mostArticleCategory.CategoryName
	result.MostArticleCounts = mostArticleCategory.ArticleCount
	result.Page = params.Page
	result.PageSize = params.PageSize
	result.TotalPages = int((totalCategories + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &result, nil
}

// GetCategoryArticleCount 获取指定分类下的文章数量
func (r *CategoryRepository) GetCategoryArticleCount(ctx context.Context, categoryID uint) int {
	var articleCount int64

	err := GetArticleDB(ctx, r.DB).Model(&schema.Article{}).Where("status = ? AND category_id = ?", "published", categoryID).Count(&articleCount).Error
	if err != nil {
		logging.Context(ctx).Error("获取分类下的文章数量失败", zap.Uint("分类ID", categoryID), zap.Error(errors.WithStack(err)))
		return 0
	}
	return int(articleCount)
}
