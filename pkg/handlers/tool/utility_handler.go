package tool

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

// 定义工具常量
const (
	// 通用工具方法
	GET_CURRENT_TIME  = "GET_CURRENT_TIME"
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
	// 获取当前时间工具
	server.AddTool(mcp.NewTool(GET_CURRENT_TIME,
		mcp.WithDescription("获取系统当前时间。用于同步集群操作时间戳，确保操作记录的准确性。常用于日志记录、资源创建时间标记等场景。返回格式：RFC3339标准时间格式。"),
	), h.GetCurrentTime)
	// 获取集群信息工具
	server.AddTool(mcp.NewTool(GET_CLUSTER_INFO,
		mcp.WithDescription("获取Kubernetes集群详细信息。包括：集群版本、节点数量、命名空间列表、API Server地址等核心信息。用于集群状态检查、版本兼容性验证、集群资源概览等场景。建议在执行关键操作前先检查集群状态。"),
	), h.GetClusterInfo)

	// 获取API资源工具
	server.AddTool(mcp.NewTool(GET_API_RESOURCES,
		mcp.WithDescription("获取集群中可用的API资源列表。可选择性地按API组过滤。返回资源的版本、种类、是否支持命名空间等信息。用于资源操作前的权限检查、API版本验证、自定义资源发现等场景。注意：某些资源可能需要特定的访问权限。"),
		mcp.WithString("group",
			mcp.Description("API组名称，例如：'apps'、'batch'等。留空则返回所有API组的资源。"),
		),
	), h.GetAPIResources)

	// 搜索资源工具
	server.AddTool(mcp.NewTool(SEARCH_RESOURCES,
		mcp.WithDescription("跨集群资源搜索工具。支持按名称、标签、注解进行模糊匹配。可指定搜索范围（命名空间）和资源类型。适用于资源定位、依赖分析、状态检查等场景。支持通配符匹配，例如：'app=nginx-*'。注意：大规模搜索可能影响性能。"),
		mcp.WithString("query",
			mcp.Description("搜索条件，支持以下格式：\n- 名称匹配：'name=nginx'\n- 标签匹配：'label=app:nginx'\n- 注解匹配：'annotation=deployment.kubernetes.io/revision:1'\n支持通配符：'*'"),
			mcp.Required(),
		),
		mcp.WithString("namespaces",
			mcp.Description("要搜索的命名空间列表，多个用逗号分隔。例如：'default,kube-system'。留空表示搜索所有命名空间。"),
		),
		mcp.WithString("kinds",
			mcp.Description("要搜索的资源类型列表，多个用逗号分隔。例如：'pods,deployments'。留空表示搜索所有类型。建议指定以提高搜索效率。"),
		),
		mcp.WithBoolean("matchLabels",
			mcp.Description("是否匹配标签。启用后将检查资源的所有标签。可能增加搜索时间。"),
			mcp.DefaultBool(true),
		),
		mcp.WithBoolean("matchAnnotations",
			mcp.Description("是否匹配注解。启用后将检查资源的所有注解。可能增加搜索时间。"),
			mcp.DefaultBool(true),
		),
	), h.SearchResources)

	// 解释资源结构工具
	server.AddTool(mcp.NewTool(EXPLAIN_RESOURCE,
		mcp.WithDescription("解释Kubernetes资源结构。提供资源定义的详细说明，包括字段含义、类型、是否必填等信息。支持递归解释嵌套字段。适用于资源配置编写、字段验证、API兼容性检查等场景。可用于学习和理解Kubernetes API结构。"),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'、'Service'等。区分大小写。"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，例如：'v1'、'apps/v1'、'networking.k8s.io/v1'等。必须与kind匹配。"),
			mcp.Required(),
		),
		mcp.WithString("field",
			mcp.Description("要解释的具体字段路径，使用点号分隔。例如：'spec.containers'、'spec.template.spec'等。留空则解释整个资源结构。"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("是否递归解释字段。启用后将显示所有子字段的详细信息。可能产生大量输出。"),
			mcp.DefaultBool(false),
		),
	), h.ExplainResource)

	// 应用清单工具
	server.AddTool(mcp.NewTool(APPLY_MANIFEST,
		mcp.WithDescription("应用Kubernetes资源清单。支持创建、更新操作，采用声明式API。可处理单个或多个资源清单。支持dry-run模式进行预检查。使用server-side apply确保安全的多方协作。适用于资源部署、配置更新、状态管理等场景。"),
		mcp.WithString("yaml",
			mcp.Description("YAML格式的资源清单。支持多文档语法（使用'---'分隔）。必须是有效的Kubernetes资源定义。"),
			mcp.Required(),
		),
		mcp.WithBoolean("dryRun",
			mcp.Description("是否执行试运行。启用后只验证和模拟执行，不实际修改集群状态。建议在重要操作前先进行试运行。"),
			mcp.DefaultBool(false),
		),
		mcp.WithString("fieldManager",
			mcp.Description("字段管理器名称，用于跟踪字段所有权。在多方管理同一资源时很重要。建议使用有意义的名称以便跟踪。"),
			mcp.DefaultString("kubernetes-mcp"),
		),
	), h.ApplyManifest)

	// 验证清单工具
	server.AddTool(mcp.NewTool(VALIDATE_MANIFEST,
		mcp.WithDescription("验证Kubernetes资源清单的合法性。检查包括：语法正确性、必填字段、字段类型、API版本兼容性等。支持验证单个或多个资源清单。适用于部署前的配置检查、CI/CD流程中的质量控制等场景。及早发现配置错误，避免部署失败。"),
		mcp.WithString("yaml",
			mcp.Description("要验证的YAML格式资源清单。支持多文档语法。将进行完整的结构和语义验证。"),
			mcp.Required(),
		),
	), h.ValidateManifest)

	// 比较清单工具
	server.AddTool(mcp.NewTool(DIFF_MANIFEST,
		mcp.WithDescription("比较清单与集群中现有资源的差异。显示详细的字段级别差异，包括新增、修改、删除的配置。支持比较复杂的嵌套结构。适用于配置更新前的影响分析、变更审计、配置偏差检测等场景。帮助理解变更范围和潜在影响。"),
		mcp.WithString("yaml",
			mcp.Description("要比较的YAML格式资源清单。将与集群中的同名资源进行比较。必须包含资源的名称和命名空间信息。"),
			mcp.Required(),
		),
	), h.DiffManifest)

	// 获取事件工具
	server.AddTool(mcp.NewTool(GET_EVENTS,
		mcp.WithDescription("获取特定资源相关的事件信息。包括：警告、错误、状态变更等事件。支持按时间范围和事件类型过滤。适用于问题诊断、状态监控、变更追踪等场景。帮助理解资源的生命周期和运行状态。注意：事件默认保留时间有限。"),
		mcp.WithString("kind",
			mcp.Description("资源类型，例如：'Pod'、'Deployment'等。必须是集群中存在的资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API版本，必须与资源类型匹配。例如：'v1'、'apps/v1'等。"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("资源名称。区分大小写，必须是目标命名空间中存在的资源。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所在的命名空间。如果资源类型是集群级别的，此参数将被忽略。"),
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
