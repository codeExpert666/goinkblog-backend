package biz

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
)

// CasbinService Casbin业务逻辑层
type CasbinService struct {
	CasbinRepository *dal.CasbinRepository
	Casbinx          Casbinx
}

// ListPolicies 获取策略列表
func (s *CasbinService) ListPolicies(ctx context.Context, req *schema.ListPolicyRequest) (*schema.ListPolicyResponse, error) {
	return s.CasbinRepository.List(ctx, req)
}

// AddPolicy 添加策略
func (s *CasbinService) AddPolicy(ctx context.Context, req *schema.CreatePolicyRequest) (*schema.PolicyItem, error) {
	// 检查策略是否已经存在
	policy := &schema.CasbinRule{
		Ptype: req.Type,
		V0:    req.Subject,
		V1:    req.Object,
		V2:    req.Action,
	}

	if exist, err := s.CasbinRepository.Exist(ctx, policy); err != nil {
		return nil, err
	} else if exist {
		return nil, errors.Conflict("策略已存在")
	}

	// 创建策略
	if err := s.CasbinRepository.Create(ctx, policy); err != nil {
		return nil, err
	}

	// 不再重新查询，因为 ID 会回写
	return &schema.PolicyItem{
		ID:      policy.ID,
		Type:    policy.Ptype,
		Subject: policy.V0,
		Object:  policy.V1,
		Action:  policy.V2,
	}, nil
}

// DeletePolicy 移除策略
func (s *CasbinService) DeletePolicy(ctx context.Context, id uint) error {
	// 检查策略是否存在
	if exist, err := s.CasbinRepository.Exist(ctx, &schema.CasbinRule{ID: id}); err != nil {
		return err
	} else if !exist {
		return errors.NotFound("策略不存在")
	}

	// 删除策略
	return s.CasbinRepository.Delete(ctx, id)
}

// Enforce 执行权限验证
func (s *CasbinService) Enforce(ctx context.Context, req *schema.EnforcerRequest) (*schema.EnforcerResponse, error) {
	enforcer := s.Casbinx.GetEnforcer()

	allow, err := enforcer.Enforce(req.Subject, req.Object, req.Action)
	if err != nil {
		return nil, err
	}

	return &schema.EnforcerResponse{Allowed: allow}, nil
}
