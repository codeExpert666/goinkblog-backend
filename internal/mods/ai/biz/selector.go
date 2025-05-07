package biz

import (
	"context"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/dal"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/ai/schema"
)

var (
	ErrNoAvailableModel = errors.ServiceUnavailableError("AI助手服务暂时不可用，请稍后再试")
	ErrRateLimited      = errors.TooManyRequests("AI服务需求量过高，系统暂时无法处理新请求")
)

// Selector 模型选择器
type Selector struct {
	models          []*schema.Model      `wire:"-"` // AI配置内存缓存
	mutex           sync.Mutex           `wire:"-"` // 控制 models 的并发访问
	modelTicker     *time.Ticker         `wire:"-"` // 定期自动加载模型配置
	weightTicker    *time.Ticker         `wire:"-"` // 定期自动更新模型权重
	Cache           cachex.Cacher        // 获取模型配置更新通知
	ModelRepository *dal.ModelRepository // 模型配置访问层
}

// Load 初始化模型选择器
func (s *Selector) Load(ctx context.Context) error {
	if err := s.loadModels(ctx); err != nil {
		return err
	}

	go s.autoLoadModels(ctx)
	return nil
}

// loadModels 从数据库加载模型配置
func (s *Selector) loadModels(ctx context.Context) error {
	models, err := s.ModelRepository.ListActive(ctx)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.models = models
	return nil
}

// autoLoadModels 定期加载最新模型列表
func (s *Selector) autoLoadModels(ctx context.Context) {
	var lastUpdated int64 // 记录最后更新时间
	s.modelTicker = time.NewTicker(time.Duration(config.C.AI.Selector.LoadModelsInterval) * time.Minute)
	for range s.modelTicker.C {
		// 从缓存中获取同步标记
		val, ok, err := s.Cache.Get(ctx, config.CacheNSForAI, config.CacheKeyForSyncToModelSelector)
		if err != nil {
			logging.Context(ctx).Error("从缓存中获取模型列表的同步标记失败", zap.Error(err),
				zap.String("cache_key", config.CacheKeyForSyncToModelSelector))
			continue
		} else if !ok {
			continue
		}

		// 解析同步标记
		updated, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logging.Context(ctx).Error("解析模型列表的同步标记失败", zap.Error(err),
				zap.String("val", val))
			continue
		}

		// 需要更新
		if lastUpdated < updated {
			if err := s.loadModels(ctx); err != nil {
				logging.Context(ctx).Error("更新选择器中的模型列表失败", zap.Error(err))
			} else {
				lastUpdated = updated
			}
		}
	}
}

// AutoWeightUpdate 定期更新权重
func (s *Selector) AutoWeightUpdate(ctx context.Context) {
	s.weightTicker = time.NewTicker(time.Duration(config.C.AI.Selector.UpdateWeightInterval) * time.Minute)
	go func() {
		for range s.weightTicker.C {
			if err := s.UpdateModelWeights(ctx); err != nil {
				logging.Context(ctx).Error("模型权重更新失败", zap.Error(err))
			}
		}
	}()
}

// SelectModel 选择一个可用的模型
func (s *Selector) SelectModel(ctx context.Context) (*schema.Model, func(success bool, latency time.Duration), error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.models) == 0 {
		return nil, nil, ErrNoAvailableModel
	}

	// 注意：SelectModel 方法会用在请求处理中，面临 ctx 随请求结束而取消的问题
	// 复制上下文以供异步操作使用
	newCtx := util.ExtendContext(context.Background(), ctx)
	newCtx = logging.ExtendLoggerContext(newCtx, ctx)

	// 使用加权轮询算法选择模型
	totalWeight := 0
	availableModels := make([]*schema.Model, 0, len(s.models))
	weightMap := make(map[int]int) // 索引到权重的映射

	now := time.Now()

	// 计算总权重并过滤出可用模型
	for _, model := range s.models {
		// 更新令牌桶
		if model.Active {
			// 计算需要补充的令牌数
			elapsed := now.Sub(model.LastRefillTime)
			tokensToAdd := int(elapsed.Minutes() * float64(model.RPM))

			if tokensToAdd > 0 {
				// 更新令牌数，不超过最大值
				model.CurrentTokens = min(model.RPM, model.CurrentTokens+tokensToAdd)
				model.LastRefillTime = now

				// 异步更新数据库
				// 这里不受管理员重置数据的影响
				go func(ctx context.Context, modelID uint, tokens int, lastRefill time.Time) {
					err := s.ModelRepository.Update(ctx, modelID, map[string]interface{}{
						"current_tokens":   tokens,
						"last_refill_time": lastRefill,
					}, false)
					if err != nil {
						logging.Context(ctx).Error("向数据库更新模型令牌数失败", zap.Error(err),
							zap.Uint("model_id", modelID), zap.Int("tokens", tokens), zap.Time("refill_time", lastRefill))
					}
				}(newCtx, model.ID, model.CurrentTokens, model.LastRefillTime)
			}

			// 检查是否有可用令牌
			if model.CurrentTokens > 0 {
				availableModels = append(availableModels, model)
				weightMap[len(availableModels)-1] = model.Weight
				totalWeight += model.Weight
			}
		}
	}

	if len(availableModels) == 0 {
		return nil, nil, ErrRateLimited
	}

	// 选择一个模型
	var selectedIdx int
	if totalWeight > 0 {
		// 加权随机选择
		r := rand.Intn(totalWeight)
		cumulativeWeight := 0
		for i, weight := range weightMap {
			cumulativeWeight += weight
			if r < cumulativeWeight {
				selectedIdx = i
				break
			}
		}
	} else {
		// 如果所有权重为0，则随机选择
		selectedIdx = rand.Intn(len(availableModels))
	}

	selectedModel := availableModels[selectedIdx]

	// 消耗一个令牌
	selectedModel.CurrentTokens--

	// 异步更新数据库
	go func(ctx context.Context, modelID uint, tokens int, lastRefill time.Time) {
		// 这里不受管理员重置数据的影响
		err := s.ModelRepository.Update(ctx, modelID, map[string]interface{}{
			"current_tokens":   tokens,
			"last_refill_time": lastRefill,
		}, false)
		if err != nil {
			logging.Context(ctx).Error("向数据库更新模型令牌数失败", zap.Error(err),
				zap.Uint("model_id", modelID), zap.Int("tokens", tokens), zap.Time("refill_time", lastRefill))
		}
	}(newCtx, selectedModel.ID, selectedModel.CurrentTokens, selectedModel.LastRefillTime)

	// 创建回调函数
	callback := func(success bool, latency time.Duration) {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		latencyMs := float64(latency.Milliseconds())

		if success {
			totalLatency := float64(selectedModel.SuccessCount)*selectedModel.AvgLatency + latencyMs
			selectedModel.SuccessCount++
			selectedModel.AvgLatency = totalLatency / float64(selectedModel.SuccessCount)
		} else {
			selectedModel.FailureCount++
		}

		// 异步更新统计数据
		go func(ctx context.Context, modelID uint, success bool, latency float64) {
			// 采用 gorm.Expr 是为了应对内存数据与数据库数据不一致的情况
			// 主要考虑到管理员可以重置数据
			if success {
				err := s.ModelRepository.Update(ctx, modelID, map[string]interface{}{
					"success_count": gorm.Expr("success_count + 1"),
					"avg_latency":   gorm.Expr("(success_count * avg_latency + ?) / (success_count + 1)", latency),
				}, false)
				if err != nil {
					logging.Context(ctx).Error("向数据库更新模型统计数据失败", zap.Error(err),
						zap.Uint("model_id", modelID), zap.Bool("model_success", success))
				}
			} else {
				err := s.ModelRepository.Update(ctx, modelID, map[string]interface{}{
					"failure_count": gorm.Expr("failure_count + 1"),
				}, false)
				if err != nil {
					logging.Context(ctx).Error("向数据库更新模型统计数据失败", zap.Error(err),
						zap.Uint("model_id", modelID), zap.Bool("model_success", success))
				}
			}
		}(newCtx, selectedModel.ID, success, latencyMs)
	}

	return selectedModel, callback, nil
}

// UpdateModelWeights 根据成功率和平均响应时间自动调整权重
func (s *Selector) UpdateModelWeights(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, model := range s.models {
		totalRequests := model.SuccessCount + model.FailureCount
		if totalRequests == 0 {
			continue
		}

		// 成功率
		successRate := float64(model.SuccessCount) / float64(totalRequests)

		// 根据成功率和响应时间调整权重
		// 成功率越高、平均延时越小，权重越高
		newWeight := int(100 * successRate * (1000 / (model.AvgLatency + 100)))
		if newWeight < 1 {
			newWeight = 1 // 最小权重为1
		}

		model.Weight = newWeight

		// 异步更新数据库
		go func(ctx context.Context, modelID uint) {
			// 采用 gorm.Expr 是为了应对内存数据与数据库数据不一致的情况
			// 主要考虑到管理员可以重置数据
			err := s.ModelRepository.Update(ctx, modelID, map[string]interface{}{
				"weight": gorm.Expr(`
					CASE 
						WHEN (success_count + failure_count) = 0 THEN weight
						ELSE GREATEST(1, CAST((100 * (success_count / (success_count + failure_count)) * (1000 / (avg_latency + 100))) AS SIGNED))
					END
				`),
			}, false)

			if err != nil {
				logging.Context(ctx).Error("向数据库更新模型权重失败", zap.Error(err),
					zap.Uint("model_id", modelID))
			}
		}(ctx, model.ID)
	}

	return nil
}

// Release 释放资源
func (s *Selector) Release(ctx context.Context) error {
	if s.modelTicker != nil {
		s.modelTicker.Stop()
	}
	if s.weightTicker != nil {
		s.weightTicker.Stop()
	}
	return nil
}
