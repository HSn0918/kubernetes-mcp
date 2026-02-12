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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
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
		logger.Any("scope", h.Scope),
		logger.Any("apiGroup", h.Group),
		logger.String("prefix", prefix),
	)
	// 注册列出资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("LIST_%s_RESOURCES", prefix),
		mcp.WithDescription(fmt.Sprintf("列出指定API组的Kubernetes资源（作用域：%s）。支持按命名空间过滤和标签选择器过滤。适用于资源监控、状态检查、依赖分析等场景。返回资源的基本信息列表。注意：在大规模集群中，建议使用标签选择器限制返回数量。", h.Scope)),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'、'Service'等。区分大小写。留空时按发现接口列出当前命名空间可访问的所有资源。"),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，例如：'v1'、'apps/v1'。当kind不为空时建议同时提供，避免同名kind歧义。"),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所在的命名空间。如果是集群级资源则忽略此参数。默认为'default'命名空间。"),
			mcp.DefaultString("default"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按资源属性进行过滤。例如：'status.phase=Running'表示只显示运行中的资源。支持多个条件，使用逗号分隔。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按资源标签进行过滤。例如：'app=nginx'表示只显示带有app=nginx标签的资源。支持多个标签，使用逗号分隔。"),
		),
		mcp.WithBoolean("showLabels",
			mcp.Description("是否显示资源的所有标签。启用后将在输出中包含完整的标签列表，有助于资源分类和管理。默认为false。"),
			mcp.DefaultBool(false),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("GET_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("获取指定API组中的资源详情（作用域：%s）。返回资源的完整定义，包括：元数据、规格配置、状态信息等。适用于资源检查、问题诊断、状态验证等场景。支持查看历史版本（如果启用了资源版本跟踪）。", h.Scope)),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'等。区分大小写，必须是集群中存在的资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，必须与资源类型匹配。例如：'v1'、'apps/v1'等。建议使用最新的稳定版本。"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("资源名称。区分大小写，必须是目标命名空间中存在的资源。支持查询已删除资源的最后状态（如果启用了垃圾回收保护）。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所在的命名空间。如果是集群级资源则忽略此参数。默认为'default'命名空间。"),
			mcp.DefaultString("default"),
		),
	), h.GetResource)

	// 注册描述资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("DESCRIBE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("详细描述指定API组中的资源（作用域：%s）。提供比GET更丰富的信息，包括：事件历史、关联资源、运行状态、配置详情等。适用于深入排查问题、监控资源状态、分析资源关系等场景。自动关联显示相关的事件信息。", h.Scope)),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'等。区分大小写，必须是集群中存在的资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，必须与资源类型匹配。例如：'v1'、'apps/v1'等。建议使用最新的稳定版本。"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("资源名称。区分大小写，必须是目标命名空间中存在的资源。将展示该资源的详细运行状态和历史信息。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所在的命名空间。如果是集群级资源则忽略此参数。默认为'default'命名空间。"),
			mcp.DefaultString("default"),
		),
	), h.DescribeResource)

	// 注册创建资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("CREATE_%s_RESOURCE", prefix),
		mcp.WithDescription("创建新的API资源。支持从YAML定义创建资源，自动处理依赖关系。适用于部署应用、创建配置、初始化资源等场景。创建前会进行资源验证和冲突检查。注意：某些资源可能需要特定的权限才能创建。"),
		mcp.WithString("yaml",
			mcp.Description("资源的YAML定义。必须是有效的Kubernetes资源清单，包含：apiVersion、kind、metadata等必要字段。支持引用ConfigMap和Secret。注意处理敏感信息。"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("UPDATE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("更新指定API组中的资源（作用域：%s）。支持声明式更新，自动处理资源版本冲突。适用于配置变更、规格调整、状态更新等场景。建议先预览变更再应用。", h.Scope)),
		mcp.WithString("yaml",
			mcp.Description("资源的YAML定义。必须是有效的Kubernetes资源清单，包含完整的资源定义。系统会根据资源名称和命名空间查找并更新目标资源。"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(fmt.Sprintf("DELETE_%s_RESOURCE", prefix),
		mcp.WithDescription(fmt.Sprintf("删除指定API组中的资源（作用域：%s）。支持级联删除关联资源。适用于资源清理、环境重置、应用卸载等场景。注意：某些资源可能有终结器（Finalizer）导致删除需要较长时间。", h.Scope)),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'等。区分大小写，必须是集群中存在的资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，必须与资源类型匹配。例如：'v1'、'apps/v1'等。建议使用最新的稳定版本。"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("要删除的资源名称。区分大小写，必须是目标命名空间中存在的资源。删除操作不可逆，请谨慎操作。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所在的命名空间。如果是集群级资源则忽略此参数。默认为'default'命名空间。"),
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
		h.Log.Debug("Using namespace from kubeconfig", logger.String("namespace", currentNamespace))
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
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespaceArg, _ := arguments["namespace"].(string)
	fieldSelector, _ := arguments["fieldSelector"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)
	showLabels, _ := arguments["showLabels"].(bool)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Listing resources",
		logger.String("kind", kind),
		logger.String("apiVersion", apiVersion),
		logger.String("namespace", namespace),
		logger.String("fieldSelector", fieldSelector),
		logger.String("labelSelector", labelSelector),
		logger.Bool("showLabels", showLabels),
		logger.Any("group", h.Group),
	)

	kind = strings.TrimSpace(kind)
	apiVersion = strings.TrimSpace(apiVersion)
	fieldSelector = strings.TrimSpace(fieldSelector)
	labelSelector = strings.TrimSpace(labelSelector)

	// kind为空时，按发现接口遍历所有可列表资源（兼容原有调用方式）
	if kind == "" {
		return h.listResourcesByDiscovery(ctx, namespace, fieldSelector, labelSelector, showLabels)
	}

	if apiVersion == "" {
		return utils.NewErrorToolResult("missing required parameter: apiVersion is required when kind is specified"), nil
	}

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
	if fieldSelector != "" {
		selector, err := fields.ParseSelector(fieldSelector)
		if err != nil {
			h.Log.Error("Failed to parse field selector",
				logger.String("fieldSelector", fieldSelector),
				logger.Any("error", err),
			)
			return utils.NewErrorToolResult(fmt.Sprintf("failed to parse field selector: %v", err)), nil
		}
		listOptions.FieldSelector = selector
	}
	if labelSelector != "" {
		// 使用 k8s.io/apimachinery/pkg/labels 包创建标签选择器
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			h.Log.Error("Failed to parse label selector",
				logger.String("labelSelector", labelSelector),
				logger.Any("error", err),
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
			logger.String("kind", kind),
			logger.String("namespace", namespace),
			logger.String("labelSelector", labelSelector),
			logger.Any("error", err),
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
		if showLabels {
			result.WriteString(fmt.Sprintf("  Labels: %v\n", item.GetLabels()))
		}
	}

	h.Log.Info("Resources listed successfully",
		logger.String("kind", kind),
		logger.String("namespace", namespace),
		logger.String("labelSelector", labelSelector),
		logger.Int("count", len(list.Items)),
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

func (h *ResourceHandler) listResourcesByDiscovery(
	ctx context.Context,
	namespace string,
	fieldSelector string,
	labelSelector string,
	showLabels bool,
) (*mcp.CallToolResult, error) {
	_, resourcesList, err := h.Client.GetDiscoveryClient().ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		h.Log.Error("Failed to discover API resources", logger.Any("error", err))
		return utils.NewErrorToolResult(fmt.Sprintf("failed to discover API resources: %v", err)), nil
	}
	if err != nil {
		h.Log.Warn("Partial API discovery error", logger.Any("error", err))
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
		LabelSelector: labelSelector,
	}

	var (
		totalKinds int
		totalItems int
		result     strings.Builder
	)

	for _, resourceList := range resourcesList {
		groupVersion := resourceList.GroupVersion
		gv, parseErr := schema.ParseGroupVersion(groupVersion)
		if parseErr != nil {
			h.Log.Warn("Skipping invalid groupVersion", logger.String("groupVersion", groupVersion), logger.Any("error", parseErr))
			continue
		}

		for _, resource := range resourceList.APIResources {
			if strings.Contains(resource.Name, "/") || !containsVerb(resource.Verbs, "list") {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}

			var list *unstructured.UnstructuredList
			if resource.Namespaced {
				list, err = h.Client.GetDynamicClient().Resource(gvr).Namespace(namespace).List(ctx, listOptions)
			} else {
				list, err = h.Client.GetDynamicClient().Resource(gvr).List(ctx, listOptions)
			}
			if err != nil {
				continue
			}
			if len(list.Items) == 0 {
				continue
			}

			totalKinds++
			totalItems += len(list.Items)
			scope := "cluster"
			if resource.Namespaced {
				scope = namespace
			}

			result.WriteString(fmt.Sprintf(
				"Kind: %s (apiVersion: %s, resource: %s, scope: %s, count: %d)\n",
				resource.Kind,
				groupVersion,
				resource.Name,
				scope,
				len(list.Items),
			))

			for _, item := range list.Items {
				result.WriteString(fmt.Sprintf("- %s\n", item.GetName()))
				if showLabels {
					result.WriteString(fmt.Sprintf("  Labels: %v\n", item.GetLabels()))
				}
			}
			result.WriteString("\n")
		}
	}

	if totalItems == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Found 0 resources in namespace %s", namespace),
				},
			},
		}, nil
	}

	header := fmt.Sprintf("Found %d resources across %d kinds in namespace %s\n\n", totalItems, totalKinds, namespace)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: header + result.String(),
			},
		},
	}, nil
}

func containsVerb(verbs []string, target string) bool {
	for _, verb := range verbs {
		if verb == target {
			return true
		}
	}
	return false
}

// GetResource 实现通用的资源获取功能
func (h *ResourceHandler) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Getting resource",
		logger.String("kind", kind),
		logger.String("apiVersion", apiVersion),
		logger.String("name", name),
		logger.String("namespace", namespace),
		logger.Any("group", h.Group),
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
			logger.String("kind", kind),
			logger.String("name", name),
			logger.String("namespace", namespace),
			logger.Any("error", err),
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
			logger.String("kind", kind),
			logger.String("name", name),
			logger.Any("error", err),
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to marshal to YAML: %v", err)), nil
	}

	h.Log.Info("Resource retrieved successfully",
		logger.String("kind", kind),
		logger.String("name", name),
		logger.String("namespace", namespace),
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
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Describing resource",
		logger.String("kind", kind),
		logger.String("apiVersion", apiVersion),
		logger.String("name", name),
		logger.String("namespace", namespace),
		logger.Any("group", h.Group),
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
			logger.String("kind", kind),
			logger.String("name", name),
			logger.String("namespace", namespace),
			logger.Any("error", err),
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
			logger.String("kind", kind),
			logger.String("name", name),
			logger.Any("error", err),
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to marshal to JSON: %v", err)), nil
	}

	h.Log.Info("Resource described successfully",
		logger.String("kind", kind),
		logger.String("name", name),
		logger.String("namespace", namespace),
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

// CreateResource 创建资源
func (h *ResourceHandler) CreateResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Log.Info("Creating resource",
		logger.String("method", request.Method),
		logger.Any("handler_group", h.Group),
		logger.String("handler_type", fmt.Sprintf("%T", h)),
	)

	// 解析YAML
	obj := &unstructured.Unstructured{}
	arguments := request.GetArguments()
	yamlStr, _ := arguments["yaml"].(string)
	if err := yaml.Unmarshal([]byte(yamlStr), obj); err != nil {
		h.Log.Error("Failed to parse YAML",
			logger.Any("error", err),
			logger.String("yaml", yamlStr),
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to parse YAML: %v", err)), nil
	}

	// 记录资源信息
	gvk := obj.GroupVersionKind()
	h.Log.Info("Parsed resource",
		logger.String("group", gvk.Group),
		logger.String("version", gvk.Version),
		logger.String("kind", gvk.Kind),
		logger.Any("expected_group", h.Group),
	)

	// 获取命名空间
	if obj.GetNamespace() == "" {
		defaultNs := h.GetNamespaceWithDefault("")
		obj.SetNamespace(defaultNs)
		h.Log.Debug("Empty namespace in resource, setting namespace", logger.String("namespace", defaultNs))
	}

	// 创建资源
	if err := h.Client.Create(ctx, obj); err != nil {
		h.Log.Error("Failed to create resource",
			logger.Any("error", err),
			logger.String("group", gvk.Group),
			logger.String("version", gvk.Version),
			logger.String("kind", gvk.Kind),
			logger.String("namespace", obj.GetNamespace()),
		)
		if errors.IsAlreadyExists(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("resource already exists: %v", err)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to create resource: %v", err)), nil
	}

	h.Log.Info("Resource created successfully",
		logger.String("group", gvk.Group),
		logger.String("version", gvk.Version),
		logger.String("kind", gvk.Kind),
		logger.String("namespace", obj.GetNamespace()),
		logger.String("name", obj.GetName()),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created %s/%s in namespace %s",
					gvk.Kind, obj.GetName(), obj.GetNamespace()),
			},
		},
	}, nil
}

// UpdateResource 实现通用的资源更新功能
func (h *ResourceHandler) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Updating resource from YAML", logger.Any("group", h.Group))

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", logger.Any("error", err))
		return utils.NewErrorToolResult(fmt.Sprintf("failed to parse YAML: %v", err)), nil
	}
	// 如果命名空间为空，使用default或kubeconfig中的
	if obj.GetNamespace() == "" {
		defaultNs := h.GetNamespaceWithDefault("")
		obj.SetNamespace(defaultNs)
		h.Log.Debug("Empty namespace in resource, setting namespace", logger.String("namespace", defaultNs))
	}

	h.Log.Debug("Parsed resource from YAML",
		logger.String("kind", obj.GetKind()),
		logger.String("name", obj.GetName()),
		logger.String("namespace", obj.GetNamespace()),
	)

	// 更新资源
	err = h.Client.Update(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to update resource",
			logger.String("kind", obj.GetKind()),
			logger.String("name", obj.GetName()),
			logger.String("namespace", obj.GetNamespace()),
			logger.Any("error", err),
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to update resource: %v", err)), nil
	}

	h.Log.Info("Resource updated successfully",
		logger.String("kind", obj.GetKind()),
		logger.String("name", obj.GetName()),
		logger.String("namespace", obj.GetNamespace()),
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
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.GetNamespaceWithDefault(namespaceArg)

	h.Log.Info("Deleting resource",
		logger.String("kind", kind),
		logger.String("apiVersion", apiVersion),
		logger.String("name", name),
		logger.String("namespace", namespace),
		logger.Any("group", h.Group),
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
			logger.String("kind", kind),
			logger.String("name", name),
			logger.String("namespace", namespace),
			logger.Any("error", err),
		)
		if errors.IsNotFound(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to delete resource: %v", err)), nil
	}

	h.Log.Info("Resource deleted successfully",
		logger.String("kind", kind),
		logger.String("name", name),
		logger.String("namespace", namespace),
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
