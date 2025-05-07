package biz

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// TagService 标签业务逻辑层
type TagService struct {
	TagRepository        *dal.TagRepository
	ArticleTagRepository *dal.ArticleTagRepository
	Trans                *util.Trans
}

// CreateTag 创建标签
func (s *TagService) CreateTag(ctx context.Context, req *schema.CreateTagRequest) (*schema.TagResponse, error) {
	// 检查标签是否已存在
	existingTag, err := s.TagRepository.GetByName(ctx, req.Name)
	if err == nil { // 已存在
		// 返回已存在的标签
		response := &schema.TagResponse{
			ID:        existingTag.ID,
			Name:      existingTag.Name,
			ArticleCount: s.TagRepository.GetTagArticleCount(ctx, existingTag.ID),
			CreatedAt: existingTag.CreatedAt,
			UpdatedAt: existingTag.UpdatedAt,
		}
		return response, nil
	} else if errors.IsNotFound(err) { // 不存在
		// 创建标签
		tag := &schema.Tag{
			Name: req.Name,
		}

		if err := s.TagRepository.Create(ctx, tag); err != nil {
			return nil, err
		}

		// 构造响应数据
		response := &schema.TagResponse{
			ID:        tag.ID,
			Name:      tag.Name,
			ArticleCount: s.TagRepository.GetTagArticleCount(ctx, tag.ID),
			CreatedAt: tag.CreatedAt,
			UpdatedAt: tag.UpdatedAt,
		}

		return response, nil
	} else {
		return nil, err
	}
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(ctx context.Context, id uint, req *schema.UpdateTagRequest) (*schema.TagResponse, error) {
	// 获取标签
	tag, err := s.TagRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查新名称是否已存在
	if req.Name != tag.Name {
		_, err := s.TagRepository.GetByName(ctx, req.Name)
		if err == nil {
			return nil, errors.Conflict("标签已存在")
		} else if !errors.IsNotFound(err) {
			return nil, err
		}
	}

	// 更新标签字段
	tag.Name = req.Name

	// 保存更新
	if err := s.TagRepository.Update(ctx, tag); err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		ArticleCount: s.TagRepository.GetTagArticleCount(ctx, tag.ID),
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}

	return response, nil
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(ctx context.Context, id uint) error {
	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 先删除关联关系
		if err := s.ArticleTagRepository.DeleteByTagID(ctx, id); err != nil {
			return err
		}
		// 再删除标签
		return s.TagRepository.Delete(ctx, id)
	})
}

// GetTagByID 通过ID获取标签
func (s *TagService) GetTagByID(ctx context.Context, id uint) (*schema.TagResponse, error) {
	// 获取标签
	tag, err := s.TagRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 构造响应数据
	response := &schema.TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		ArticleCount: s.TagRepository.GetTagArticleCount(ctx, tag.ID),
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}

	return response, nil
}

// GetAllTags 获取所有标签
func (s *TagService) GetAllTags(ctx context.Context) ([]schema.TagResponse, error) {
	return s.TagRepository.GetAll(ctx)
}

// GetTagList 获取标签列表（带分页）
func (s *TagService) GetTagList(ctx context.Context, params *schema.TagQueryParams) (*schema.TagPaginationResult, error) {
	return s.TagRepository.GetList(ctx, params)
}

// GetHotTags 获取热门标签
func (s *TagService) GetHotTags(ctx context.Context, limit int) ([]schema.TagResponse, error) {
	return s.TagRepository.GetHotTags(ctx, limit)
}
