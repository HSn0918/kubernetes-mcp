package v1

import (
	"context"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
)

// ResourceHandlerImpl Storage资源处理程序实现
type ResourceHandlerImpl struct {
	handler     base.Handler
	baseHandler interfaces.BaseResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Storage资源处理程序
func NewResourceHandler(client kubernetes.Client) interfaces.ResourceHandler {
	baseHandler := base.NewHandler(client, interfaces.NamespaceScope, interfaces.StorageAPIGroup)
	baseResourceHandler := base.NewResourceHandlerPtr(baseHandler, "STORAGE")
	return &ResourceHandlerImpl{
		handler:     baseHandler,
		baseHandler: baseResourceHandler,
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用父类的处理方法
	return h.baseHandler.Handle(ctx, request)
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 使用父类的注册方法
	h.baseHandler.Register(server)
}

// GetScope 实现ToolHandler接口
func (h *ResourceHandlerImpl) GetScope() interfaces.ResourceScope {
	return h.handler.GetScope()
}

// GetAPIGroup 实现ToolHandler接口
func (h *ResourceHandlerImpl) GetAPIGroup() interfaces.APIGroup {
	return h.handler.GetAPIGroup()
}

// ListResources 实现ResourceHandler接口
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.ListResources(ctx, request)
}

// GetResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.GetResource(ctx, request)
}

// DescribeResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) DescribeResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DescribeResource(ctx, request)
}

// CreateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.CreateResource(ctx, request)
}

// UpdateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.UpdateResource(ctx, request)
}

// DeleteResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DeleteResource(ctx, request)
}
