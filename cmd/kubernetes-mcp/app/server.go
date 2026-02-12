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

	serverCmd.PersistentFlags().StringVar(&cfg.Kubeconfig, "kubeconfig", cfg.Kubeconfig, "Path to kubeconfig file")

	transportCmd := &cobra.Command{
		Use:   "transport",
		Short: "Set transport type for MCP server",
		Long:  `Set the transport mechanism (stdio, sse, or streamable) for the MCP server.`,
	}

	sseCmd := newTransportCommand(
		cfg,
		config.TransportSSE,
		"Use Server-Sent Events (SSE) transport",
		`Use Server-Sent Events (SSE) as the transport mechanism for the MCP server.`,
		true,
	)
	streamableCmd := newTransportCommand(
		cfg,
		config.TransportStreamable,
		"Use StreamableHTTP transport (supports streaming)",
		`Use StreamableHTTP as the transport mechanism for the MCP server. This mode supports streaming operations and progress notifications.`,
		true,
	)
	stdioCmd := newTransportCommand(
		cfg,
		config.TransportStdio,
		"Use standard input/output transport",
		`Use standard input/output as the transport mechanism for the MCP server.`,
		false,
	)

	bindNetworkFlags(sseCmd, cfg, true)
	bindNetworkFlags(streamableCmd, cfg, false)

	transportCmd.AddCommand(sseCmd)
	transportCmd.AddCommand(streamableCmd)
	transportCmd.AddCommand(stdioCmd)

	serverCmd.AddCommand(transportCmd)

	return serverCmd
}

func newTransportCommand(cfg *config.Config, transport, short, long string, enableHealth bool) *cobra.Command {
	return &cobra.Command{
		Use:   transport,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Transport = transport
			return startMCPServer(cfg, enableHealth)
		},
	}
}

func bindNetworkFlags(cmd *cobra.Command, cfg *config.Config, includeBaseURL bool) {
	cmd.Flags().IntVar(&cfg.Port, "port", cfg.Port, "Port to use for network transport")
	cmd.Flags().IntVar(&cfg.HealthPort, "health-port", cfg.HealthPort, "Port for health check endpoints (/healthz, /readyz)")
	cmd.Flags().StringVar(&cfg.AllowOrigins, "allow-origins", cfg.AllowOrigins, "Cross-Origin Resource Sharing allowed origins, comma separated or * for all")
	if includeBaseURL {
		cmd.Flags().StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "Base URL for SSE server (e.g. http://example.com:8080), defaults to http://localhost:<port>")
	}
}

func startMCPServer(cfg *config.Config, enableHealth bool) error {
	log := logger.GetLogger()
	if enableHealth {
		health.StartHealthServer(cfg.HealthPort, log)
		log.Info(
			"Starting MCP server",
			logger.String("transport", cfg.Transport),
			logger.Int("port", cfg.Port),
		)
	} else {
		log.Info("Starting MCP server", logger.String("transport", cfg.Transport))
	}

	handlerProvider := handlers.NewHandlerProvider()
	serverFactory := server.NewServerFactory(handlerProvider)
	mcpServer, err := serverFactory.CreateServer(cfg)
	if err != nil {
		return err
	}

	if enableHealth {
		health.SetReady()
	}
	if err := mcpServer.Start(); err != nil {
		if enableHealth {
			health.SetNotReady()
		}
		return err
	}
	return nil
}
