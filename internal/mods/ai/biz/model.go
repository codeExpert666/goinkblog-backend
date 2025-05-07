package biz

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
)

// ModelService 模型业务逻辑
type ModelService struct {
	ModelRepository *dal.ModelRepository
}

// GetModel 根据ID获取模型配置
func (s *ModelService) GetModel(ctx context.Context, id uint) (*schema.Model, error) {
	// 检查是否存在
	exist, err := s.ModelRepository.Exist(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound("不存在ID为 %d 的文章", id)
	}
	return s.ModelRepository.Get(ctx, id)
}

// CreateModel 创建模型配置
func (s *ModelService) CreateModel(ctx context.Context, req *schema.CreateModelRequest) (*schema.Model, error) {
	// 检查模型存在性
	if exist, err := s.ModelRepository.Exist(ctx, map[string]interface{}{
		"provider":   req.Provider,
		"api_key":    req.APIKey,
		"endpoint":   req.Endpoint,
		"model_name": req.ModelName,
	}); err != nil {
		return nil, err
	} else if exist {
		return nil, errors.Conflict("相关模型已存在")
	}

	model := &schema.Model{
		Provider:    req.Provider,
		APIKey:      req.APIKey,
		Endpoint:    req.Endpoint,
		ModelName:   req.ModelName,
		Temperature: *req.Temperature,
		Timeout:     req.Timeout,
		Active:      req.Active,
		Description: req.Description,
		RPM:         req.RPM,
		Weight:      req.Weight,
	}

	if err := s.ModelRepository.Create(ctx, model); err != nil {
		return nil, err
	}

	// GORM 在执行 CreateModel 操作时通常会尝试将数据库生成的主键值回写到模型实例中，
	// 但对于其他自动生成的字段（如 created_at 和 updated_at ），回写行为并不总是可靠的，尤其是在某些特定的 GORM 版本或配置下。
	// 最佳实践是重新查询以获取完整的数据库记录，这里 ID 一定存在，不可能 404
	return s.ModelRepository.Get(ctx, model.ID)
}

// UpdateModel 更新模型配置
func (s *ModelService) UpdateModel(ctx context.Context, id uint, req *schema.UpdateModelRequest) (*schema.Model, error) {
	// 检查 id 是否有效
	exist, err := s.ModelRepository.Exist(ctx, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, errors.NotFound("模型 %d 不存在", id)
	}

	updates := make(map[string]interface{})

	// 更改字段
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.Endpoint != nil {
		updates["endpoint"] = *req.Endpoint
	}
	if req.ModelName != nil {
		updates["model_name"] = *req.ModelName
	}
	if req.Temperature != nil {
		updates["temperature"] = *req.Temperature
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.RPM != nil {
		updates["rpm"] = *req.RPM
	}
	if req.Weight != nil {
		updates["weight"] = *req.Weight
	}

	// 保存更新
	if err := s.ModelRepository.Update(ctx, id, updates, true); err != nil {
		return nil, err
	}

	// 重新查询以获得最新记录
	return s.ModelRepository.Get(ctx, id)
}

// DeleteModel 删除模型配置
func (s *ModelService) DeleteModel(ctx context.Context, id uint) error {
	// 检查模型是否存在
	if exist, err := s.ModelRepository.Exist(ctx, map[string]interface{}{"id": id}); err != nil {
		return err
	} else if !exist {
		return errors.NotFound("模型不存在")
	}

	return s.ModelRepository.Delete(ctx, id)
}

// ListModels 列出所有模型配置
func (s *ModelService) ListModels(ctx context.Context, req *schema.ListModelsRequest) (*schema.ListModelsResponse, error) {
	return s.ModelRepository.List(ctx, req)
}

// ResetStats 重置模型统计信息
func (s *ModelService) ResetStats(ctx context.Context, id uint) error {
	return s.ModelRepository.Update(ctx, id, map[string]interface{}{
		"success_count": 0,
		"failure_count": 0,
		"avg_latency":   0,
	}, true)
}

// ResetAllStats 重置所有模型统计信息
func (s *ModelService) ResetAllStats(ctx context.Context) error {
	return s.ModelRepository.UpdateAll(ctx, map[string]interface{}{
		"success_count": 0,
		"failure_count": 0,
		"avg_latency":   0,
	})
}

// GetOverviewStats 获取总体统计数据
func (s *ModelService) GetOverviewStats(ctx context.Context) *schema.ModelOverallStatsResponse {
	return s.ModelRepository.GetOverviewStats(ctx)
}
