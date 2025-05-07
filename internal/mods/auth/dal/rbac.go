package dal

import (
	"context"
	"fmt"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetCasbinDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.CasbinRule{})
}

// CasbinRepository Casbin数据访问层
type CasbinRepository struct {
	Cache cachex.Cacher
	DB    *gorm.DB
}

func (r *CasbinRepository) Count(ctx context.Context, cond *schema.CasbinRule) (int64, error) {
	var count int64
	if err := GetCasbinDB(ctx, r.DB).Where(cond).Count(&count).Error; err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

func (r *CasbinRepository) List(ctx context.Context, req *schema.ListPolicyRequest) (*schema.ListPolicyResponse, error) {
	var resp schema.ListPolicyResponse

	db := GetCasbinDB(ctx, r.DB)

	// 应用过滤条件
	if req.Type != "" {
		db = db.Where("ptype = ?", req.Type)
	}
	if req.Subject != "" {
		db = db.Where("v0 LIKE ?", "%"+req.Subject+"%")
	}
	if req.Object != "" {
		db = db.Where("v1 LIKE ?", "%"+req.Object+"%")
	}
	if req.Type != "g" && req.Action != "" { // g 策略的 Action 字段无效
		db = db.Where("v2 = ?", req.Action)
	}
	// 特殊条件，策略（p, admin, /api/auth/rbac/*, *）不允许查看，以防误删导致权限管理接口无法访问
	db = db.Where("v1 NOT LIKE ?", "/api/auth/rbac/%")

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 应用分页
	offset := (req.Page - 1) * req.PageSize
	db = db.Offset(offset).Limit(req.PageSize)

	// 查询数据
	var rules []*schema.CasbinRule
	if err := db.Find(&rules).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// 将查询结果转换为响应格式
	var items []*schema.PolicyItem
	for _, rule := range rules {
		item := &schema.PolicyItem{
			ID:      rule.ID,
			Type:    rule.Ptype,
			Subject: rule.V0,
			Object:  rule.V1,
		}
		// p 策略有 Action 字段
		if rule.Ptype == "p" {
			item.Action = rule.V2
		}

		items = append(items, item)
	}

	// 填充响应
	resp.Items = items
	resp.Total = total
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.TotalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &resp, nil
}

func (r *CasbinRepository) Get(ctx context.Context, id uint) (*schema.CasbinRule, error) {
	var rule *schema.CasbinRule
	if err := GetCasbinDB(ctx, r.DB).Where("id = ?", id).First(rule).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return rule, nil
}

func (r *CasbinRepository) Exist(ctx context.Context, cond *schema.CasbinRule) (bool, error) {
	var count int64
	if err := GetCasbinDB(ctx, r.DB).Where(cond).Count(&count).Error; err != nil {
		return false, errors.WithStack(err)
	}
	return count > 0, nil
}

func (r *CasbinRepository) Create(ctx context.Context, rule *schema.CasbinRule) error {
	if err := GetCasbinDB(ctx, r.DB).Create(rule).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if err := r.syncToCasbin(ctx); err != nil {
		logging.Context(ctx).Error("向缓存中存入Casbin同步标记失败", zap.Error(err))
	}

	return nil
}

func (r *CasbinRepository) Delete(ctx context.Context, id uint) error {
	if err := GetCasbinDB(ctx, r.DB).Where("id = ?", id).Delete(&schema.CasbinRule{}).Error; err != nil {
		return errors.WithStack(err)
	}

	// 向缓存中存入同步标记
	if err := r.syncToCasbin(ctx); err != nil {
		logging.Context(ctx).Error("向缓存中存入Casbin同步标记失败", zap.Error(err))
	}

	return nil
}

// syncToCasbin 同步策略变更到Casbin
func (r *CasbinRepository) syncToCasbin(ctx context.Context) error {
	return r.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin,
		fmt.Sprintf("%d", time.Now().Unix()))
}
