package dal

import (
	"context"
	"fmt"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

func GetModelDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.Model{})
}

// ModelRepository 模型数据访问层
type ModelRepository struct {
	Cache cachex.Cacher
	DB    *gorm.DB
}

// Exist 检测模型是否存在
func (r *ModelRepository) Exist(ctx context.Context, cond map[string]interface{}) (bool, error) {
	var count int64
	err := GetModelDB(ctx, r.DB).Where(cond).Count(&count).Error
	if err != nil {
		return false, errors.WithStack(err)
	}
	return count > 0, nil
}

// Create 创建模型
func (r *ModelRepository) Create(ctx context.Context, model *schema.Model) error {
	if err := GetModelDB(ctx, r.DB).Create(model).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if err := r.syncToSelector(ctx); err != nil {
		logging.Context(ctx).Error("向缓存中存入选择器同步标记失败", zap.Error(err))
	}

	return nil
}

// Update 更新单个模型
func (r *ModelRepository) Update(ctx context.Context, id uint, updates map[string]interface{}, sync bool) error {
	if err := GetModelDB(ctx, r.DB).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if sync {
		if err := r.syncToSelector(ctx); err != nil {
			logging.Context(ctx).Error("向缓存中存入选择器同步标记失败", zap.Error(err))
		}
	}

	return nil
}

// UpdateAll 更新所有模型
func (r *ModelRepository) UpdateAll(ctx context.Context, updates map[string]interface{}) error {
	if err := GetModelDB(ctx, r.DB).Where("1=1").Updates(updates).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if err := r.syncToSelector(ctx); err != nil {
		logging.Context(ctx).Error("向缓存中存入选择器同步标记失败", zap.Error(err))
	}

	return nil
}

// Delete 删除模型
func (r *ModelRepository) Delete(ctx context.Context, id uint) error {
	if err := GetModelDB(ctx, r.DB).Where("id = ?", id).Delete(&schema.Model{}).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if err := r.syncToSelector(ctx); err != nil {
		logging.Context(ctx).Error("向缓存中存入选择器同步标记失败", zap.Error(err))
	}

	return nil
}

// Get 获取模型
func (r *ModelRepository) Get(ctx context.Context, id uint) (*schema.Model, error) {
	var model schema.Model
	if err := GetModelDB(ctx, r.DB).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &model, nil
}

// List 列出所有模型
func (r *ModelRepository) List(ctx context.Context, req *schema.ListModelsRequest) (*schema.ListModelsResponse, error) {
	var resp schema.ListModelsResponse

	db := GetModelDB(ctx, r.DB)

	// 应用过滤条件
	if req.Provider != "" {
		db = db.Where("provider = ?", req.Provider)
	}
	if req.ModelName != "" {
		db = db.Where("model_name LIKE ?", "%"+req.ModelName+"%")
	}
	if req.Active != "" {
		db = db.Where("active = ?", req.Active)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用排序
	if req.SortByWeight == "desc" {
		db = db.Order("weight DESC")
	} else if req.SortByWeight == "asc" {
		db = db.Order("weight ASC")
	}
	if req.SortByRPM == "desc" {
		db = db.Order("rpm DESC")
	} else if req.SortByRPM == "asc" {
		db = db.Order("rpm ASC")
	}
	if req.SortByCurrentTokens == "desc" {
		db = db.Order("current_tokens DESC")
	} else if req.SortByCurrentTokens == "asc" {
		db = db.Order("current_tokens ASC")
	}
	if req.SortBySuccessCount == "desc" {
		db = db.Order("success_count DESC")
	} else if req.SortBySuccessCount == "asc" {
		db = db.Order("success_count ASC")
	}
	if req.SortByFailureCount == "desc" {
		db = db.Order("failure_count DESC")
	} else if req.SortByFailureCount == "asc" {
		db = db.Order("failure_count ASC")
	}
	if req.SortByAvgLatency != "" {
		// 排除 AvgLatency 为 0 的记录（表示尚不存在有效的延迟数据）
		db = db.Where("avg_latency > 0")
		if req.SortByAvgLatency == "desc" {
			db = db.Order("avg_latency DESC")
		} else {
			db = db.Order("avg_latency ASC")
		}
	}

	// 应用分页
	offset := (req.Page - 1) * req.PageSize
	db = db.Offset(offset).Limit(req.PageSize)

	// 查询数据
	var configs []*schema.Model
	if err := db.Find(&configs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 构造响应
	resp.Models = configs
	resp.Total = total
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &resp, nil
}

// ListActive 列出所有激活的模型
func (r *ModelRepository) ListActive(ctx context.Context) ([]*schema.Model, error) {
	var models []*schema.Model
	if err := GetModelDB(ctx, r.DB).Where("active = ?", true).Find(&models).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return models, nil
}

// GetOverviewStats 获取总体统计数据
func (r *ModelRepository) GetOverviewStats(ctx context.Context) *schema.ModelOverallStatsResponse {
	response := &schema.ModelOverallStatsResponse{}

	// 获取总模型数
	var totalModels int64
	if err := GetModelDB(ctx, r.DB).Count(&totalModels).Error; err != nil {
		logging.Context(ctx).Error("获取模型总数失败", zap.Error(err))
	}
	response.TotalModels = totalModels

	// 获取激活模型数
	var activeModels int64
	if err := GetModelDB(ctx, r.DB).Where("active = ?", true).Count(&activeModels).Error; err != nil {
		logging.Context(ctx).Error("获取激活模型总数失败", zap.Error(err))
	}
	response.ActiveModels = activeModels

	// 获取成功数、失败数、可用 token 数
	var totalSuccess, totalFailure int64
	var totalAvailableTokens int64
	if err := GetModelDB(ctx, r.DB).
		Select("SUM(success_count) as total_success, SUM(failure_count) as total_failure, SUM(current_tokens) as total_tokens").
		Row().Scan(&totalSuccess, &totalFailure, &totalAvailableTokens); err != nil {
		logging.Context(ctx).Error("获取模型响应成功数、响应失败数、可用 token 数失败", zap.Error(err))
	}
	response.TotalSuccess = totalSuccess
	response.TotalFailure = totalFailure
	response.TotalRequests = totalSuccess + totalFailure
	response.TotalAvailableTokens = totalAvailableTokens

	// 计算总体成功率
	if response.TotalRequests > 0 {
		response.OverallSuccessRate = float64(response.TotalSuccess) / float64(response.TotalRequests) * 100
	}

	// 计算总体平均延迟（加权平均）
	var totalLatency, totalSuccessForLatency float64
	if err := GetModelDB(ctx, r.DB).
		Where("avg_latency > 0").
		Select("SUM(avg_latency * success_count) as total_latency, SUM(success_count) as total_success_for_latency").
		Row().Scan(&totalLatency, &totalSuccessForLatency); err != nil {
		logging.Context(ctx).Error("获取模型总体平均响应延迟失败", zap.Error(err))
	}
	if totalSuccessForLatency > 0 {
		response.OverallAvgLatency = totalLatency / totalSuccessForLatency
	}

	// 获取使用最多的模型
	var mostUsedModel schema.Model
	if err := GetModelDB(ctx, r.DB).
		Order("(success_count + failure_count) DESC").
		First(&mostUsedModel).Error; err != nil {
		logging.Context(ctx).Error("获取使用最多模型的信息失败", zap.Error(err))
	}
	response.MostUsedModel.ID = mostUsedModel.ID
	response.MostUsedModel.Provider = mostUsedModel.Provider
	response.MostUsedModel.ModelName = mostUsedModel.ModelName
	response.MostUsedModel.RequestCount = mostUsedModel.SuccessCount + mostUsedModel.FailureCount

	// 获取成功率最高的模型
	var mostSuccessfulModel schema.Model
	if err := GetModelDB(ctx, r.DB).
		Where("success_count > 0").
		Order("(success_count / (success_count + failure_count)) DESC").
		First(&mostSuccessfulModel).Error; err != nil {
		logging.Context(ctx).Error("获取成功率最高模型的信息失败", zap.Error(err))
	}
	response.MostSuccessfulModel.ID = mostSuccessfulModel.ID
	response.MostSuccessfulModel.Provider = mostSuccessfulModel.Provider
	response.MostSuccessfulModel.ModelName = mostSuccessfulModel.ModelName
	totalRequests := mostSuccessfulModel.SuccessCount + mostSuccessfulModel.FailureCount
	if totalRequests > 0 {
		response.MostSuccessfulModel.SuccessRate = float64(mostSuccessfulModel.SuccessCount) / float64(totalRequests) * 100
	}

	// 获取最快的模型
	var fastestModel schema.Model
	if err := GetModelDB(ctx, r.DB).
		Where("avg_latency > 0").
		Order("avg_latency ASC").
		First(&fastestModel).Error; err != nil {
		logging.Context(ctx).Error("获取最快模型的信息失败", zap.Error(err))
	}
	response.FastestModel.ID = fastestModel.ID
	response.FastestModel.Provider = fastestModel.Provider
	response.FastestModel.ModelName = fastestModel.ModelName
	response.FastestModel.AvgLatency = fastestModel.AvgLatency

	return response
}

// syncToSelector 同步模型变更到 Selector
func (r *ModelRepository) syncToSelector(ctx context.Context) error {
	return r.Cache.Set(ctx, config.CacheNSForAI, config.CacheKeyForSyncToModelSelector,
		fmt.Sprintf("%d", time.Now().Unix()))
}
