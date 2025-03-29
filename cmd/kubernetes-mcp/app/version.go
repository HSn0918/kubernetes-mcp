package app

import (
	"fmt"

	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/spf13/cobra"
)

// 版本信息
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetLogger()
			versionInfo := fmt.Sprintf("Kubernetes-mcp version %s (commit: %s, build date: %s)\n",
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
