package base

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
)

// 定义工具常量
const (
	// 通用工具方法
	GET_CLUSTER_INFO  = "GET_CLUSTER_INFO"
	GET_API_RESOURCES = "GET_API_RESOURCES"
	SEARCH_RESOURCES  = "SEARCH_RESOURCES"
	EXPLAIN_RESOURCE  = "EXPLAIN_RESOURCE"
	APPLY_MANIFEST    = "APPLY_MANIFEST"
	VALIDATE_MANIFEST = "VALIDATE_MANIFEST"
	DIFF_MANIFEST     = "DIFF_MANIFEST"
	GET_EVENTS        = "GET_EVENTS"
)

// UtilityHandler 提供通用工具功能
type UtilityHandler struct {
	Handler
}

// 确保实现了接口
var _ interfaces.ToolHandler = (*UtilityHandler)(nil)

// NewUtilityHandler 创建新的通用工具处理程序
func NewUtilityHandler(client client.KubernetesClient) interfaces.ToolHandler {
	return &UtilityHandler{
		Handler: NewBaseHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Register 注册通用工具方法
func (h *UtilityHandler) Register(server *server.MCPServer) {
	h.Log.Info("Registering utility handlers")

	// 获取集群信息工具
	server.AddTool(mcp.NewTool(GET_CLUSTER_INFO,
		mcp.WithDescription("Get Kubernetes cluster information"),
	), h.GetClusterInfo)

	// 获取API资源工具
	server.AddTool(mcp.NewTool(GET_API_RESOURCES,
		mcp.WithDescription("Get available API resources in the cluster"),
		mcp.WithString("group",
			mcp.Description("API Group (optional)"),
		),
	), h.GetAPIResources)

	// 搜索资源工具
	server.AddTool(mcp.NewTool(SEARCH_RESOURCES,
		mcp.WithDescription("Search resources across the cluster"),
		mcp.WithString("query",
			mcp.Description("Search query (name, label, annotation pattern)"),
			mcp.Required(),
		),
		mcp.WithString("namespaces",
			mcp.Description("Comma-separated list of namespaces to search (default: all)"),
		),
		mcp.WithString("kinds",
			mcp.Description("Comma-separated list of resource kinds to search (default: all)"),
		),
		mcp.WithBoolean("matchLabels",
			mcp.Description("Whether to match labels in search"),
			mcp.DefaultBool(true),
		),
		mcp.WithBoolean("matchAnnotations",
			mcp.Description("Whether to match annotations in search"),
			mcp.DefaultBool(true),
		),
	), h.SearchResources)

	// 解释资源结构工具
	server.AddTool(mcp.NewTool(EXPLAIN_RESOURCE,
		mcp.WithDescription("Explain resource structure"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version"),
			mcp.Required(),
		),
		mcp.WithString("field",
			mcp.Description("Specific field path to explain (e.g. 'spec.containers')"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Explain fields recursively"),
			mcp.DefaultBool(false),
		),
	), h.ExplainResource)

	// 应用清单工具
	server.AddTool(mcp.NewTool(APPLY_MANIFEST,
		mcp.WithDescription("Apply Kubernetes manifest(s)"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest(s) to apply"),
			mcp.Required(),
		),
		mcp.WithBoolean("dryRun",
			mcp.Description("Perform a dry run without making changes"),
			mcp.DefaultBool(false),
		),
		mcp.WithString("fieldManager",
			mcp.Description("Name of the field manager"),
			mcp.DefaultString("kubernetes-mcp"),
		),
	), h.ApplyManifest)

	// 验证清单工具
	server.AddTool(mcp.NewTool(VALIDATE_MANIFEST,
		mcp.WithDescription("Validate Kubernetes manifest(s)"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest(s) to validate"),
			mcp.Required(),
		),
	), h.ValidateManifest)

	// 比较清单工具
	server.AddTool(mcp.NewTool(DIFF_MANIFEST,
		mcp.WithDescription("Compare manifest with live resource"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest to compare"),
			mcp.Required(),
		),
	), h.DiffManifest)

	// 获取事件工具
	server.AddTool(mcp.NewTool(GET_EVENTS,
		mcp.WithDescription("Get events for a resource"),
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
	), h.GetEvents)
}

// Handle 实现接口方法
func (h *UtilityHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case GET_CLUSTER_INFO:
		return h.GetClusterInfo(ctx, request)
	case GET_API_RESOURCES:
		return h.GetAPIResources(ctx, request)
	case SEARCH_RESOURCES:
		return h.SearchResources(ctx, request)
	case EXPLAIN_RESOURCE:
		return h.ExplainResource(ctx, request)
	case APPLY_MANIFEST:
		return h.ApplyManifest(ctx, request)
	case VALIDATE_MANIFEST:
		return h.ValidateManifest(ctx, request)
	case DIFF_MANIFEST:
		return h.DiffManifest(ctx, request)
	case GET_EVENTS:
		return h.GetEvents(ctx, request)
	default:
		return nil, fmt.Errorf("unknown utility method: %s", request.Method)
	}
}

// GetClusterInfo 获取集群信息
func (h *UtilityHandler) GetClusterInfo(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Getting cluster info")

	// 构建响应
	var result strings.Builder
	result.WriteString("Kubernetes Cluster Information:\n")

	// 注意：当前未实现获取集群版本的方法，这里仅作为框架
	// TODO: 实现 KubernetesClient.GetServerVersion()
	result.WriteString("Version: (implementation needed)\n")
	result.WriteString("Build Date: (implementation needed)\n")
	result.WriteString("Go Version: (implementation needed)\n")
	result.WriteString("Platform: (implementation needed)\n")

	// TODO: 添加更多集群信息到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// GetAPIResources 获取API资源列表
func (h *UtilityHandler) GetAPIResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	group, _ := arguments["group"].(string)

	h.Log.Info("Getting API resources", "group", group)

	// TODO: 实现获取API资源列表的逻辑

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("API Resources:\n"))
	// TODO: 添加API资源信息到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// SearchResources 搜索资源
func (h *UtilityHandler) SearchResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	query, _ := arguments["query"].(string)
	namespacesStr, _ := arguments["namespaces"].(string)
	kindsStr, _ := arguments["kinds"].(string)
	matchLabels, _ := arguments["matchLabels"].(bool)
	matchAnnotations, _ := arguments["matchAnnotations"].(bool)

	h.Log.Info("Searching resources",
		"query", query,
		"namespaces", namespacesStr,
		"kinds", kindsStr,
		"matchLabels", matchLabels,
		"matchAnnotations", matchAnnotations,
	)

	// TODO: 实现资源搜索逻辑

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search Results for '%s':\n\n", query))
	// TODO: 添加搜索结果到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// ExplainResource 解释资源结构
func (h *UtilityHandler) ExplainResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	field, _ := arguments["field"].(string)
	recursive, _ := arguments["recursive"].(bool)

	h.Log.Info("Explaining resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"field", field,
		"recursive", recursive,
	)

	// TODO: 实现资源结构解释逻辑

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Resource Structure for %s (%s):\n\n", kind, apiVersion))
	// TODO: 添加资源结构说明到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// ApplyManifest 应用资源清单
func (h *UtilityHandler) ApplyManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)
	dryRun, _ := arguments["dryRun"].(bool)
	fieldManager, _ := arguments["fieldManager"].(string)

	h.Log.Info("Applying manifest",
		"dryRun", dryRun,
		"fieldManager", fieldManager,
	)

	// TODO: 实现应用资源清单的逻辑
	_ = yamlStr // 标记变量为已使用，防止编译错误

	// 构建响应
	var result strings.Builder
	if dryRun {
		result.WriteString("Dry Run: Resources that would be applied:\n\n")
	} else {
		result.WriteString("Applied Resources:\n\n")
	}
	// TODO: 添加应用结果到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// ValidateManifest 验证资源清单
func (h *UtilityHandler) ValidateManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Validating manifest")

	// TODO: 实现验证资源清单的逻辑
	_ = yamlStr // 标记变量为已使用，防止编译错误

	// 构建响应
	var result strings.Builder
	result.WriteString("Validation Results:\n\n")
	// TODO: 添加验证结果到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// DiffManifest 比较资源清单与集群中的资源
func (h *UtilityHandler) DiffManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Diffing manifest")

	// TODO: 实现比较资源清单的逻辑
	_ = yamlStr // 标记变量为已使用，防止编译错误

	// 构建响应
	var result strings.Builder
	result.WriteString("Diff Results:\n\n")
	// TODO: 添加比较结果到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// GetEvents 获取资源的事件
func (h *UtilityHandler) GetEvents(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间
	namespace := namespaceArg
	if namespace == "" {
		namespace = "default"
	}

	h.Log.Info("Getting resource events",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	// TODO: 实现获取资源事件的逻辑

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Events for %s/%s in namespace %s:\n\n", kind, name, namespace))
	// TODO: 添加事件列表到结果中

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}
