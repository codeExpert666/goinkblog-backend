package biz

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
)

// CategoryService 分类业务逻辑层
type CategoryService struct {
	CategoryRepository *dal.CategoryRepository
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(ctx context.Context, req *schema.CreateCategoryRequest) (*schema.CategoryResponse, error) {
	// 检查分类是否已存在
	_, err := s.CategoryRepository.GetByName(ctx, req.Name)
	if err == nil { // 已存在
		return nil, errors.Conflict("分类已存在")
	} else if errors.IsNotFound(err) { // 不存在
		// 创建分类
		category := &schema.Category{
			Name:        req.Name,
			Description: req.Description,
		}

		if err := s.CategoryRepository.Create(ctx, category); err != nil {
			return nil, err
		}

		// 构造响应数据
		response := &schema.CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			ArticleCount: s.CategoryRepository.GetCategoryArticleCount(ctx, category.ID),
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		}

		return response, nil
	} else {
		return nil, err
	}
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(ctx context.Context, id uint, req *schema.UpdateCategoryRequest) (*schema.CategoryResponse, error) {
	// 获取分类
	category, err := s.CategoryRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新分类字段
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}

	// 保存更新
	if err := s.CategoryRepository.Update(ctx, category); err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ArticleCount: s.CategoryRepository.GetCategoryArticleCount(ctx, category.ID),
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return response, nil
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(ctx context.Context, id uint) error {
	return s.CategoryRepository.Delete(ctx, id)
}

// GetCategoryByID 通过ID获取分类
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uint) (*schema.CategoryResponse, error) {
	// 获取分类
	category, err := s.CategoryRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ArticleCount: s.CategoryRepository.GetCategoryArticleCount(ctx, category.ID),
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	return response, nil
}

// GetAllCategories 获取所有分类
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]schema.CategoryResponse, error) {
	return s.CategoryRepository.GetAll(ctx)
}

// GetCategoryList 获取分类列表（带分页）
func (s *CategoryService) GetCategoryList(ctx context.Context, params *schema.CategoryQueryParams) (*schema.CategoryPaginationResult, error) {
	return s.CategoryRepository.GetList(ctx, params)
}
