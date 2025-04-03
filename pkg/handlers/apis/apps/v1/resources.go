package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

// ResourceHandlerImpl Apps资源处理程序实现
type ResourceHandlerImpl struct {
	handler     base.Handler
	baseHandler interfaces.BaseResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Apps资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewHandler(client, interfaces.NamespaceScope, interfaces.AppsAPIGroup)
	baseResourceHandler := base.NewResourceHandlerPtr(baseHandler, "APPS")
	return &ResourceHandlerImpl{
		handler:     baseHandler,
		baseHandler: baseResourceHandler,
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 检查是否是LIST_APPS_RESOURCES方法，使用我们的特殊实现
	if request.Method == fmt.Sprintf("LIST_%s_RESOURCES", h.baseHandler.GetResourcePrefix()) {
		return h.ListResources(ctx, request)
	}
	// 其他方法使用父类的处理方法
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

// ListResources 重写父类的列表方法，添加Apps特有的信息展示
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.handler.Log.Info("Listing Apps resources",
		"kind", kind,
		"apiVersion", apiVersion,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建列表对象
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    kind + "List",
	})

	// 列出资源
	err := h.handler.Client.List(ctx, list, &clientpkg.ListOptions{Namespace: namespace})
	if err != nil {
		h.handler.Log.Error("Failed to list Apps resources",
			"kind", kind,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list Apps resources: %v", err)
	}

	// 构建响应，为 Apps 资源特别定制
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d %s resources in namespace %s:\n\n", len(list.Items), kind, namespace))

	// 显示 Apps 资源的特定信息，例如 Deployments 的副本数
	for _, item := range list.Items {
		name := item.GetName()
		labels := item.GetLabels()

		result.WriteString(fmt.Sprintf("- %s\n", name))

		// 获取特定于资源类型的详细信息
		switch kind {
		case "Deployment":
			// 显示副本数和状态
			spec, found, _ := unstructured.NestedMap(item.Object, "spec")
			if found {
				replicas, exists, _ := unstructured.NestedInt64(spec, "replicas")
				if exists {
					result.WriteString(fmt.Sprintf("  Replicas: %d\n", replicas))
				}
			}

			status, found, _ := unstructured.NestedMap(item.Object, "status")
			if found {
				availableReplicas, exists, _ := unstructured.NestedInt64(status, "availableReplicas")
				if exists {
					result.WriteString(fmt.Sprintf("  Available: %d\n", availableReplicas))
				}

				readyReplicas, exists, _ := unstructured.NestedInt64(status, "readyReplicas")
				if exists {
					result.WriteString(fmt.Sprintf("  Ready: %d\n", readyReplicas))
				}
			}

		case "StatefulSet":
			// 显示副本数和状态
			spec, found, _ := unstructured.NestedMap(item.Object, "spec")
			if found {
				replicas, exists, _ := unstructured.NestedInt64(spec, "replicas")
				if exists {
					result.WriteString(fmt.Sprintf("  Replicas: %d\n", replicas))
				}
			}

		case "DaemonSet":
			// 显示节点调度情况
			status, found, _ := unstructured.NestedMap(item.Object, "status")
			if found {
				numberReady, exists, _ := unstructured.NestedInt64(status, "numberReady")
				if exists {
					result.WriteString(fmt.Sprintf("  Ready: %d\n", numberReady))
				}

				desiredNumberScheduled, exists, _ := unstructured.NestedInt64(status, "desiredNumberScheduled")
				if exists {
					result.WriteString(fmt.Sprintf("  Desired: %d\n", desiredNumberScheduled))
				}
			}
		}

		// 显示通用标签信息
		if len(labels) > 0 {
			result.WriteString("  Labels:\n")
			for k, v := range labels {
				result.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
			}
		}

		result.WriteString("\n")
	}

	h.handler.Log.Info("Apps resources listed successfully",
		"kind", kind,
		"namespace", namespace,
		"count", len(list.Items),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
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
