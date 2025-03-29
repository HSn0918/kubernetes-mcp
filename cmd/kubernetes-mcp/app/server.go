package app

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/hsn0918/kubernetes-mcp/pkg/server"
	"github.com/spf13/cobra"
)

func NewServerCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the MCP server",
		Long:  `Start the Model Capable Protocol (MCP) server for Kubernetes operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.GetLogger()
			log.Info("Starting MCP server", "transport", cfg.Transport, "port", cfg.Port)

			// 创建处理程序提供者
			handlerProvider := handlers.NewHandlerProvider()

			// 创建服务器
			serverFactory := server.NewServerFactory(handlerProvider)
			server, err := serverFactory.CreateServer(cfg)
			if err != nil {
				return err
			}

			// 启动服务器
			return server.Start()
		},
	}

	// 服务器标志
	cmd.Flags().StringVar(&cfg.Transport, "transport", cfg.Transport, "Transport type (stdio or sse)")
	cmd.Flags().IntVar(&cfg.Port, "port", cfg.Port, "Port to use for SSE transport")
	cmd.Flags().StringVar(&cfg.Kubeconfig, "kubeconfig", cfg.Kubeconfig, "Path to kubeconfig file")

	return cmd
}
