package networking

import (
	"context"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ResourceHandlerImpl Networking资源处理程序实现
type ResourceHandlerImpl struct {
	handler         base.Handler
	resourceHandler *base.ResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Networking资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.NetworkingAPIGroup)
	resourceHandler := base.NewResourceHandler(baseHandler, "NETWORKING")
	return &ResourceHandlerImpl{
		handler:         baseHandler,
		resourceHandler: &resourceHandler,
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用父类的处理方法
	return h.resourceHandler.Handle(ctx, request)
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 使用父类的注册方法
	h.resourceHandler.Register(server)
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
	return h.resourceHandler.ListResources(ctx, request)
}

// GetResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.resourceHandler.GetResource(ctx, request)
}

// CreateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.resourceHandler.CreateResource(ctx, request)
}

// UpdateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.resourceHandler.UpdateResource(ctx, request)
}

// DeleteResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.resourceHandler.DeleteResource(ctx, request)
}
