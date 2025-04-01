package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
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

	// 构建JSON响应
	nodeInfos := make([]models.NodeInfo, 0, len(nodes.Items))

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

		// 获取节点角色
		roles := []string{}
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
				roles = append(roles, "control-plane")
			} else if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				roles = append(roles, role)
			}
		}

		// 获取节点污点
		taints := make([]models.Taint, 0, len(node.Spec.Taints))
		for _, taint := range node.Spec.Taints {
			taints = append(taints, models.Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: string(taint.Effect),
			})
		}

		// 获取可分配资源
		allocatableCPU := node.Status.Allocatable.Cpu().String()
		allocatableMemory := node.Status.Allocatable.Memory().String()
		allocatablePods := node.Status.Allocatable.Pods().String()

		// 构建节点信息
		nodeInfo := models.NodeInfo{
			Name:              node.Name,
			Status:            status,
			KubeletVersion:    kubeletVersion,
			OSImage:           osImage,
			KernelVersion:     kernelVersion,
			Architecture:      architecture,
			InternalIP:        internalIP,
			ExternalIP:        externalIP,
			Roles:             roles,
			Labels:            node.Labels,
			Taints:            taints,
			AllocatableCPU:    allocatableCPU,
			AllocatableMemory: allocatableMemory,
			AllocatablePods:   allocatablePods,
			CreationTime:      node.CreationTimestamp.Time,
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	// 创建完整响应
	response := models.NodeListResponse{
		Count:       len(nodeInfos),
		Nodes:       nodeInfos,
		RetrievedAt: time.Now(),
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	h.Log.Info("Nodes listed successfully", "count", len(nodes.Items))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
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
