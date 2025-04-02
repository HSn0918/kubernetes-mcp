package server

import (
	"fmt"
	"net/http"
	"strconv"

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

	// 服务器在CreateServer时已完全配置好，直接启动
	addr := ":" + strconv.Itoa(s.port)
	return s.sseServer.Start(addr)
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

	// 准备服务器选项
	serverOptions := []server.ServerOption{
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	}

	// 添加钩子选项
	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(id any, method mcp.MCPMethod, message any) {
		log.Debug("Request received", "id", id, "method", method, "message", message)
	})
	hooks.AddOnSuccess(func(id any, method mcp.MCPMethod, message any, result any) {
		log.Info("Request successful", "id", id, "method", method)
	})
	hooks.AddOnError(func(id any, method mcp.MCPMethod, message any, err error) {
		log.Error("Request failed", "id", id, "method", method, "error", err)
	})
	serverOptions = append(serverOptions, server.WithHooks(hooks))

	// 创建基本MCP服务器
	mcpServer := server.NewMCPServer(
		"Kubernetes-mcp",
		"1.6.0",
		serverOptions...,
	)

	// 注册所有处理程序
	f.handlerProvider.RegisterAllHandlers(mcpServer)

	// 根据传输方式创建服务器
	if cfg.Transport == "sse" {
		// 配置服务器地址和基础URL
		port := cfg.Port
		addr := ":" + strconv.Itoa(port)

		// 使用配置中的BaseURL，如果未设置则使用默认的localhost
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:" + strconv.Itoa(port)
			log.Info("BaseURL not set, using default", "baseURL", baseURL)
		} else {
			log.Info("Using configured BaseURL", "baseURL", baseURL)
		}

		// 创建自定义的HTTP服务器，添加CORS支持
		httpServer := &http.Server{
			Addr: addr,
			// 应用CORS中间件，允许所有源
			Handler: middlewares.CreateCorsHandlerFunc(cfg.AllowOrigins, http.DefaultServeMux),
		}

		// 创建SSE服务器选项
		sseOptions := []server.SSEOption{
			server.WithBaseURL(baseURL),
			server.WithHTTPServer(httpServer), // 使用配置了CORS的HTTP服务器
		}

		// 创建SSE服务器
		mcpSseServer := server.NewSSEServer(mcpServer, sseOptions...)

		// 返回配置好的服务器实例
		return &sseServer{
			mcpServer:    mcpServer,
			sseServer:    mcpSseServer,
			port:         port,
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
