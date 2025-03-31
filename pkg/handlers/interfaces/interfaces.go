package interfaces

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ResourceScope 定义资源的作用域
type ResourceScope string

const (
	// ClusterScope 表示集群作用域资源
	ClusterScope ResourceScope = "cluster"
	// NamespaceScope 表示命名空间作用域资源
	NamespaceScope ResourceScope = "namespace"
)

// APIGroup 定义资源的API组
type APIGroup string

const (
	// CoreAPIGroup 代表核心API组 (v1)
	CoreAPIGroup APIGroup = "core"
	// AppsAPIGroup 代表应用API组 (apps/v1)
	AppsAPIGroup APIGroup = "apps"
	// BatchAPIGroup 代表批处理API组 (batch/v1)
	BatchAPIGroup APIGroup = "batch"
	// NetworkingAPIGroup 代表网络API组 (networking.k8s.io/v1)
	NetworkingAPIGroup APIGroup = "networking"
	// RbacAPIGroup 代表RBAC API组 (rbac.authorization.k8s.io/v1)
	RbacAPIGroup APIGroup = "rbac"
	// StorageAPIGroup 代表存储API组 (storage.k8s.io/v1)
	StorageAPIGroup APIGroup = "storage"
	// ApiextensionsAPIGroup 代表API扩展API组 (apiextensions.k8s.io/v1)
	ApiextensionsAPIGroup APIGroup = "apiextensions"
	// PolicyAPIGroup 代表策略API组 (policy/v1beta1)
	PolicyAPIGroup APIGroup = "policy"
	// AutoscalingAPIGroup 代表自动伸缩API组 (autoscaling/v1)
	AutoscalingAPIGroup APIGroup = "autoscaling"
)

// ToolHandler 定义MCP工具处理接口
type ToolHandler interface {
	// Handle 处理工具请求
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// Register 注册工具到MCP服务器
	Register(server *server.MCPServer)

	// GetScope 返回处理程序的作用域（集群或命名空间）
	GetScope() ResourceScope

	// GetAPIGroup 返回处理程序的API组
	GetAPIGroup() APIGroup
}

// ResourceHandler 定义资源处理接口
type ResourceHandler interface {
	ToolHandler

	// ListResources 列出资源
	ListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// GetResource 获取资源
	GetResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// DescribeResource 详细描述资源
	DescribeResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// CreateResource 创建资源
	CreateResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// UpdateResource 更新资源
	UpdateResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// DeleteResource 删除资源
	DeleteResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// NamespaceHandler 定义命名空间处理接口
type NamespaceHandler interface {
	ToolHandler

	// ListNamespaces 列出命名空间
	ListNamespaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// HandlerProvider 提供所有工具处理程序
type HandlerProvider interface {
	// GetHandlers 返回所有处理程序
	GetHandlers() []ToolHandler

	// RegisterAllHandlers 注册所有处理程序
	RegisterAllHandlers(server *server.MCPServer)
}

// HandlerFactory 提供创建各种资源处理程序的工厂方法
type HandlerFactory interface {
	// CreateCoreHandler 创建核心资源处理程序
	CreateCoreHandler() ResourceHandler

	// CreateAppsHandler 创建应用资源处理程序
	CreateAppsHandler() ResourceHandler

	// CreateBatchHandler 创建批处理资源处理程序
	CreateBatchHandler() ResourceHandler

	// CreateNetworkingHandler 创建网络资源处理程序
	CreateNetworkingHandler() ResourceHandler

	// CreateStorageHandler 创建存储资源处理程序
	CreateStorageHandler() ResourceHandler

	// CreateRbacHandler 创建RBAC资源处理程序
	CreateRbacHandler() ResourceHandler

	// CreatePolicyHandler 创建策略资源处理程序
	CreatePolicyHandler() ResourceHandler

	// CreateApiExtensionsHandler 创建API扩展资源处理程序
	CreateApiExtensionsHandler() ResourceHandler

	// CreateAutoscalingHandler 创建自动伸缩资源处理程序
	CreateAutoscalingHandler() ResourceHandler

	// CreateNamespaceHandler 创建命名空间处理程序
	CreateNamespaceHandler() NamespaceHandler

	// CreateNodeHandler 创建节点处理程序
	CreateNodeHandler() ToolHandler
}

// BaseResourceHandler 定义资源处理器的基础实现
type BaseResourceHandler interface {
	ResourceHandler
	GetResourcePrefix() string
	GetNamespaceWithDefault(incomingNamespace string) string
}
