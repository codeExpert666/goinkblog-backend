package cmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/urfave/cli/v2"
)

// StopCmd 定义停止服务的命令
func StopCmd() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop server",
		Action: func(c *cli.Context) error {
			lockFile := fmt.Sprintf("%s.lock", c.App.Name)
			pidBytes, err := os.ReadFile(lockFile)
			if err != nil {
				fmt.Printf("读取 lock 文件失败: %s, 错误为: %s\n", lockFile, err.Error())
				return err
			}

			pid, err := strconv.Atoi(string(pidBytes))
			if err != nil {
				fmt.Printf("从 lock 文件解析进程 ID 失败: %s, 错误为: %s\n", lockFile, err.Error())
				return err
			}

			// 查找进程
			process, err := os.FindProcess(pid)
			if err != nil {
				fmt.Printf("无法找到 ID 为 %d 的进程, 错误为: %s\n", pid, err.Error())
				return err
			}

			// 发送 SIGTERM 信号给进程以优雅退出
			err = process.Signal(syscall.SIGTERM)
			if err != nil {
				fmt.Printf("无法终止 ID 为 %d 的进程, 错误为: %s\n", pid, err.Error())
				return err
			}

			err = os.Remove(lockFile)
			if err != nil {
				fmt.Printf("无法删除 lock 文件: %s, 错误为: %s\n", lockFile, err.Error())
				return err
			}

			fmt.Printf("服务 %s 停止成功\n", c.App.Name)
			return nil
		},
	}
}
