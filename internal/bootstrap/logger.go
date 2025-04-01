package bootstrap

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/gormx"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/spf13/cast"
)

// 初始化日志钩子
func initLoggerHook(_ context.Context, cfg *logging.HookConfig) (*logging.Hook, error) {
	extra := cfg.Extra
	if extra == nil {
		extra = make(map[string]string)
	}
	extra["appname"] = config.C.General.AppName

	db, err := gormx.New(gormx.Config{
		Debug:        cast.ToBool(cfg.Options["debug"]),
		DSN:          cast.ToString(cfg.Options["dsn"]),
		MaxLifetime:  cast.ToInt(cfg.Options["max_life_time"]),
		MaxIdleTime:  cast.ToInt(cfg.Options["max_idle_time"]),
		MaxOpenConns: cast.ToInt(cfg.Options["max_open_conns"]),
		MaxIdleConns: cast.ToInt(cfg.Options["max_idle_conns"]),
		TablePrefix:  config.C.Storage.DB.TablePrefix,
	})
	if err != nil {
		return nil, err
	}

	hook := logging.NewHook(logging.NewGormHook(db),
		logging.SetHookExtra(cfg.Extra),
		logging.SetHookMaxJobs(cfg.MaxBuffer),
		logging.SetHookMaxWorkers(cfg.MaxThread))
	return hook, nil
}
