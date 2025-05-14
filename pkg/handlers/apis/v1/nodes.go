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

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
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
func NewNodeHandler(client kubernetes.Client) interfaces.ToolHandler {
	return &NodeHandlerImpl{
		Handler: base.NewHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *NodeHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case LIST_NODES:
		return h.ListNodes(ctx, request)
	default:
		return utils.NewErrorToolResult(fmt.Sprintf("unknown node method: %s", request.Method)), nil
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
		mcp.WithDescription("获取Kubernetes集群中所有节点的列表。提供节点的详细信息，包括状态、容量、可分配资源、标签、污点等。适用于集群管理、资源规划、节点维护等场景。支持节点健康状态监控和资源分配决策。"),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按节点属性进行过滤。例如：'spec.unschedulable=false'表示只显示可调度节点。支持多个条件，使用逗号分隔。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按节点标签进行过滤。例如：'kubernetes.io/role=master'表示只显示主节点。支持多个标签，使用逗号分隔。"),
		),
		mcp.WithBoolean("showLabels",
			mcp.Description("是否显示节点的所有标签。启用后将在输出中包含完整的标签列表，有助于标签管理和节点分类。默认为false。"),
			mcp.DefaultBool(false),
		),
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
		return utils.NewErrorToolResult(fmt.Sprintf("failed to list nodes: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("JSON序列化失败: %v", err)), nil
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
