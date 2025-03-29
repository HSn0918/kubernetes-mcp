package core

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// ResourceHandlerImpl 核心资源处理程序实现
type ResourceHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的核心资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	return &ResourceHandlerImpl{
		Handler: base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case base.LIST_RESOURCES:
		return h.ListResources(ctx, request)
	case base.GET_RESOURCE:
		return h.GetResource(ctx, request)
	case base.CREATE_RESOURCE:
		return h.CreateResource(ctx, request)
	case base.UPDATE_RESOURCE:
		return h.UpdateResource(ctx, request)
	case base.DELETE_RESOURCE:
		return h.DeleteResource(ctx, request)
	default:
		return nil, fmt.Errorf("unknown resource method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	h.Log.Info("Registering core resource handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
	)

	// 注册列出资源工具
	server.AddTool(mcp.NewTool(base.LIST_RESOURCES,
		mcp.WithDescription("List Kubernetes resources (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(base.GET_RESOURCE,
		mcp.WithDescription("Get a specific Kubernetes resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
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
	server.AddTool(mcp.NewTool(base.CREATE_RESOURCE,
		mcp.WithDescription("Create a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(base.UPDATE_RESOURCE,
		mcp.WithDescription("Update a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(base.DELETE_RESOURCE,
		mcp.WithDescription("Delete a Kubernetes resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
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

// ListResources 实现接口方法
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Listing resources",
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
		h.Log.Error("Failed to list resources",
			"kind", kind,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list resources: %v", err)
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d %s resources in namespace %s:\n\n", len(list.Items), kind, namespace))

	for _, item := range list.Items {
		result.WriteString(fmt.Sprintf("Name: %s\n", item.GetName()))
	}

	h.Log.Info("Resources listed successfully",
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
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Getting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建对象
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	// 获取资源
	err := h.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		h.Log.Error("Failed to get resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("core resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)
		}
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	// 转换为YAML
	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		h.Log.Error("Failed to marshal resource to YAML",
			"kind", kind,
			"name", name,
			"error", err,
		)
		return nil, fmt.Errorf("failed to marshal to YAML: %v", err)
	}

	h.Log.Info("Resource retrieved successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(yamlData),
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

	h.Log.Info("Creating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}
	if obj.GetNamespace() == "" {
		obj.SetNamespace("default")
		h.Log.Debug("Empty namespace, using default namespace")
	}
	h.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 创建资源
	err = h.Client.Create(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to create resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	h.Log.Info("Resource created successfully",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created %s/%s in namespace %s",
					obj.GetKind(), obj.GetName(), obj.GetNamespace()),
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

	h.Log.Info("Updating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	h.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 更新资源
	err = h.Client.Update(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to update resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to update resource: %v", err)
	}

	h.Log.Info("Resource updated successfully",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated %s/%s in namespace %s",
					obj.GetKind(), obj.GetName(), obj.GetNamespace()),
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

	h.Log.Info("Deleting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建对象
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	// 删除资源
	err := h.Client.Delete(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to delete resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to delete resource: %v", err)
	}

	h.Log.Info("Resource deleted successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted %s/%s from namespace %s", kind, name, namespace),
			},
		},
	}, nil
}
