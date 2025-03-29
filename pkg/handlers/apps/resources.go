package apps

import (
	"context"
	"fmt"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceHandlerImpl Apps资源处理程序实现
type ResourceHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Apps资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	return &ResourceHandlerImpl{
		Handler: base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.AppsAPIGroup),
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case base.LIST_APPS_RESOURCES:
		return h.ListResources(ctx, request)
	case base.GET_APPS_RESOURCE:
		return h.GetResource(ctx, request)
	case base.CREATE_APPS_RESOURCE:
		return h.CreateResource(ctx, request)
	case base.UPDATE_APPS_RESOURCE:
		return h.UpdateResource(ctx, request)
	case base.DELETE_APPS_RESOURCE:
		return h.DeleteResource(ctx, request)
	default:
		return nil, fmt.Errorf("unknown apps resource method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	h.Log.Info("Registering apps resource handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
	)

	// 注册列出资源工具
	server.AddTool(mcp.NewTool(base.LIST_APPS_RESOURCES,
		mcp.WithDescription("List Apps Kubernetes resources (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Deployment, StatefulSet, DaemonSet, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (apps/v1)"),
			mcp.DefaultString("apps/v1"),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(base.GET_APPS_RESOURCE,
		mcp.WithDescription("Get a specific Apps resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Deployment, StatefulSet, DaemonSet, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (apps/v1)"),
			mcp.DefaultString("apps/v1"),
		),
		mcp.WithString("name",
			mcp.Description("Name of the resource"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.GetResource)

	// 注册创建资源工具
	server.AddTool(mcp.NewTool(base.CREATE_APPS_RESOURCE,
		mcp.WithDescription("Create an Apps resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the Apps resource"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(base.UPDATE_APPS_RESOURCE,
		mcp.WithDescription("Update an Apps resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the Apps resource"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(base.DELETE_APPS_RESOURCE,
		mcp.WithDescription("Delete an Apps resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Deployment, StatefulSet, DaemonSet, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (apps/v1)"),
			mcp.DefaultString("apps/v1"),
		),
		mcp.WithString("name",
			mcp.Description("Name of the resource"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.DeleteResource)
}

// 以下为核心实现方法，为了简化代码，重用之前ResourceHandlerImpl的业务逻辑实现

// 从core.ResourceHandlerImpl直接复制过来的核心方法调用代码
// 实际使用时可以考虑共享这些逻辑或通过嵌入组合方式实现

// ListResources 实现接口方法
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Listing Apps resources",
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
	err := h.Client.List(ctx, list, &clientpkg.ListOptions{Namespace: namespace})
	if err != nil {
		h.Log.Error("Failed to list Apps resources",
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
		result.WriteString(fmt.Sprintf("Name: %s\n", item.GetName()))

		// 获取额外的特定于 Apps 资源的信息
		if kind == "Deployment" {
			replicas, found, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
			if found {
				result.WriteString(fmt.Sprintf("  Replicas: %d\n", replicas))
			}

			// 获取标签选择器信息
			matchLabels, found, _ := unstructured.NestedMap(item.Object, "spec", "selector", "matchLabels")
			if found && len(matchLabels) > 0 {
				result.WriteString("  Selector: ")
				for k, v := range matchLabels {
					result.WriteString(fmt.Sprintf("%s=%v ", k, v))
				}
				result.WriteString("\n")
			}
		} else if kind == "StatefulSet" {
			replicas, found, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
			if found {
				result.WriteString(fmt.Sprintf("  Replicas: %d\n", replicas))
			}

			// 获取更新策略
			updateStrategy, found, _ := unstructured.NestedString(item.Object, "spec", "updateStrategy", "type")
			if found {
				result.WriteString(fmt.Sprintf("  Update Strategy: %s\n", updateStrategy))
			}
		} else if kind == "DaemonSet" {
			// 获取更新策略
			updateStrategy, found, _ := unstructured.NestedString(item.Object, "spec", "updateStrategy", "type")
			if found {
				result.WriteString(fmt.Sprintf("  Update Strategy: %s\n", updateStrategy))
			}
		}

		result.WriteString("\n")
	}

	h.Log.Info("Apps resources listed successfully",
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

// GetResource 实现接口方法
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// TODO: 实现具体的资源获取逻辑
	h.Log.Info("Getting Apps resource")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Apps resource fetch not implemented yet",
			},
		},
	}, nil
}

// CreateResource 实现接口方法
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// TODO: 实现具体的资源创建逻辑
	h.Log.Info("Creating Apps resource")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Apps resource creation not implemented yet",
			},
		},
	}, nil
}

// UpdateResource 实现接口方法
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// TODO: 实现具体的资源更新逻辑
	h.Log.Info("Updating Apps resource")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Apps resource update not implemented yet",
			},
		},
	}, nil
}

// DeleteResource 实现接口方法
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// TODO: 实现具体的资源删除逻辑
	h.Log.Info("Deleting Apps resource")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Apps resource deletion not implemented yet",
			},
		},
	}, nil
}
