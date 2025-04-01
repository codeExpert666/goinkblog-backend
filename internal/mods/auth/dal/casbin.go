package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

func GetCasbinDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.CasbinRule{})
}

// CasbinRepository Casbin数据访问层
type CasbinRepository struct {
	DB *gorm.DB
}

// InitCasbinEnforcer 初始化Casbin执行器
func (r *CasbinRepository) InitCasbinEnforcer(ctx context.Context) (*casbin.Enforcer, error) {
	// 获取模型配置文件路径
	modelPath := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.ModelFile)

	// 创建GORM适配器
	tableName := new(schema.CasbinRule).TableName()
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(
		r.DB,
		&schema.CasbinRule{},
		tableName,
	)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("创建Casbin GORM适配器失败: %v", err))
	}

	// 从文件加载模型
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("从文件加载Casbin模型失败: %v", err))
	}

	// 创建执行器
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("创建Casbin执行器失败: %v", err))
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, errors.WithStack(fmt.Errorf("加载Casbin策略失败: %v", err))
	}

	logging.Context(ctx).Info("Casbin执行器初始化成功")
	return enforcer, nil
}

func (r *CasbinRepository) Create(ctx context.Context, rule *schema.CasbinRule) error {
	result := GetCasbinDB(ctx, r.DB).Create(rule)
	return errors.WithStack(result.Error)
}

func (r *CasbinRepository) DeleteAll(ctx context.Context) error {
	result := GetCasbinDB(ctx, r.DB).Where("1=1").Delete(&schema.CasbinRule{})
	return errors.WithStack(result.Error)
}
