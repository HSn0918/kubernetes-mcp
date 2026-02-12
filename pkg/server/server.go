package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/hsn0918/kubernetes-mcp/pkg/middlewares"
)

const (
	serverName            = "Kubernetes-mcp"
	serverVersion         = "1.8.0"
	streamableEndpoint    = "/mcp"
	defaultTransportAlias = "http"
)

type stdioServer struct {
	mcpServer *server.MCPServer
	log       logger.Logger
}

type sseServer struct {
	mcpServer    *server.MCPServer
	sseServer    *server.SSEServer
	port         int
	log          logger.Logger
	allowOrigins string
}

type streamableHTTPServer struct {
	mcpServer            *server.MCPServer
	streamableHTTPServer *server.StreamableHTTPServer
	port                 int
	log                  logger.Logger
	allowOrigins         string
}

type serverFactoryImpl struct {
	handlerProvider interfaces.HandlerProvider
}

var _ MCPServer = &stdioServer{}
var _ MCPServer = &sseServer{}
var _ MCPServer = &streamableHTTPServer{}
var _ Factory = &serverFactoryImpl{}

func (s *stdioServer) GetServer() *server.MCPServer {
	return s.mcpServer
}

func (s *stdioServer) Start() error {
	s.log.Info("Starting stdio server")
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *stdioServer) Stop() error {
	s.log.Info("Stopping stdio server")
	return nil
}

func (s *sseServer) GetServer() *server.MCPServer {
	return s.mcpServer
}

func (s *sseServer) Start() error {
	s.log.Info(
		"Starting SSE server",
		logger.Int("port", s.port),
		logger.String("allowOrigins", s.allowOrigins),
	)
	return s.sseServer.Start(addrFromPort(s.port))
}

func (s *sseServer) Stop() error {
	s.log.Info("Stopping SSE server")
	return nil
}

func (s *streamableHTTPServer) GetServer() *server.MCPServer {
	return s.mcpServer
}

func (s *streamableHTTPServer) Start() error {
	s.log.Info(
		"Starting StreamableHTTP server",
		logger.Int("port", s.port),
		logger.String("allowOrigins", s.allowOrigins),
	)
	return s.streamableHTTPServer.Start(addrFromPort(s.port))
}

func (s *streamableHTTPServer) Stop() error {
	s.log.Info("Stopping StreamableHTTP server")
	return nil
}

func (f *serverFactoryImpl) CreateServer(cfg *config.Config) (MCPServer, error) {
	log := logger.GetLogger()
	mcpServer := newMCPServer(log)
	f.handlerProvider.RegisterAllHandlers(mcpServer)

	switch cfg.Transport {
	case config.TransportSSE:
		return newSSETransportServer(cfg, mcpServer, log), nil
	case config.TransportStreamable, defaultTransportAlias:
		return newStreamableTransportServer(cfg, mcpServer, log), nil
	default:
		return &stdioServer{mcpServer: mcpServer, log: log}, nil
	}
}

func newMCPServer(log logger.Logger) *server.MCPServer {
	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Debug(
			"Request received",
			logger.Any("id", id),
			logger.String("method", string(method)),
			logger.Any("message", message),
		)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		log.Info(
			"Request successful",
			logger.Any("id", id),
			logger.String("method", string(method)),
		)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Error(
			"Request failed",
			logger.Any("id", id),
			logger.String("method", string(method)),
			logger.Any("error", err),
		)
	})

	options := []server.ServerOption{
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
	}

	return server.NewMCPServer(serverName, serverVersion, options...)
}

func newSSETransportServer(cfg *config.Config, mcpServer *server.MCPServer, log logger.Logger) MCPServer {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:" + strconv.Itoa(cfg.Port)
		log.Info("BaseURL not set, using default", logger.String("baseURL", baseURL))
	} else {
		log.Info("Using configured BaseURL", logger.String("baseURL", baseURL))
	}

	httpServer := &http.Server{}
	mcpSSEServer := server.NewSSEServer(mcpServer,
		server.WithBaseURL(baseURL),
		server.WithHTTPServer(httpServer),
	)
	httpServer.Handler = middlewares.CorsMiddleware(cfg.AllowOrigins, mcpSSEServer)

	return &sseServer{
		mcpServer:    mcpServer,
		sseServer:    mcpSSEServer,
		port:         cfg.Port,
		log:          log,
		allowOrigins: cfg.AllowOrigins,
	}
}

func newStreamableTransportServer(cfg *config.Config, mcpServer *server.MCPServer, log logger.Logger) MCPServer {
	httpServer := &http.Server{}
	mcpStreamableServer := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath(streamableEndpoint),
		server.WithStateLess(false),
		server.WithStreamableHTTPServer(httpServer),
	)
	httpServer.Handler = middlewares.CorsMiddleware(cfg.AllowOrigins, mcpStreamableServer)

	return &streamableHTTPServer{
		mcpServer:            mcpServer,
		streamableHTTPServer: mcpStreamableServer,
		port:                 cfg.Port,
		log:                  log,
		allowOrigins:         cfg.AllowOrigins,
	}
}

func NewServerFactory(handlerProvider interfaces.HandlerProvider) Factory {
	return &serverFactoryImpl{handlerProvider: handlerProvider}
}

func addrFromPort(port int) string {
	return ":" + strconv.Itoa(port)
}
