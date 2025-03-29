package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolHandler 定义了MCP工具处理接口
type ToolHandler interface {
	// Handle 处理工具请求
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// Register 注册工具到MCP服务器
	Register(server *server.MCPServer)
}

// ResourceHandler 定义了资源处理接口
type ResourceHandler interface {
	ToolHandler

	// ListResources 列出资源
	ListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// GetResource 获取资源
	GetResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// CreateResource 创建资源
	CreateResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// UpdateResource 更新资源
	UpdateResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// DeleteResource 删除资源
	DeleteResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// NamespaceHandler 定义了命名空间处理接口
type NamespaceHandler interface {
	ToolHandler

	// ListNamespaces 列出命名空间
	ListNamespaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// HandlerProvider 提供所有工具处理程序
type HandlerProvider interface {
	// GetResourceHandler 获取资源处理程序
	GetResourceHandler() ResourceHandler

	// GetNamespaceHandler 获取命名空间处理程序
	GetNamespaceHandler() NamespaceHandler

	// RegisterAllHandlers 注册所有处理程序
	RegisterAllHandlers(server *server.MCPServer)
}
