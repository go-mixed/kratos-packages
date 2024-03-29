package command

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func NewRootCmd() ICmder {
	exePath, _ := os.Executable()
	cmd := &cobra.Command{
		Use:           filepath.Base(exePath),
		SilenceErrors: true, // 关闭错误输出
		SilenceUsage:  true, // 关闭错误输出
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// 关闭错误输出
	cmd.PersistentFlags().StringP("conf", "c", "../../configs", "config file path")
	cmd.PersistentFlags().String("conf-driver", "file", "config driver")

	_ = cmd.Execute()
	// 开启错误输出
	cmd.SilenceErrors = false
	cmd.SilenceUsage = false

	cmd.Run = func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	}
	return &BaseCmd{cmd}
}
