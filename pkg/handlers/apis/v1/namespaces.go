package v1

import (
	"context"
	"encoding/json"
	"fmt"
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
	LIST_NAMESPACES = "LIST_NAMESPACES"
)

// NamespaceHandlerImpl 命名空间处理程序实现
type NamespaceHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.NamespaceHandler = &NamespaceHandlerImpl{}

// NewNamespaceHandler 创建新的命名空间处理程序
func NewNamespaceHandler(client kubernetes.Client) interfaces.NamespaceHandler {
	return &NamespaceHandlerImpl{
		Handler: base.NewHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *NamespaceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case LIST_NAMESPACES:
		return h.ListNamespaces(ctx, request)
	default:
		return utils.NewErrorToolResult(fmt.Sprintf("unknown namespace method: %s", request.Method)), nil
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
		mcp.WithDescription("获取Kubernetes集群中所有命名空间的列表。提供命名空间的详细信息，包括状态、资源配额、限制范围等。适用于多租户管理、资源隔离、访问控制等场景。帮助了解集群的逻辑分区和资源分配情况。"),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按命名空间属性进行过滤。例如：'status.phase=Active'表示只显示活动状态的命名空间。支持多个条件，使用逗号分隔。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按命名空间标签进行过滤。例如：'environment=production'表示只显示生产环境的命名空间。支持多个标签，使用逗号分隔。"),
		),
		mcp.WithBoolean("showLabels",
			mcp.Description("是否显示命名空间的所有标签。启用后将在输出中包含完整的标签列表，有助于命名空间分类和管理。默认为false。"),
			mcp.DefaultBool(false),
		),
	), h.ListNamespaces)
}

// ListNamespaces 列出所有命名空间
func (h *NamespaceHandlerImpl) ListNamespaces(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Listing namespaces")

	// 获取所有命名空间
	namespaces := &corev1.NamespaceList{}
	err := h.Client.List(ctx, namespaces)
	if err != nil {
		h.Log.Error("Failed to list namespaces", "error", err)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to list namespaces: %v", err)), nil
	}

	// 构建命名空间信息列表
	namespaceInfos := make([]models.NamespaceInfo, 0, len(namespaces.Items))

	for _, ns := range namespaces.Items {
		// 获取命名空间状态
		status := string(ns.Status.Phase)

		// 构建命名空间信息
		nsInfo := models.NamespaceInfo{
			Name:         ns.Name,
			Status:       status,
			Labels:       ns.Labels,
			Annotations:  ns.Annotations,
			CreationTime: ns.CreationTimestamp.Time,
		}

		namespaceInfos = append(namespaceInfos, nsInfo)
	}

	// 创建完整响应
	response := models.NamespaceListResponse{
		Count:       len(namespaceInfos),
		Namespaces:  namespaceInfos,
		RetrievedAt: time.Now(),
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return utils.NewErrorToolResult(fmt.Sprintf("JSON序列化失败: %v", err)), nil
	}

	h.Log.Info("Namespaces listed successfully", "count", len(namespaces.Items))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}
