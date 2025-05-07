package biz

import (
	"bufio"
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Casbinx struct {
	adapter          *gormadapter.Adapter `wire:"-"` // GORM 适配器
	model            model.Model          `wire:"-"` // Casbin 模型
	enforcer         *atomic.Value        `wire:"-"` // Casbin 执行器
	ticker           *time.Ticker         `wire:"-"` // 定时自动加载策略
	Cache            cachex.Cacher        // 策略同步通知
	CasbinRepository *dal.CasbinRepository
}

// GetEnforcer 获取当前的 casbin 执行器实例
func (c *Casbinx) GetEnforcer() *casbin.Enforcer {
	if v := c.enforcer.Load(); v != nil {
		return v.(*casbin.Enforcer)
	}
	return nil
}

// Init 初始化Casbin模块
func (c *Casbinx) Init(ctx context.Context) error {
	// 将策略配置导入到数据库
	if path := config.C.Middleware.Casbin.PolicyFile; path != "" {
		if count, err := c.CasbinRepository.Count(ctx, &schema.CasbinRule{}); err != nil {
			return err
		} else if count == 0 { // 只有当数据库中不存在策略时才进行导入
			policyFile := filepath.Join(config.C.General.WorkDir, path)
			err = c.importPolicyFromFile(ctx, policyFile)
			if err != nil {
				logging.Context(ctx).Error("导入 Casbin 策略到数据库失败", zap.Error(err))
				return err
			}
		} else {
			logging.Context(ctx).Info("数据库中已存在 Casbin 策略，跳过导入")
		}
	}

	// 初始化GORM适配器
	gormadapter.TurnOffAutoMigrate(c.CasbinRepository.DB) // 不允许适配器自动创建表
	tableName := new(schema.CasbinRule).TableName()
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(
		c.CasbinRepository.DB,
		&schema.CasbinRule{},
		tableName,
	)
	if err != nil {
		logging.Context(ctx).Error("创建Casbin GORM适配器失败", zap.Error(err))
		return err
	}
	c.adapter = adapter

	// 从文件加载模型
	modelPath := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.ModelFile)
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		logging.Context(ctx).Error("从文件加载Casbin模型失败", zap.Error(err))
		return err
	}
	c.model = m

	// 初始化 Casbin 执行器
	c.enforcer = new(atomic.Value)
	if err := c.load(ctx); err != nil {
		return err
	}

	// 后台定时检查策略更新并重新加载策略
	go c.autoLoad(ctx)
	return nil
}

func (c *Casbinx) load(ctx context.Context) error {
	// 创建执行器
	e, err := casbin.NewEnforcer(c.model, c.adapter)
	if err != nil {
		logging.Context(ctx).Error("创建Casbin执行器失败", zap.Error(err))
		return err
	}

	// 加载策略
	if err := e.LoadPolicy(); err != nil {
		logging.Context(ctx).Error("加载Casbin策略失败", zap.Error(err))
		return err
	}

	c.enforcer.Store(e)
	logging.Context(ctx).Info("Casbin执行器设置成功")
	return nil
}

func (c *Casbinx) autoLoad(ctx context.Context) {
	// 记录上次更新时间
	var lastUpdated int64

	c.ticker = time.NewTicker(time.Duration(config.C.Middleware.Casbin.AutoLoadInterval) * time.Second)
	for range c.ticker.C {
		// 从缓存中获取同步标记
		val, ok, err := c.Cache.Get(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin)
		if err != nil {
			logging.Context(ctx).Error("从缓存中获取 Casbin 同步标记失败", zap.Error(err), zap.String("key", config.CacheKeyForSyncToCasbin))
			continue
		} else if !ok {
			continue
		}

		// 解析更新时间
		updated, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logging.Context(ctx).Error("解析 Casbin 同步标记缓存值失败", zap.Error(err), zap.String("val", val))
			continue
		}

		// 如果有更新，重新加载策略
		if lastUpdated < updated {
			if err := c.load(ctx); err != nil {
				logging.Context(ctx).Error("更新 Casbin 执行器失败", zap.Error(err))
			} else {
				lastUpdated = updated
			}
		}
	}
}

func (c *Casbinx) Release(ctx context.Context) error {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	return nil
}

// importPolicyFromFile 从策略文件导入策略到数据库
func (c *Casbinx) importPolicyFromFile(ctx context.Context, filePath string) error {
	// 检测文件的后缀
	if filepath.Ext(filePath) != ".csv" {
		return errors.Errorf("文件后缀必须为 .csv")
	}

	// 解析策略文件
	rules, err := c.parsePolicyFile(ctx, filePath)
	if err != nil {
		return err
	}

	// 将策略保存到数据库
	var successCount, failureCount int
	for _, rule := range rules {
		// 创建策略
		if err := c.CasbinRepository.Create(ctx, rule); err != nil {
			logging.Context(ctx).Warn("创建策略出错", zap.Error(err), zap.Any("rule", rule))
			failureCount++
			continue
		}
		successCount++
	}

	logging.Context(ctx).Info("导入 Casbin 策略到数据库完成",
		zap.Int("total_count", len(rules)),
		zap.Int("success_count", successCount),
		zap.Int("failure_count", failureCount),
	)
	return nil
}

// parsePolicyFile 解析策略文件
func (c *Casbinx) parsePolicyFile(ctx context.Context, filePath string) ([]*schema.CasbinRule, error) {
	// 打开策略文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf("打开策略文件失败: %v", err)
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
		rule, err := c.parsePolicyLine(ctx, lineNum, line)
		if err != nil {
			// 记录错误但继续解析
			logging.Context(ctx).Warn(err.Error())
			continue
		}

		rules = append(rules, rule)
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return nil, errors.Errorf("读取策略文件失败: %v", err)
	}

	return rules, nil
}

// parsePolicyLine 解析单行策略
func (c *Casbinx) parsePolicyLine(ctx context.Context, lineNum int, line string) (*schema.CasbinRule, error) {
	// 解析策略行
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) < 3 {
		return nil, errors.Errorf("策略文件第%d行格式错误: %s", lineNum, line)
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
