package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// VersionCmd 定义版本命令
func VersionCmd(version string) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version",
		Action: func(c *cli.Context) error {
			fmt.Println(version)
			return nil
		},
	}
}
