package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codeExpert666/goinkblog-backend/internal/bootstrap"
	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/urfave/cli/v2"
)

// StartCmd 定义启动服务的命令
func StartCmd() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "workdir",
				Aliases:     []string{"d"},
				Usage:       "Working directory",
				DefaultText: "configs",
				Value:       "configs",
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Runtime configuration files or directory (relative to workdir, multiple separated by commas)",
				DefaultText: "dev",
				Value:       "dev",
			},
			&cli.StringFlag{
				Name:    "static",
				Aliases: []string{"s"},
				Usage:   "Static files directory",
				DefaultText: "static",
				Value:       "static",
			},
			&cli.BoolFlag{
				Name:  "daemon",
				Usage: "Run as a daemon",
			},
		},
		Action: func(c *cli.Context) error {
			workDir := c.String("workdir")
			staticDir := c.String("static")
			configs := c.String("config")

			if c.Bool("daemon") {
				bin, err := filepath.Abs(os.Args[0])
				if err != nil {
					fmt.Printf("为命令获取绝对路径失败: %s \n", err.Error())
					return err
				}

				args := []string{"start"}
				args = append(args, "-d", workDir)
				args = append(args, "-c", configs)
				args = append(args, "-s", staticDir)
				fmt.Printf("执行命令: %s %s \n", bin, strings.Join(args, " "))
				command := exec.Command(bin, args...)

				// 将标准输出和标准错误输出重定向到日志文件
				stdLogFile := fmt.Sprintf("%s.log", c.App.Name)
				file, err := os.OpenFile(stdLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					fmt.Printf("打开日志文件失败: %s \n", err.Error())
					return err
				}
				defer file.Close()

				command.Stdout = file
				command.Stderr = file

				// 不等待命令执行完成
				err = command.Start()
				if err != nil {
					fmt.Printf("启动守护进程失败: %s \n", err.Error())
					return err
				}

				pid := command.Process.Pid
				_ = os.WriteFile(fmt.Sprintf("%s.lock", c.App.Name), []byte(fmt.Sprintf("%d", pid)), 0666)
				fmt.Printf("服务 %s 的守护进程已启动，进程 ID 为: %d \n", config.C.General.AppName, pid)
				os.Exit(0)
			}

			err := bootstrap.Run(context.Background(), bootstrap.RunConfig{
				WorkDir:   workDir,
				Configs:   configs,
				StaticDir: staticDir,
			})
			if err != nil {
				panic(err)
			}
			return nil
		},
	}
}
