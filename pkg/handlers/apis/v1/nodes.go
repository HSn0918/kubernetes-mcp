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
	LIST_NODES = "LIST_NODES"
)

// NodeHandlerImpl 节点处理程序实现
type NodeHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.ToolHandler = &NodeHandlerImpl{}

// NewNodeHandler 创建新的节点处理程序
func NewNodeHandler(client client.KubernetesClient) interfaces.ToolHandler {
	return &NodeHandlerImpl{
		Handler: base.NewBaseHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *NodeHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case LIST_NODES:
		return h.ListNodes(ctx, request)
	default:
		return nil, fmt.Errorf("unknown node method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *NodeHandlerImpl) Register(server *server.MCPServer) {
	h.Log.Info("Registering node handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
	)

	// 注册列出节点工具
	server.AddTool(mcp.NewTool(LIST_NODES,
		mcp.WithDescription("List all nodes (Cluster-scoped)"),
	), h.ListNodes)
}

// ListNodes 列出所有节点
func (h *NodeHandlerImpl) ListNodes(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Listing nodes")

	// 创建节点列表
	nodes := &corev1.NodeList{}

	// 获取所有节点
	err := h.Client.List(ctx, nodes)
	if err != nil {
		h.Log.Error("Failed to list nodes", "error", err)
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d nodes:\n\n", len(nodes.Items)))

	for _, node := range nodes.Items {
		// 获取节点状态
		var status string
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// 获取基本信息
		kubeletVersion := node.Status.NodeInfo.KubeletVersion
		osImage := node.Status.NodeInfo.OSImage
		kernelVersion := node.Status.NodeInfo.KernelVersion
		architecture := node.Status.NodeInfo.Architecture

		// 获取地址
		var internalIP, externalIP string
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				internalIP = addr.Address
			} else if addr.Type == corev1.NodeExternalIP {
				externalIP = addr.Address
			}
		}

		// 添加节点信息到结果中
		result.WriteString(fmt.Sprintf("- %s (Status: %s)\n", node.Name, status))
		result.WriteString(fmt.Sprintf("  Kubelet Version: %s\n", kubeletVersion))
		result.WriteString(fmt.Sprintf("  OS: %s, Kernel: %s, Arch: %s\n", osImage, kernelVersion, architecture))

		if internalIP != "" {
			result.WriteString(fmt.Sprintf("  Internal IP: %s\n", internalIP))
		}

		if externalIP != "" {
			result.WriteString(fmt.Sprintf("  External IP: %s\n", externalIP))
		}

		// 添加角色标签（如果有）
		roles := []string{}
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
				roles = append(roles, "control-plane")
			} else if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				roles = append(roles, role)
			}
		}

		if len(roles) > 0 {
			result.WriteString(fmt.Sprintf("  Roles: %s\n", strings.Join(roles, ", ")))
		}

		result.WriteString("\n")
	}

	h.Log.Info("Nodes listed successfully", "count", len(nodes.Items))

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
func (h *NodeHandlerImpl) GetScope() interfaces.ResourceScope {
	return h.Scope
}

// GetAPIGroup 实现ToolHandler接口
func (h *NodeHandlerImpl) GetAPIGroup() interfaces.APIGroup {
	return h.Group
}
