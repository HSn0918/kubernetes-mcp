package app

import (
	"github.com/spf13/cobra"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers"
	"github.com/hsn0918/kubernetes-mcp/pkg/health"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/hsn0918/kubernetes-mcp/pkg/server"
)

func NewServerCommand(cfg *config.Config) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the MCP server",
		Long:  `Start the Model Capable Protocol (MCP) server for Kubernetes operations.`,
	}

	// 添加共享标志到父命令
	serverCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")
	serverCmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "Log format (console, json)")
	serverCmd.PersistentFlags().StringVar(&cfg.Kubeconfig, "kubeconfig", cfg.Kubeconfig, "Path to kubeconfig file")

	// 创建传输子命令
	transportCmd := &cobra.Command{
		Use:   "transport",
		Short: "Set transport type for MCP server",
		Long:  `Set the transport mechanism (stdio, sse, or streamable) for the MCP server.`,
	}

	// 添加SSE传输子命令
	sseCmd := &cobra.Command{
		Use:   "sse",
		Short: "Use Server-Sent Events (SSE) transport",
		Long:  `Use Server-Sent Events (SSE) as the transport mechanism for the MCP server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Transport = "sse"
			log := logger.GetLogger()
			health.StartHealthServer(cfg.HealthPort, log)

			log.Info("Starting MCP server", "transport", cfg.Transport, "port", cfg.Port)
			// 创建处理程序提供者
			handlerProvider := handlers.NewHandlerProvider()

			// 创建服务器
			serverFactory := server.NewServerFactory(handlerProvider)
			server, err := serverFactory.CreateServer(cfg)
			if err != nil {
				return err
			}
			health.SetReady()
			err = server.Start()
			if err != nil {
				health.SetNotReady()
				return err
			}
			// 启动服务器
			return nil
		},
	}

	// 添加StreamableHTTP传输子命令
	streamableCmd := &cobra.Command{
		Use:   "streamable",
		Short: "Use StreamableHTTP transport (supports streaming)",
		Long:  `Use StreamableHTTP as the transport mechanism for the MCP server. This mode supports streaming operations and progress notifications.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Transport = "streamable"
			log := logger.GetLogger()
			health.StartHealthServer(cfg.HealthPort, log)

			log.Info("Starting MCP server", "transport", cfg.Transport, "port", cfg.Port)
			// 创建处理程序提供者
			handlerProvider := handlers.NewHandlerProvider()

			// 创建服务器
			serverFactory := server.NewServerFactory(handlerProvider)
			server, err := serverFactory.CreateServer(cfg)
			if err != nil {
				return err
			}
			health.SetReady()
			err = server.Start()
			if err != nil {
				health.SetNotReady()
				return err
			}
			// 启动服务器
			return nil
		},
	}

	// 添加stdio传输子命令
	stdioCmd := &cobra.Command{
		Use:   "stdio",
		Short: "Use standard input/output transport",
		Long:  `Use standard input/output as the transport mechanism for the MCP server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Transport = "stdio"
			log := logger.GetLogger()

			log.Info("Starting MCP server", "transport", cfg.Transport)
			// 创建处理程序提供者
			handlerProvider := handlers.NewHandlerProvider()

			// 创建服务器
			serverFactory := server.NewServerFactory(handlerProvider)
			server, err := serverFactory.CreateServer(cfg)
			if err != nil {
				return err
			}
			err = server.Start()
			if err != nil {
				return err
			}
			// 启动服务器
			return nil
		},
	}

	// 为SSE子命令添加特定的标志
	sseCmd.Flags().IntVar(&cfg.Port, "port", cfg.Port, "Port to use for SSE transport")
	sseCmd.Flags().IntVar(&cfg.HealthPort, "health-port", cfg.HealthPort, "Port for health check endpoints (/healthz, /readyz)")
	sseCmd.Flags().StringVar(&cfg.AllowOrigins, "allow-origins", cfg.AllowOrigins, "Cross-Origin Resource Sharing allowed origins, comma separated or * for all")
	sseCmd.Flags().StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "Base URL for SSE server (e.g. http://example.com:8080), defaults to http://localhost:<port>")

	// 为StreamableHTTP子命令添加特定的标志
	streamableCmd.Flags().IntVar(&cfg.Port, "port", cfg.Port, "Port to use for StreamableHTTP transport")
	streamableCmd.Flags().IntVar(&cfg.HealthPort, "health-port", cfg.HealthPort, "Port for health check endpoints (/healthz, /readyz)")
	streamableCmd.Flags().StringVar(&cfg.AllowOrigins, "allow-origins", cfg.AllowOrigins, "Cross-Origin Resource Sharing allowed origins, comma separated or * for all")

	// 添加子命令到传输命令
	transportCmd.AddCommand(sseCmd)
	transportCmd.AddCommand(streamableCmd)
	transportCmd.AddCommand(stdioCmd)

	// 添加传输命令到服务器命令
	serverCmd.AddCommand(transportCmd)

	return serverCmd
}
