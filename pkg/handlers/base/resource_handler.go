package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

// 确保ResourceHandler实现了BaseResourceHandler接口
var _ interfaces.BaseResourceHandler = (*ResourceHandler)(nil)

// ResourceHandler 提供通用的资源处理功能
type ResourceHandler struct {
	Handler
	resourcePrefix string
}

// NewResourceHandler 创建新的资源处理器
func NewResourceHandler(h Handler, resourcePrefix string) ResourceHandler {
	return ResourceHandler{
		Handler:        h,
		resourcePrefix: resourcePrefix,
	}
}

// NewResourceHandlerPtr NewResourceHandler 创建新的资源处理器（指针版本）
func NewResourceHandlerPtr(h Handler, resourcePrefix string) *ResourceHandler {
	return lo.ToPtr(NewResourceHandler(h, resourcePrefix))
}

// Register 注册通用资源处理工具
func (h *ResourceHandler) Register(server *server.MCPServer) {
	prefix := h.resourcePrefix
	h.Log.Info("Registering resource handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
		"prefix", prefix,
	)
	// 注册列出资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("LIST_%s_RESOURCES", prefix),
		mcp.WithDescription(fmt.Sprintf("List %s Kubernetes resources (%s-scoped)", h.Group, h.Scope)),
		mcp.WithString("kind",
			mcp.Description("Kind of resource"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Label selector (e.g. 'key1=value1,key2=value2')"),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("GET_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("Get a specific %s resource (%s-scoped)", h.Group, h.Scope)),
		mcp.WithString("kind",
			mcp.Description("Kind of resource"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version"),
			mcp.Required(),
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

	// 注册描述资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("DESCRIBE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("Describe a specific %s resource with detailed information (%s-scoped)", h.Group, h.Scope)),
		mcp.WithString("kind",
			mcp.Description("Kind of resource"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("Name of the resource"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.DescribeResource)

	// 注册创建资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("CREATE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("Create a %s resource from YAML", h.Group)),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("UPDATE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("Update a %s resource from YAML", h.Group)),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("DELETE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("Delete a %s resource (%s-scoped)", h.Group, h.Scope)),
		mcp.WithString("kind",
			mcp.Description("Kind of resource"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version"),
			mcp.Required(),
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

// GetNamespaceWithDefault 获取命名空间，如果为空则使用kubeconfig中的命名空间，再为空则使用default
func (h *ResourceHandler) GetNamespaceWithDefault(incomingNamespace string) string {
	if incomingNamespace != "" {
		return incomingNamespace
	}

	// 尝试从客户端配置获取当前命名空间
	currentNamespace, err := h.Client.GetCurrentNamespace()
	if err == nil && currentNamespace != "" {
		h.Log.Debug("Using namespace from kubeconfig", "namespace", currentNamespace)
		return currentNamespace
	}

	// 如果客户端配置没有命名空间，使用default
	h.Log.Debug("No namespace provided and none in kubeconfig, using default namespace")
	return "default"
}

// ListResources 实现通用的资源列表功能
func (h *ResourceHandler) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespaceArg, _ := arguments["namespace"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Listing resources",
		"kind", kind,
		"apiVersion", apiVersion,
		"namespace", namespace,
		"labelSelector", labelSelector,
		"group", h.Group,
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

	// 创建列表选项
	listOptions := &clientpkg.ListOptions{Namespace: namespace}
	if labelSelector != "" {
		// 使用 k8s.io/apimachinery/pkg/labels 包创建标签选择器
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			h.Log.Error("Failed to parse label selector",
				"labelSelector", labelSelector,
				"error", err,
			)
			return utils.NewErrorToolResult(fmt.Sprintf("failed to parse label selector: %v", err)), nil
		}

		// 为列表选项设置标签选择器
		listOptions.LabelSelector = selector
	}

	// 列出资源
	err := h.Client.List(ctx, list, listOptions)
	if err != nil {
		h.Log.Error("Failed to list resources",
			"kind", kind,
			"namespace", namespace,
			"labelSelector", labelSelector,
			"error", err,
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to list resources: %v", err)), nil
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d %s resources", len(list.Items), kind))

	if namespace != "" {
		result.WriteString(fmt.Sprintf(" in namespace %s", namespace))
	}

	if labelSelector != "" {
		result.WriteString(fmt.Sprintf(" with label selector '%s'", labelSelector))
	}

	result.WriteString(":\n\n")

	for _, item := range list.Items {
		result.WriteString(fmt.Sprintf("Name: %s\n", item.GetName()))
	}

	h.Log.Info("Resources listed successfully",
		"kind", kind,
		"namespace", namespace,
		"labelSelector", labelSelector,
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

// GetResource 实现通用的资源获取功能
func (h *ResourceHandler) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Getting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
		"group", h.Group,
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
			return utils.NewErrorToolResult(fmt.Sprintf("resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to get resource: %v", err)), nil
	}

	// 转换为YAML
	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		h.Log.Error("Failed to marshal resource to YAML",
			"kind", kind,
			"name", name,
			"error", err,
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to marshal to YAML: %v", err)), nil
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

// DescribeResource 实现通用的资源详细描述功能
func (h *ResourceHandler) DescribeResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Describing resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
		"group", h.Group,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建对象
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	// 获取资源
	err := h.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		h.Log.Error("Failed to get resource for description",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		if errors.IsNotFound(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to describe resource: %v", err)), nil
	}

	// 构建资源描述
	description := models.NewResourceDescriptionFromUnstructured(obj)

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(description, "", "  ")
	if err != nil {
		h.Log.Error("Failed to marshal resource description to JSON",
			"kind", kind,
			"name", name,
			"error", err,
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to marshal to JSON: %v", err)), nil
	}

	h.Log.Info("Resource described successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// CreateResource 实现通用的资源创建功能
func (h *ResourceHandler) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Creating resource from YAML", "group", h.Group)

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to parse YAML: %v", err)), nil
	}

	// 验证资源组
	if !h.validateResourceGroup(obj) {
		return utils.NewErrorToolResult(fmt.Sprintf("invalid resource group: %s, expected: %s", obj.GroupVersionKind().Group, h.Group)), nil
	}

	// 如果命名空间为空，使用default或kubeconfig中的
	if obj.GetNamespace() == "" {
		defaultNs := h.GetNamespaceWithDefault("")
		obj.SetNamespace(defaultNs)
		h.Log.Debug("Empty namespace in resource, setting namespace", "namespace", defaultNs)
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
		return utils.NewErrorToolResult(fmt.Sprintf("failed to create resource: %v", err)), nil
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

// UpdateResource 实现通用的资源更新功能
func (h *ResourceHandler) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Updating resource from YAML", "group", h.Group)

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to parse YAML: %v", err)), nil
	}

	// 验证资源组
	if !h.validateResourceGroup(obj) {
		return utils.NewErrorToolResult(fmt.Sprintf("invalid resource group: %s, expected: %s", obj.GroupVersionKind().Group, h.Group)), nil
	}

	// 如果命名空间为空，使用default或kubeconfig中的
	if obj.GetNamespace() == "" {
		defaultNs := h.GetNamespaceWithDefault("")
		obj.SetNamespace(defaultNs)
		h.Log.Debug("Empty namespace in resource, setting namespace", "namespace", defaultNs)
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
		return utils.NewErrorToolResult(fmt.Sprintf("failed to update resource: %v", err)), nil
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

// DeleteResource 实现通用的资源删除功能
func (h *ResourceHandler) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Deleting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
		"group", h.Group,
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
		if errors.IsNotFound(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to delete resource: %v", err)), nil
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
				Text: fmt.Sprintf("Successfully deleted %s/%s from namespace %s",
					kind, name, namespace),
			},
		},
	}, nil
}

// validateResourceGroup 验证资源是否属于正确的API组
func (h *ResourceHandler) validateResourceGroup(obj *unstructured.Unstructured) bool {
	gvk := obj.GroupVersionKind()
	switch h.Group {
	case interfaces.CoreAPIGroup:
		// 核心API组: pods, services, configmaps, secrets, namespaces等
		return gvk.Group == ""
	case interfaces.AppsAPIGroup:
		// 应用API组: deployments, statefulsets, daemonsets等
		return gvk.Group == "apps"
	case interfaces.BatchAPIGroup:
		// 批处理API组: jobs, cronjobs等
		return gvk.Group == "batch"
	case interfaces.NetworkingAPIGroup:
		// 网络API组: ingresses, networkpolicies等
		return gvk.Group == "networking.k8s.io"
	case interfaces.StorageAPIGroup:
		// 存储API组: storageclasses, volumeattachments等
		return gvk.Group == "storage.k8s.io"
	case interfaces.RbacAPIGroup:
		// RBAC API组: roles, rolebindings, clusterroles, clusterrolebindings等
		return gvk.Group == "rbac.authorization.k8s.io"
	case interfaces.ApiextensionsAPIGroup:
		// API扩展API组: customresourcedefinitions等
		return gvk.Group == "apiextensions.k8s.io"
	case interfaces.PolicyAPIGroup:
		// 策略API组: podsecuritypolicies, poddisruptionbudgets等
		return gvk.Group == "policy"
	case interfaces.AutoscalingAPIGroup:
		// 自动伸缩API组: horizontalpodautoscalers等
		return gvk.Group == "autoscaling"
	default:
		return false
	}
}

// Handle 处理通用资源请求
func (h *ResourceHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prefix := h.resourcePrefix
	switch request.Method {
	case fmt.Sprintf("LIST_%s_RESOURCES", prefix):
		return h.ListResources(ctx, request)
	case fmt.Sprintf("GET_%s_RESOURCE", prefix):
		return h.GetResource(ctx, request)
	case fmt.Sprintf("DESCRIBE_%s_RESOURCE", prefix):
		return h.DescribeResource(ctx, request)
	case fmt.Sprintf("CREATE_%s_RESOURCE", prefix):
		return h.CreateResource(ctx, request)
	case fmt.Sprintf("UPDATE_%s_RESOURCE", prefix):
		return h.UpdateResource(ctx, request)
	case fmt.Sprintf("DELETE_%s_RESOURCE", prefix):
		return h.DeleteResource(ctx, request)
	default:
		return utils.NewErrorToolResult(fmt.Sprintf("unknown %s resource method: %s", strings.ToLower(prefix), request.Method)), nil
	}
}

// GetResourcePrefix 获取资源前缀
func (h *ResourceHandler) GetResourcePrefix() string {
	return h.resourcePrefix
}
