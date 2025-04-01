package main

import (
	"os"

	"github.com/codeExpert666/goinkblog-backend/cmd"
	"github.com/urfave/cli/v2"
)

// Usage: go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "v1.0.0"

// @title GoInk Blog API
// @version v1.0.0
// @description GoInk Blog是一个轻量级但功能完备的智能博客系统，采用前后端分离架构，集成AI功能增强内容创作体验。
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @schemes http https
// @basePath /api/v1
func main() {
	app := cli.NewApp()
	app.Name = "goinkblog"
	app.Version = VERSION
	app.Usage = "GoInk Blog Backend Service"
	app.Commands = []*cli.Command{
		cmd.StartCmd(),
		cmd.StopCmd(),
		cmd.VersionCmd(VERSION),
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
