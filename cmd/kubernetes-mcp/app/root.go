package app

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/spf13/cobra"
)

const Name = "Kubernetes-mcp"

func NewRootCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "Kubernetes MCP server",
		Long:  `A server that implements Model Capable Protocol (MCP) for Kubernetes operations.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 更新日志级别
			logger.InitializeDefaultLogger(cfg.LogLevel, cfg.LogFormat)
		},
	}

	// 全局标志
	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "Log format (console, json)")

	// 添加子命令
	cmd.AddCommand(NewServerCommand(cfg))
	cmd.AddCommand(NewVersionCommand())

	return cmd
}
