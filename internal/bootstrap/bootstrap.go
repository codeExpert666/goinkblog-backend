package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"strings"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/wirex"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"go.uber.org/zap"
)

// RunConfig 运行配置
type RunConfig struct {
	WorkDir   string // 工作目录
	Configs   string // 配置文件或目录(逗号分隔)
	StaticDir string // 静态文件目录
}

// 运行服务
func Run(ctx context.Context, runCfg RunConfig) error {
	defer func() {
		if err := zap.L().Sync(); err != nil {
			fmt.Printf("failed to sync zap logger: %s \n", err.Error())
		}
	}()

	// 加载配置
	workDir := runCfg.WorkDir
	staticDir := runCfg.StaticDir
	config.MustLoad(workDir, strings.Split(runCfg.Configs, ",")...)
	config.C.General.WorkDir = workDir
	config.C.Middleware.Static.Dir = staticDir
	config.C.PreLoad()
	config.C.Print()

	// 初始化日志
	cleanLoggerFn, err := logging.InitWithConfig(ctx, &config.C.Logger, initLoggerHook)
	if err != nil {
		return err
	}
	ctx = logging.NewTag(ctx, logging.TagKeyMain)

	logging.Context(ctx).Info("正在启动服务 ...",
		zap.String("version", config.C.General.Version),
		zap.Int("pid", os.Getpid()),
		zap.String("workdir", workDir),
		zap.String("config", runCfg.Configs),
		zap.String("static", staticDir),
	)

	// 启动pprof服务
	if addr := config.C.General.PprofAddr; addr != "" {
		logging.Context(ctx).Info("pprof 服务已启动，监听端口： " + addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				logging.Context(ctx).Error("pprof 服务启动失败", zap.Error(err))
			}
		}()
	}

	// 构建依赖注入器
	injector, cleanInjectorFn, err := wirex.BuildInjector(ctx)
	if err != nil {
		return err
	}

	if err := injector.M.Init(ctx); err != nil {
		return err
	}

	return util.Run(ctx, func(ctx context.Context) (func(), error) {
		cleanHTTPServerFn, err := startHTTPServer(ctx, injector)
		if err != nil {
			return cleanInjectorFn, err
		}

		return func() {
			if err := injector.M.Release(ctx); err != nil {
				logging.Context(ctx).Error("释放依赖注入器失败", zap.Error(err))
			}

			if cleanHTTPServerFn != nil {
				cleanHTTPServerFn()
			}
			if cleanInjectorFn != nil {
				cleanInjectorFn()
			}
			if cleanLoggerFn != nil {
				cleanLoggerFn()
			}
		}, nil
	})
}
