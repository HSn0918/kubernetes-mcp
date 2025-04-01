package app

import (
	"fmt"

	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/spf13/cobra"
)

// 版本信息
var (
	Version   = "0.1.0"
	Commit    = "none"
	BuildDate = "unknown"
)

// Logo ASCII 艺术字
const logo = `
 ██ ▄█▀ █    ██  ▄▄▄▄   ▓█████  ███▄ ▄███▓ ▄████▄   ██▓███
 ██▄█▒  ██  ▓██▒▓█████▄ ▓█   ▀ ▓██▒▀█▀ ██▒▒██▀ ▀█  ▓██░  ██▒
▓███▄░ ▓██  ▒██░▒██▒ ▄██▒███   ▓██    ▓██░▒▓█    ▄ ▓██░ ██▓▒
▓██ █▄ ▓▓█  ░██░▒██░█▀  ▒▓█  ▄ ▒██    ▒██ ▒▓▓▄ ▄██▒▒██▄█▓▒ ▒
▒██▒ █▄▒▒█████▓ ░▓█  ▀█▓░▒████▒▒██▒   ░██▒▒ ▓███▀ ░▒██▒ ░  ░
▒ ▒▒ ▓▒░▒▓▒ ▒ ▒ ░▒▓███▀▒░░ ▒░ ░░ ▒░   ░  ░░ ░▒ ▒  ░▒▓▒░ ░  ░
░ ░▒ ▒░░░▒░ ░ ░ ▒░▒   ░  ░ ░  ░░  ░      ░  ░  ▒   ░▒ ░
░ ░░ ░  ░░░ ░ ░  ░    ░    ░   ░      ░   ░        ░░
░  ░      ░      ░         ░  ░       ░   ░ ░
                     ░                     ░
`

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetLogger()

			// 先打印 logo
			fmt.Print(logo)

			// 打印版本信息
			versionInfo := fmt.Sprintf("Kubernetes-MCP version %s (commit: %s, build date: %s)\n",
				Version, Commit, BuildDate)

			fmt.Print(versionInfo)

			log.Info("Version info displayed",
				"version", Version,
				"commit", Commit,
				"buildDate", BuildDate,
			)
		},
	}

	return cmd
}
