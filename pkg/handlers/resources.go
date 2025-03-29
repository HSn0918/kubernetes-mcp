package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// resourceHandlerImpl 资源处理程序实现
type resourceHandlerImpl struct {
	baseHandler
}

// 确保实现了接口
var _ ResourceHandler = &resourceHandlerImpl{}

// NewResourceHandler 创建新的资源处理程序
func NewResourceHandler(client client.KubernetesClient) ResourceHandler {
	return &resourceHandlerImpl{
		baseHandler: newBaseHandler(client),
	}
}

// Handle 实现接口方法
func (h *resourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case LIST_RESOURCES:
		return h.ListResources(ctx, request)
	case GET_RESOURCE:
		return h.GetResource(ctx, request)
	case CREATE_RESOURCE:
		return h.CreateResource(ctx, request)
	case UPDATE_RESOURCE:
		return h.UpdateResource(ctx, request)
	case DELETE_RESOURCE:
		return h.DeleteResource(ctx, request)
	default:
		return nil, fmt.Errorf("unknown resource method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *resourceHandlerImpl) Register(server *server.MCPServer) {
	h.log.Info("Registering resource handlers")

	// 注册列出资源工具
	server.AddTool(mcp.NewTool(LIST_RESOURCES,
		mcp.WithDescription("List Kubernetes resources"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Deployment, Service, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1, apps/v1, etc.)"),
			mcp.DefaultString("v1"),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(GET_RESOURCE,
		mcp.WithDescription("Get a specific Kubernetes resource"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Deployment, Service, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1, apps/v1, etc.)"),
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
	server.AddTool(mcp.NewTool(CREATE_RESOURCE,
		mcp.WithDescription("Create a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(UPDATE_RESOURCE,
		mcp.WithDescription("Update a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(DELETE_RESOURCE,
		mcp.WithDescription("Delete a Kubernetes resource"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Deployment, Service, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1, apps/v1, etc.)"),
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
func (h *resourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.log.Info("Listing resources",
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
	err := h.client.List(ctx, list, &clientpkg.ListOptions{Namespace: namespace})
	if err != nil {
		h.log.Error("Failed to list resources",
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

	h.log.Info("Resources listed successfully",
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
func (h *resourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.log.Info("Getting resource",
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
	err := h.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		h.log.Error("Failed to get resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	// 转换为YAML
	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		h.log.Error("Failed to marshal resource to YAML",
			"kind", kind,
			"name", name,
			"error", err,
		)
		return nil, fmt.Errorf("failed to marshal to YAML: %v", err)
	}

	h.log.Info("Resource retrieved successfully",
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
func (h *resourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.log.Info("Creating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	h.log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 创建资源
	err = h.client.Create(ctx, obj)
	if err != nil {
		h.log.Error("Failed to create resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	h.log.Info("Resource created successfully",
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
func (h *resourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.log.Info("Updating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	h.log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 更新资源
	err = h.client.Update(ctx, obj)
	if err != nil {
		h.log.Error("Failed to update resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to update resource: %v", err)
	}

	h.log.Info("Resource updated successfully",
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
func (h *resourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.log.Info("Deleting resource",
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
	err := h.client.Delete(ctx, obj)
	if err != nil {
		h.log.Error("Failed to delete resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to delete resource: %v", err)
	}

	h.log.Info("Resource deleted successfully",
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
