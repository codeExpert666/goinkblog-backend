package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"
)

// Run 运行 handler 直到收到终止信号
func Run(ctx context.Context, handler func(ctx context.Context) (func(), error)) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	cleanFn, err := handler(ctx)
	if err != nil {
		return err
	}

EXIT:
	for {
		sig := <-sc
		logging.Context(ctx).Info("接收到信号", zap.String("signal", sig.String()))

		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT: // 程序退出信号，正常退出
			state = 0
			break EXIT
		case syscall.SIGHUP: // 终端连接断开，忽略
		default: // 其他信号，异常退出
			break EXIT
		}
	}

	cleanFn()
	logging.Context(ctx).Info("服务已退出，再见...")
	time.Sleep(time.Millisecond * 100) // 确保所有日志写入
	os.Exit(state)
	return nil
}
