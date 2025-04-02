package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/hsn0918/kubernetes-mcp/pkg/middlewares"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// stdioServer 标准输入/输出模式服务器
type stdioServer struct {
	mcpServer *server.MCPServer
	log       logger.Logger
}

// sseServer Server-Sent Events模式服务器
type sseServer struct {
	mcpServer    *server.MCPServer
	sseServer    *server.SSEServer
	port         int
	log          logger.Logger
	allowOrigins string
}

// serverFactoryImpl 服务器工厂实现
type serverFactoryImpl struct {
	handlerProvider interfaces.HandlerProvider
}

// 确保实现了接口
var _ MCPServer = &stdioServer{}
var _ MCPServer = &sseServer{}
var _ Factory = &serverFactoryImpl{}

// GetServer 实现接口方法
func (s *stdioServer) GetServer() *server.MCPServer {
	return s.mcpServer
}

// Start 实现接口方法
func (s *stdioServer) Start() error {
	s.log.Info("Starting stdio server")
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("server error: %v", err)
	}
	return nil
}

// Stop 实现接口方法
func (s *stdioServer) Stop() error {
	s.log.Info("Stopping stdio server")
	// stdio服务器不需要额外的停止逻辑
	return nil
}

// GetServer 实现接口方法
func (s *sseServer) GetServer() *server.MCPServer {
	return s.mcpServer
}

// Start 实现接口方法
func (s *sseServer) Start() error {
	s.log.Info("Starting SSE server", "port", s.port, "allowOrigins", s.allowOrigins)

	// 定义一个处理函数，用于代理到SSE服务器
	// 实际上，我们将使用SSE服务器的Start方法，但是我们需要自己实现CORS
	// 我们只能通过自定义http服务器来实现
	addr := ":" + strconv.Itoa(s.port)

	// 创建HTTP服务器
	server := &http.Server{
		Addr: addr,
		// 使用middlewares包中的CreateCorsHandlerFunc创建CORS处理函数
		Handler: middlewares.CreateCorsHandlerFunc(s.allowOrigins, http.DefaultServeMux),
	}

	// 在启动我们的服务器之前，确保SSE服务器已经注册了它的处理器到DefaultServeMux
	// 我们在一个单独的goroutine中启动SSE服务器，但不真正使用它
	// 这只是为了让它注册处理器
	go func() {
		s.sseServer.Start(addr)
	}()

	// 短暂暂停，确保SSE服务器已经注册处理器
	time.Sleep(100 * time.Millisecond)

	// 启动我们的HTTP服务器
	s.log.Info("SSE server starting with CORS support", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %v", err)
	}

	return nil
}

// Stop 实现接口方法
func (s *sseServer) Stop() error {
	s.log.Info("Stopping SSE server")
	// 可以添加额外的SSE服务器清理逻辑
	return nil
}

// CreateServer 实现接口方法
func (f *serverFactoryImpl) CreateServer(cfg *config.Config) (MCPServer, error) {
	log := logger.GetLogger()

	// 创建基本MCP服务器
	mcpServer := server.NewMCPServer(
		"Kubernetes-mcp",
		"1.1.0",
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

	// 设置钩子
	setupHooks(mcpServer)

	// 注册所有处理程序
	f.handlerProvider.RegisterAllHandlers(mcpServer)

	// 根据传输方式创建服务器
	if cfg.Transport == "sse" {
		baseURL := "http://localhost:" + strconv.Itoa(cfg.Port)

		// 创建SSE服务器
		mcpSseServer := server.NewSSEServer(
			mcpServer,
			server.WithBaseURL(baseURL),
		)

		return &sseServer{
			mcpServer:    mcpServer,
			sseServer:    mcpSseServer,
			port:         cfg.Port,
			log:          log,
			allowOrigins: cfg.AllowOrigins,
		}, nil
	} else {
		return &stdioServer{
			mcpServer: mcpServer,
			log:       log,
		}, nil
	}
}

// NewServerFactory 创建新的服务器工厂
func NewServerFactory(handlerProvider interfaces.HandlerProvider) Factory {
	return &serverFactoryImpl{
		handlerProvider: handlerProvider,
	}
}

// setupHooks 设置MCP服务器钩子
func setupHooks(mcpServer *server.MCPServer) {
	log := logger.GetLogger()
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(id any, method mcp.MCPMethod, message any) {
		log.Debug("Request received",
			"id", id,
			"method", method,
			"message", message,
		)
	})

	hooks.AddOnSuccess(func(id any, method mcp.MCPMethod, message any, result any) {
		log.Info("Request successful",
			"id", id,
			"method", method,
		)
	})

	hooks.AddOnError(func(id any, method mcp.MCPMethod, message any, err error) {
		log.Error("Request failed",
			"id", id,
			"method", method,
			"error", err,
		)
	})

	server.WithHooks(hooks)(mcpServer)
}
