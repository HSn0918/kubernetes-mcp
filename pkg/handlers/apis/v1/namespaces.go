package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
)

// 定义常量
const (
	LIST_NAMESPACES = "kubernetes.listNamespaces"
)

// NamespaceHandlerImpl 命名空间处理程序实现
type NamespaceHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.NamespaceHandler = &NamespaceHandlerImpl{}

// NewNamespaceHandler 创建新的命名空间处理程序
func NewNamespaceHandler(client client.KubernetesClient) interfaces.NamespaceHandler {
	return &NamespaceHandlerImpl{
		Handler: base.NewBaseHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *NamespaceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case LIST_NAMESPACES:
		return h.ListNamespaces(ctx, request)
	default:
		return nil, fmt.Errorf("unknown namespace method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *NamespaceHandlerImpl) Register(server *server.MCPServer) {
	h.Log.Info("Registering namespace handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
	)

	// 注册列出命名空间工具
	server.AddTool(mcp.NewTool(LIST_NAMESPACES,
		mcp.WithDescription("List all namespaces (Cluster-scoped)"),
	), h.ListNamespaces)
}

// ListNamespaces 实现接口方法
func (h *NamespaceHandlerImpl) ListNamespaces(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Listing namespaces")

	// 创建命名空间列表
	namespaces := &corev1.NamespaceList{}

	// 获取所有命名空间
	err := h.Client.List(ctx, namespaces)
	if err != nil {
		h.Log.Error("Failed to list namespaces", "error", err)
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d namespaces:\n\n", len(namespaces.Items)))

	for _, ns := range namespaces.Items {
		result.WriteString(fmt.Sprintf("- %s (Status: %s)\n", ns.Name, ns.Status.Phase))
	}

	h.Log.Info("Namespaces listed successfully", "count", len(namespaces.Items))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// GetScope 实现ToolHandler接口
func (h *NamespaceHandlerImpl) GetScope() interfaces.ResourceScope {
	return h.Scope
}

// GetAPIGroup 实现ToolHandler接口
func (h *NamespaceHandlerImpl) GetAPIGroup() interfaces.APIGroup {
	return h.Group
}
