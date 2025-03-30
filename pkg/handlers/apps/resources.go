package apps

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

// ResourceHandlerImpl Apps资源处理程序实现
type ResourceHandlerImpl struct {
	handler         base.Handler
	resourceHandler *base.ResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的Apps资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.AppsAPIGroup)
	resourceHandler := base.NewResourceHandler(baseHandler, "APPS")
	return &ResourceHandlerImpl{
		handler:         baseHandler,
		resourceHandler: &resourceHandler,
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 检查是否是LIST_APPS_RESOURCES方法，使用我们的特殊实现
	if request.Method == fmt.Sprintf("LIST_%s_RESOURCES", h.resourceHandler.GetResourcePrefix()) {
		return h.ListResources(ctx, request)
	}
	// 其他方法使用父类的处理方法
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

// GetResource 实现接口方法
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.handler.Log.Info("Getting Apps resource")
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)
	h.handler.Log.Info("Getting Apps resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)
	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	err := h.handler.Client.Get(ctx, clientpkg.ObjectKey{Namespace: namespace, Name: name}, obj)
	if err != nil {
		h.handler.Log.Error("Failed to get Apps resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("apps resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)
		}
		return nil, fmt.Errorf("failed to get Apps resource (Kind: %s, Name: %s): %v", kind, name, err)
	}
	h.handler.Log.Info("Apps resource retrieved successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("YAML for %s '%s' in namespace '%s':\n\n%s", kind, name, namespace, obj),
			},
		},
	}, nil
}

// CreateResource 实现接口方法
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)
	h.handler.Log.Info("Creating Apps resource")
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.handler.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}
	if obj.GetNamespace() == "" {
		obj.SetNamespace("default")
		h.handler.Log.Debug("Empty namespace, using default namespace")
	}
	h.handler.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)
	err = h.handler.Client.Create(ctx, obj)
	if err != nil {
		h.handler.Log.Error("create apps failed")
		return nil, fmt.Errorf("failed to create Apps resource: %v", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Apps resource '%s/%s' created successfully in namespace %s", obj.GetKind(), obj.GetName(), obj.GetNamespace()),
			},
		},
	}, nil
}

// UpdateResource 实现接口方法
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)
	h.handler.Log.Info("Updating resource from YAML")
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.handler.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}
	h.handler.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)
	err = h.handler.Client.Update(ctx, obj)

	if err != nil {
		h.handler.Log.Error("Failed to update resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to update resource: %v", err)
	}
	h.handler.Log.Info("Resource updated successfully",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Apps resource '%s/%s' updated successfully in namespace %s", obj.GetKind(), obj.GetName(), obj.GetNamespace()),
			},
		},
	}, nil
}

// DeleteResource 实现接口方法
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)

	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.handler.Log.Info("Deleting Apps resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)
	gvk := utils.ParseGVK(apiVersion, kind)

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	err := h.handler.Client.Delete(ctx, obj)
	if err != nil {
		h.handler.Log.Error("Failed to delete resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to delete resource: %v", err)
	}
	h.handler.Log.Info("Resource deleted successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Apps resource deletion not implemented yet",
			},
		},
	}, nil
}
