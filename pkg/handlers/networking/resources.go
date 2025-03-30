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
	base.ResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Networking资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.NetworkingAPIGroup)
	return &ResourceHandlerImpl{
		ResourceHandler: base.NewResourceHandler(baseHandler, "NETWORKING"),
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用父类的处理方法
	return h.ResourceHandler.Handle(ctx, request)
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 使用父类的注册方法
	h.ResourceHandler.Register(server)
}
