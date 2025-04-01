package biz

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/casbin/casbin/v2"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// CasbinService Casbin业务逻辑层
type CasbinService struct {
	CasbinRepository *dal.CasbinRepository
	Trans            *util.Trans
	Enforcer         *casbin.Enforcer `wire:"-"`
	Mutex            sync.RWMutex     `wire:"-"`
}

// SetEnforcer 设置Casbin执行器
func (s *CasbinService) SetEnforcer(ctx context.Context) error {
	// 执行器存在则直接返回
	if s.Enforcer != nil {
		return nil
	}

	// 初始化执行器
	enforcer, err := s.CasbinRepository.InitCasbinEnforcer(ctx)
	if err != nil {
		return err
	}

	s.Enforcer = enforcer
	logging.Context(ctx).Info("Casbin执行器创建成功")
	return nil
}

// ReloadPolicy 重新加载策略
func (s *CasbinService) ReloadPolicy(ctx context.Context) error {
	// 添加写锁
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Enforcer == nil {
		return nil
	}

	if err := s.Enforcer.LoadPolicy(); err != nil {
		return err
	}

	logging.Context(ctx).Info("Casbin策略已重新加载")
	return nil
}

// GetPolicies 获取策略列表
func (s *CasbinService) GetPolicies(ctx context.Context, req *schema.PolicyListRequest) (*schema.PolicyListResponse, error) {
	var policies [][]string
	var filteredPolicies [][]string

	// 根据类型获取不同的策略
	var err error
	s.Mutex.RLock()
	if req.Type == "" || req.Type == "p" {
		policies, err = s.Enforcer.GetPolicy()
	} else if req.Type == "g" {
		policies, err = s.Enforcer.GetGroupingPolicy()
	}
	s.Mutex.RUnlock()
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("获取Casbin策略失败: %v", err))
	}

	// 过滤策略
	for _, policy := range policies {
		if req.Type == "p" && len(policy) >= 3 {
			// 如果是p类型策略，检查subject、object和action
			if (req.Subject == "" || strings.Contains(policy[0], req.Subject)) &&
				(req.Object == "" || strings.Contains(policy[1], req.Object)) &&
				(req.Action == "" || strings.Contains(policy[2], req.Action)) {
				filteredPolicies = append(filteredPolicies, policy)
			}
		} else if req.Type == "g" && len(policy) >= 2 {
			// 如果是g类型策略，检查user和role
			if (req.Subject == "" || strings.Contains(policy[0], req.Subject)) &&
				(req.Object == "" || strings.Contains(policy[1], req.Object)) {
				filteredPolicies = append(filteredPolicies, policy)
			}
		} else if req.Type == "" {
			// 如果未指定类型，添加所有策略
			filteredPolicies = append(filteredPolicies, policy)
		}
	}

	// 计算分页
	total := len(filteredPolicies)
	start := (req.Page - 1) * req.PageSize
	if start >= total {
		start = 0
	}
	end := start + req.PageSize
	if end > total {
		end = total
	}

	// 构建响应
	items := make([]schema.PolicyListItem, 0, end-start)
	for i := start; i < end && i < len(filteredPolicies); i++ {
		policy := filteredPolicies[i]
		item := schema.PolicyListItem{
			ID:      uint(i + 1), // 使用索引作为ID
			Type:    req.Type,    // 策略类型
			Subject: "",
			Object:  "",
			Action:  "",
		}

		// 根据策略类型设置字段
		if req.Type == "p" && len(policy) >= 3 {
			item.Subject = policy[0]
			item.Object = policy[1]
			item.Action = policy[2]
		} else if req.Type == "g" && len(policy) >= 2 {
			item.Subject = policy[0]
			item.Object = policy[1]
		} else if req.Type == "" && len(policy) >= 3 {
			// 未指定类型，尝试识别
			item.Type = "p"
			item.Subject = policy[0]
			item.Object = policy[1]
			item.Action = policy[2]
		}

		items = append(items, item)
	}

	return &schema.PolicyListResponse{
		Total: total,
		Items: items,
	}, nil
}

// AddPolicy 添加策略
func (s *CasbinService) AddPolicy(ctx context.Context, req *schema.PolicyRequest) error {
	s.Mutex.Lock()

	// 添加权限策略
	added, err := s.Enforcer.AddPolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return errors.WithStack(fmt.Errorf("添加Casbin策略失败: %v", err))
	}

	if !added {
		return errors.Conflict("策略已存在")
	}

	// 保存到数据库
	if err := s.Enforcer.SavePolicy(); err != nil {
		return errors.WithStack(fmt.Errorf("保存Casbin策略到数据库失败: %v", err))
	}

	s.Mutex.Unlock()

	logging.Context(ctx).Info("成功添加策略",
		zap.String("subject", req.Subject),
		zap.String("object", req.Object),
		zap.String("action", req.Action))
	return nil
}

// RemovePolicy 移除策略
func (s *CasbinService) RemovePolicy(ctx context.Context, req *schema.PolicyRequest) error {
	s.Mutex.Lock()

	// 移除权限策略
	removed, err := s.Enforcer.RemovePolicy(req.Subject, req.Object, req.Action)
	if err != nil {
		return errors.WithStack(fmt.Errorf("移除Casbin策略失败: %v", err))
	}

	if !removed {
		return errors.NotFound("策略不存在")
	}

	// 保存到数据库
	if err := s.Enforcer.SavePolicy(); err != nil {
		return errors.WithStack(fmt.Errorf("保存Casbin策略到数据库失败: %v", err))
	}

	s.Mutex.Unlock()

	logging.Context(ctx).Info("成功移除策略",
		zap.String("subject", req.Subject),
		zap.String("object", req.Object),
		zap.String("action", req.Action))
	return nil
}

// AddRoleForUser 为用户添加角色
func (s *CasbinService) AddRoleForUser(ctx context.Context, req *schema.RoleRequest) error {
	s.Mutex.Lock()

	// 添加角色关系
	added, err := s.Enforcer.AddGroupingPolicy(req.User, req.Role)
	if err != nil {
		return errors.WithStack(fmt.Errorf("为用户添加角色失败: %v", err))
	}

	if !added {
		return errors.Conflict("用户角色关系已存在")
	}

	// 保存到数据库
	if err := s.Enforcer.SavePolicy(); err != nil {
		return errors.WithStack(fmt.Errorf("保存Casbin策略到数据库失败: %v", err))
	}

	s.Mutex.Unlock()

	logging.Context(ctx).Info("成功为用户添加角色",
		zap.String("user", req.User),
		zap.String("role", req.Role))
	return nil
}

// RemoveRoleForUser 为用户移除角色
func (s *CasbinService) RemoveRoleForUser(ctx context.Context, req *schema.RoleRequest) error {
	s.Mutex.Lock()

	// 移除角色关系
	removed, err := s.Enforcer.RemoveGroupingPolicy(req.User, req.Role)
	if err != nil {
		return errors.WithStack(fmt.Errorf("为用户移除角色失败: %v", err))
	}

	if !removed {
		return errors.NotFound("用户角色关系不存在")
	}

	// 保存到数据库
	if err := s.Enforcer.SavePolicy(); err != nil {
		return errors.WithStack(fmt.Errorf("保存Casbin策略到数据库失败: %v", err))
	}

	s.Mutex.Unlock()

	logging.Context(ctx).Info("成功为用户移除角色",
		zap.String("user", req.User),
		zap.String("role", req.Role))
	return nil
}

// Enforce 执行权限验证
func (s *CasbinService) Enforce(ctx context.Context, req *schema.EnforcerRequest) (*schema.EnforcerResponse, error) {
	// 执行权限验证
	s.Mutex.RLock()
	allowed, err := s.Enforcer.Enforce(req.Subject, req.Object, req.Action)
	s.Mutex.RUnlock()

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("执行Casbin权限验证失败: %v", err))
	}

	return &schema.EnforcerResponse{
		Allowed: allowed,
	}, nil
}

// GetAllRoles 获取所有角色
func (s *CasbinService) GetAllRoles(ctx context.Context) ([]string, error) {
	s.Mutex.RLock()
	roles, err := s.Enforcer.GetAllRoles()
	s.Mutex.RUnlock()

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("获取Casbin角色失败: %v", err))
	}
	return roles, nil
}

// GetRolesForUser 获取用户的所有角色
func (s *CasbinService) GetRolesForUser(ctx context.Context, user string) ([]string, error) {
	s.Mutex.RLock()
	roles, err := s.Enforcer.GetRolesForUser(user)
	s.Mutex.RUnlock()

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("获取用户角色失败: %v", err))
	}
	return roles, nil
}

// GetUsersForRole 获取具有指定角色的所有用户
func (s *CasbinService) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	s.Mutex.RLock()
	users, err := s.Enforcer.GetUsersForRole(role)
	s.Mutex.RUnlock()

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("获取角色用户失败: %v", err))
	}
	return users, nil
}

// MiddlewareEnforce 用于在中间件中方便地执行权限验证
func (s *CasbinService) MiddlewareEnforce(ctx context.Context, sub, path, method string) (bool, error) {
	// 添加读锁
	s.Mutex.RLock()
	allowed, err := s.Enforcer.Enforce(sub, path, method)
	s.Mutex.RUnlock()

	if err != nil {
		logging.Context(ctx).Error("中间件权限验证出错", zap.Error(err))
		return false, err
	}
	return allowed, nil
}

// ImportPolicyFromFile 从策略文件导入策略到数据库
func (s *CasbinService) ImportPolicyFromFile(ctx context.Context, filePath string) error {
	// 检测文件的后缀
	if filepath.Ext(filePath) != ".csv" {
		return errors.WithStack(fmt.Errorf("文件后缀必须为 .csv"))
	}

	// 解析策略文件
	rules, err := s.parsePolicyFile(ctx, filePath)
	if err != nil {
		return err
	}

	// 将策略保存到数据库
	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		// 清空现有策略表
		if err := s.CasbinRepository.DeleteAll(ctx); err != nil {
			return err
		}

		// 批量插入规则
		for _, rule := range rules {
			if err := s.CasbinRepository.Create(ctx, rule); err != nil {
				return err
			}
		}

		logging.Context(ctx).Info(fmt.Sprintf("成功导入 %d 条策略到数据库", len(rules)))

		return nil
	})
}

// parsePolicyFile 解析策略文件
func (s *CasbinService) parsePolicyFile(ctx context.Context, filePath string) ([]*schema.CasbinRule, error) {
	// 打开策略文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("打开策略文件失败: %v", err))
	}
	defer file.Close()

	// 读取并解析策略文件
	var rules []*schema.CasbinRule
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行或注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析单行策略
		rule, err := s.parsePolicyLine(ctx, lineNum, line)
		if err != nil {
			// 记录错误但继续解析
			logging.Context(ctx).Warn(err.Error())
			continue
		}

		rules = append(rules, rule)
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(fmt.Errorf("读取策略文件失败: %v", err))
	}

	return rules, nil
}

// parsePolicyLine 解析单行策略
func (s *CasbinService) parsePolicyLine(ctx context.Context, lineNum int, line string) (*schema.CasbinRule, error) {
	// 解析策略行
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("策略文件第%d行格式错误: %s", lineNum, line)
	}

	// 创建 CasbinRule 实例
	rule := schema.CasbinRule{
		Ptype: parts[0],
		V0:    parts[1],
		V1:    parts[2],
	}

	// 设置可选字段
	if len(parts) > 3 {
		rule.V2 = parts[3]
	}
	if len(parts) > 4 {
		rule.V3 = parts[4]
	}
	if len(parts) > 5 {
		rule.V4 = parts[5]
	}
	if len(parts) > 6 {
		rule.V5 = parts[6]
	}

	return &rule, nil
}
