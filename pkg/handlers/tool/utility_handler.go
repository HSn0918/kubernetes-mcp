package tool

import (
	"context"
	"fmt"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
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
	base.Handler
}

// 确保实现了接口
var _ interfaces.ToolHandler = (*UtilityHandler)(nil)

// NewUtilityHandler 创建新的通用工具处理程序
func NewUtilityHandler(client kubernetes.Client) interfaces.ToolHandler {
	return &UtilityHandler{
		Handler: base.NewHandler(client, interfaces.ClusterScope, interfaces.Tool),
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
		return utils.NewErrorToolResult(fmt.Sprintf("unknown utility method: %s", request.Method)), nil
	}
}
