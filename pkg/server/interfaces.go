package server

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer 定义了MCP服务器接口
type MCPServer interface {
	// Start 启动服务器
	Start() error

	// Stop 停止服务器
	Stop() error

	// GetServer 获取底层MCP服务器
	GetServer() *server.MCPServer
}

// ServerFactory 创建服务器的工厂接口
type ServerFactory interface {
	// CreateServer 创建新的MCP服务器
	CreateServer(config *config.Config) (MCPServer, error)
}
