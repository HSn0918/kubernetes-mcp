package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
)

// 提示词类型常量
const (
	KUBERNETES_YAML_PROMPT    = "KUBERNETES_YAML_PROMPT"
	KUBERNETES_QUERY_PROMPT   = "KUBERNETES_QUERY_PROMPT"
	TROUBLESHOOT_PODS_PROMPT  = "TROUBLESHOOT_PODS_PROMPT"
	TROUBLESHOOT_NODES_PROMPT = "TROUBLESHOOT_NODES_PROMPT"
	TROUBLESHOOT_NET_PROMPT   = "TROUBLESHOOT_NETWORK_PROMPT"
)

// PromptHandler 处理Kubernetes相关提示词
type PromptHandler struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.ToolHandler = (*PromptHandler)(nil)

// NewPromptHandler 创建新的提示词处理程序
func NewPromptHandler(client kubernetes.Client) *PromptHandler {
	return &PromptHandler{
		Handler: base.NewHandler(client, interfaces.ClusterScope, interfaces.Prompt),
	}
}

// Handle 处理工具请求
func (h *PromptHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Log.Info("Handle called for prompt handler, method: ", request.Method)

	// 根据方法名称分发到相应的处理函数
	switch request.Method {
	case KUBERNETES_YAML_PROMPT:
		return h.handleKubernetesYAMLPrompt(ctx, request)
	case KUBERNETES_QUERY_PROMPT:
		return h.handleKubernetesQueryPrompt(ctx, request)
	case TROUBLESHOOT_PODS_PROMPT:
		return h.handleTroubleshootPodsPrompt(ctx, request)
	case TROUBLESHOOT_NODES_PROMPT:
		return h.handleTroubleshootNodesPrompt(ctx, request)
	case TROUBLESHOOT_NET_PROMPT:
		return h.handleTroubleshootNetworkPrompt(ctx, request)
	default:
		return nil, nil
	}
}

// handleKubernetesYAMLPrompt 处理 Kubernetes YAML 生成提示词工具请求
func (h *PromptHandler) handleKubernetesYAMLPrompt(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用models中定义的模板
	template := models.KubernetesYAMLPrompt

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建带有JSON的响应文本
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes YAML 生成提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: promptText.String(),
			},
		},
	}, nil
}

// handleKubernetesQueryPrompt 处理 Kubernetes 查询提示词工具请求
func (h *PromptHandler) handleKubernetesQueryPrompt(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用models中定义的模板
	template := models.KubernetesQueryPrompt

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建带有JSON的响应文本
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes操作指导提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: promptText.String(),
			},
		},
	}, nil
}

// handleTroubleshootPodsPrompt 处理 Pod 问题排查提示词工具请求
func (h *PromptHandler) handleTroubleshootPodsPrompt(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用models中定义的模板
	template := models.TroubleshootPodsPrompt

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建带有JSON的响应文本
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes Pod问题排查提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: promptText.String(),
			},
		},
	}, nil
}

// handleTroubleshootNodesPrompt 处理节点问题排查提示词工具请求
func (h *PromptHandler) handleTroubleshootNodesPrompt(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用models中定义的模板
	template := models.TroubleshootNodesPrompt

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建带有JSON的响应文本
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes节点问题排查提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: promptText.String(),
			},
		},
	}, nil
}

// handleTroubleshootNetworkPrompt 处理网络问题排查提示词工具请求
func (h *PromptHandler) handleTroubleshootNetworkPrompt(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用models中定义的模板
	template := models.TroubleshootNetworkPrompt

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建带有JSON的响应文本
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes网络问题排查提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: promptText.String(),
			},
		},
	}, nil
}

// Register 注册提示词到MCP服务器
func (h *PromptHandler) Register(s *server.MCPServer) {
	h.Log.Info("Registering prompt handlers")

	// Kubernetes YAML 生成提示词
	s.AddPrompt(mcp.NewPrompt(KUBERNETES_YAML_PROMPT,
		mcp.WithPromptDescription("生成标准的Kubernetes YAML资源清单。支持常见资源类型的配置生成，包括必要的元数据、规格定义和状态字段。可用于快速创建新资源或作为已有资源的模板。生成的YAML符合Kubernetes最佳实践规范。"),
		mcp.WithArgument("resource_type",
			mcp.ArgumentDescription("要生成的资源类型。支持所有标准Kubernetes资源，例如：\n- 工作负载：Deployment、StatefulSet、DaemonSet、Job、CronJob\n- 服务发现：Service、Ingress\n- 配置与存储：ConfigMap、Secret、PersistentVolumeClaim\n- 安全相关：ServiceAccount、Role、RoleBinding\n注意：区分大小写，必须使用正确的资源类型名称。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("name",
			mcp.ArgumentDescription("资源的名称。要求：\n- 符合DNS子域名规范（小写字母、数字、中划线）\n- 最长63个字符\n- 在同一命名空间中唯一\n建议使用有意义的描述性名称，便于识别和管理。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("资源所属的命名空间。如果不指定，将使用'default'命名空间。建议：\n- 使用命名空间进行资源隔离\n- 遵循团队的命名空间命名规范\n- 注意某些资源（如Node、PV）是集群级别的，不需要命名空间"),
		),
	), h.KubernetesYAMLPrompt)

	// 同时将YAML提示词作为工具注册
	s.AddTool(mcp.NewTool(KUBERNETES_YAML_PROMPT,
		mcp.WithDescription("生成标准的Kubernetes YAML资源清单。支持常见资源类型的配置生成，包括必要的元数据、规格定义和状态字段。可用于快速创建新资源或作为已有资源的模板。生成的YAML符合Kubernetes最佳实践规范。"),
		mcp.WithString("resource_type",
			mcp.Description("要生成的资源类型。支持所有标准Kubernetes资源，例如：\n- 工作负载：Deployment、StatefulSet、DaemonSet、Job、CronJob\n- 服务发现：Service、Ingress\n- 配置与存储：ConfigMap、Secret、PersistentVolumeClaim\n- 安全相关：ServiceAccount、Role、RoleBinding\n注意：区分大小写，必须使用正确的资源类型名称。"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("资源的名称。要求：\n- 符合DNS子域名规范（小写字母、数字、中划线）\n- 最长63个字符\n- 在同一命名空间中唯一\n建议使用有意义的描述性名称，便于识别和管理。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("资源所属的命名空间。如果不指定，将使用'default'命名空间。建议：\n- 使用命名空间进行资源隔离\n- 遵循团队的命名空间命名规范\n- 注意某些资源（如Node、PV）是集群级别的，不需要命名空间"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleKubernetesYAMLPrompt(ctx, request)
	})

	// Kubernetes查询提示词
	s.AddPrompt(mcp.NewPrompt(KUBERNETES_QUERY_PROMPT,
		mcp.WithPromptDescription("提供详细的Kubernetes操作指导。基于任务描述和上下文信息，生成具体的操作步骤、命令示例和最佳实践建议。包括问题诊断、资源管理、配置优化等各个方面的指导。"),
		mcp.WithArgument("task",
			mcp.ArgumentDescription("需要执行的具体任务描述。建议包含：\n- 具体目标（如：扩展部署副本数、更新容器镜像）\n- 相关资源（如：具体的Deployment名称、Service名称）\n- 特殊要求（如：零停机时间、资源限制）\n- 操作环境（如：生产环境、测试环境）\n越详细的描述将获得越精准的指导。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("context",
			mcp.ArgumentDescription("补充的上下文信息，可以包括：\n- 集群版本和特性\n- 现有配置和限制\n- 相关的依赖服务\n- 历史操作记录\n- 团队规范和要求\n这些信息有助于生成更符合实际情况的操作指导。"),
		),
	), h.KubernetesQueryPrompt)

	// 同时将查询提示词作为工具注册
	s.AddTool(mcp.NewTool(KUBERNETES_QUERY_PROMPT,
		mcp.WithDescription("提供详细的Kubernetes操作指导。基于任务描述和上下文信息，生成具体的操作步骤、命令示例和最佳实践建议。包括问题诊断、资源管理、配置优化等各个方面的指导。"),
		mcp.WithString("task",
			mcp.Description("需要执行的具体任务描述。建议包含：\n- 具体目标（如：扩展部署副本数、更新容器镜像）\n- 相关资源（如：具体的Deployment名称、Service名称）\n- 特殊要求（如：零停机时间、资源限制）\n- 操作环境（如：生产环境、测试环境）\n越详细的描述将获得越精准的指导。"),
			mcp.Required(),
		),
		mcp.WithString("context",
			mcp.Description("补充的上下文信息，可以包括：\n- 集群版本和特性\n- 现有配置和限制\n- 相关的依赖服务\n- 历史操作记录\n- 团队规范和要求\n这些信息有助于生成更符合实际情况的操作指导。"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleKubernetesQueryPrompt(ctx, request)
	})

	// Pod问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_PODS_PROMPT,
		mcp.WithPromptDescription("针对Kubernetes Pod问题的系统化排查指南。基于Pod状态和日志信息，提供详细的问题分析和解决方案。包括常见问题的诊断流程、排查命令和修复建议。支持处理容器启动、运行、健康检查等各个阶段的问题。"),
		mcp.WithArgument("pod_status",
			mcp.ArgumentDescription("Pod的当前状态。常见状态包括：\n- CrashLoopBackOff：容器反复崩溃\n- ImagePullBackOff：镜像拉取失败\n- Pending：等待调度或资源\n- Error：容器异常退出\n- ContainerCreating：容器创建中\n- RunContainerError：容器启动失败\n准确的状态信息对诊断问题至关重要。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("pod_logs",
			mcp.ArgumentDescription("Pod的日志内容。建议包含：\n- 容器的标准输出和错误输出\n- 最近的错误信息\n- 关键的应用日志\n- 系统级别的警告或错误\n- 初始化容器的日志（如果有）\n详细的日志有助于准确定位问题原因。"),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("Pod所在的命名空间。这将帮助：\n- 确定资源访问权限\n- 检查命名空间级别的配置\n- 排查网络策略问题\n- 分析资源配额影响"),
		),
	), h.TroubleshootPodsPrompt)

	// 同时将Pod问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_PODS_PROMPT,
		mcp.WithDescription("针对Kubernetes Pod问题的系统化排查指南。基于Pod状态和日志信息，提供详细的问题分析和解决方案。包括常见问题的诊断流程、排查命令和修复建议。支持处理容器启动、运行、健康检查等各个阶段的问题。"),
		mcp.WithString("pod_status",
			mcp.Description("Pod的当前状态。常见状态包括：\n- CrashLoopBackOff：容器反复崩溃\n- ImagePullBackOff：镜像拉取失败\n- Pending：等待调度或资源\n- Error：容器异常退出\n- ContainerCreating：容器创建中\n- RunContainerError：容器启动失败\n准确的状态信息对诊断问题至关重要。"),
			mcp.Required(),
		),
		mcp.WithString("pod_logs",
			mcp.Description("Pod的日志内容。建议包含：\n- 容器的标准输出和错误输出\n- 最近的错误信息\n- 关键的应用日志\n- 系统级别的警告或错误\n- 初始化容器的日志（如果有）\n详细的日志有助于准确定位问题原因。"),
		),
		mcp.WithString("namespace",
			mcp.Description("Pod所在的命名空间。这将帮助：\n- 确定资源访问权限\n- 检查命名空间级别的配置\n- 排查网络策略问题\n- 分析资源配额影响"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootPodsPrompt(ctx, request)
	})

	// 节点问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_NODES_PROMPT,
		mcp.WithPromptDescription("提供全面的Kubernetes节点问题排查指南。基于节点状态和条件信息，分析节点层面的问题，包括资源压力、系统故障、网络异常等。提供系统化的诊断步骤和解决方案。"),
		mcp.WithArgument("node_status",
			mcp.ArgumentDescription("节点的当前状态。典型状态包括：\n- Ready：节点正常运行\n- NotReady：节点异常\n- MemoryPressure：内存压力\n- DiskPressure：磁盘压力\n- NetworkUnavailable：网络异常\n- PIDPressure：进程数量压力\n状态信息反映了节点的健康状况和可用性。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("node_conditions",
			mcp.ArgumentDescription("节点的详细状况信息。建议包含：\n- 各个条件的状态（True/False/Unknown）\n- 最后一次转换时间\n- 状态持续时间\n- 具体的错误信息或警告\n- 系统资源使用情况\n这些信息有助于深入分析节点问题。"),
		),
	), h.TroubleshootNodesPrompt)

	// 同时将节点问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_NODES_PROMPT,
		mcp.WithDescription("提供全面的Kubernetes节点问题排查指南。基于节点状态和条件信息，分析节点层面的问题，包括资源压力、系统故障、网络异常等。提供系统化的诊断步骤和解决方案。"),
		mcp.WithString("node_status",
			mcp.Description("节点的当前状态。典型状态包括：\n- Ready：节点正常运行\n- NotReady：节点异常\n- MemoryPressure：内存压力\n- DiskPressure：磁盘压力\n- NetworkUnavailable：网络异常\n- PIDPressure：进程数量压力\n状态信息反映了节点的健康状况和可用性。"),
			mcp.Required(),
		),
		mcp.WithString("node_conditions",
			mcp.Description("节点的详细状况信息。建议包含：\n- 各个条件的状态（True/False/Unknown）\n- 最后一次转换时间\n- 状态持续时间\n- 具体的错误信息或警告\n- 系统资源使用情况\n这些信息有助于深入分析节点问题。"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootNodesPrompt(ctx, request)
	})

	// 网络问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_NET_PROMPT,
		mcp.WithPromptDescription("专门针对Kubernetes集群网络问题的排查指南。涵盖服务发现、DNS解析、网络策略、负载均衡等各个网络组件的问题诊断。提供系统化的网络故障排除流程和解决方案。"),
		mcp.WithArgument("problem_type",
			mcp.ArgumentDescription("网络问题的具体类型。常见问题包括：\n- 服务不可达：Service访问失败\n- DNS解析失败：无法解析服务名称\n- Ingress异常：外部访问问题\n- 网络策略问题：Pod间通信受阻\n- 跨节点通信故障：节点间网络异常\n- 负载均衡问题：流量分发异常\n准确的问题类型有助于快速定位故障。"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("service_name",
			mcp.ArgumentDescription("相关服务的名称。建议提供：\n- 完整的服务名称\n- 服务类型（ClusterIP/NodePort/LoadBalancer）\n- 端口信息\n- 选择器标签\n这些信息有助于理解服务配置和网络拓扑。"),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("服务所在的命名空间。这将帮助：\n- 确定网络策略范围\n- 检查DNS解析配置\n- 分析跨命名空间通信\n- 排查服务发现问题"),
		),
	), h.TroubleshootNetworkPrompt)

	// 同时将网络问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_NET_PROMPT,
		mcp.WithDescription("专门针对Kubernetes集群网络问题的排查指南。涵盖服务发现、DNS解析、网络策略、负载均衡等各个网络组件的问题诊断。提供系统化的网络故障排除流程和解决方案。"),
		mcp.WithString("problem_type",
			mcp.Description("网络问题的具体类型。常见问题包括：\n- 服务不可达：Service访问失败\n- DNS解析失败：无法解析服务名称\n- Ingress异常：外部访问问题\n- 网络策略问题：Pod间通信受阻\n- 跨节点通信故障：节点间网络异常\n- 负载均衡问题：流量分发异常\n准确的问题类型有助于快速定位故障。"),
			mcp.Required(),
		),
		mcp.WithString("service_name",
			mcp.Description("相关服务的名称。建议提供：\n- 完整的服务名称\n- 服务类型（ClusterIP/NodePort/LoadBalancer）\n- 端口信息\n- 选择器标签\n这些信息有助于理解服务配置和网络拓扑。"),
		),
		mcp.WithString("namespace",
			mcp.Description("服务所在的命名空间。这将帮助：\n- 确定网络策略范围\n- 检查DNS解析配置\n- 分析跨命名空间通信\n- 排查服务发现问题"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootNetworkPrompt(ctx, request)
	})
}

// KubernetesYAMLPrompt 处理 Kubernetes YAML 生成提示词
func (h *PromptHandler) KubernetesYAMLPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("生成Kubernetes YAML提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes YAML 生成",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位资深的Kubernetes YAML配置专家，具有丰富的容器编排和资源配置经验。你的职责是：\n\n1. 生成符合Kubernetes最佳实践的YAML资源清单\n2. 确保配置的安全性、可维护性和可扩展性\n3. 遵循以下原则：\n   - 资源限制和请求的合理设置\n   - 适当的健康检查和就绪性探针\n   - 合理的标签和注解管理\n   - 安全上下文的正确配置\n   - 版本控制和回滚策略\n   - 资源命名的规范性\n\n在生成配置时，你应该：\n- 根据用户需求选择合适的API版本\n- 添加必要的注释说明\n- 提供配置参数的合理默认值\n- 说明关键配置项的作用\n- 注意配置的向后兼容性"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我需要一个Kubernetes资源清单，请根据我提供的要求生成符合最佳实践的YAML配置。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你创建一个完整的YAML配置。为了生成最适合你需求的配置，请提供以下信息：\n\n1. 资源类型（例如：Deployment、Service等）\n2. 资源名称和命名空间\n3. 具体的配置需求：\n   - 容器镜像和版本\n   - 资源限制和请求\n   - 端口配置\n   - 环境变量\n   - 持久化需求\n   - 特殊的安全要求\n   - 其他特定需求\n\n我会根据你的需求生成配置，并确保：\n- 符合Kubernetes最佳实践\n- 包含必要的安全设置\n- 提供合适的资源限制\n- 添加适当的标签和注解\n- 配置合理的健康检查"),
			),
		},
	), nil
}

// KubernetesQueryPrompt 处理 Kubernetes 查询提示词
func (h *PromptHandler) KubernetesQueryPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("生成Kubernetes操作指导提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes操作指导",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位经验丰富的Kubernetes运维专家，精通集群运维和问题处理。你的职责是：\n\n1. 提供清晰、准确的Kubernetes操作指导\n2. 确保操作的安全性和可控性\n3. 遵循以下原则：\n   - 优先考虑操作的安全性\n   - 提供详细的步骤说明\n   - 说明每个操作的影响\n   - 包含必要的验证步骤\n   - 提供回滚方案\n\n在提供指导时，你应该：\n- 理解用户的具体需求\n- 评估操作风险\n- 提供最佳实践建议\n- 说明注意事项\n- 包含故障排除建议"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我需要在Kubernetes集群中执行操作，请提供专业的指导。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会为你提供详细的操作指导。为了给出最准确的建议，请提供以下信息：\n\n1. 具体的操作目标\n2. 操作环境信息：\n   - 集群版本\n   - 当前状态\n   - 相关资源信息\n3. 特殊要求：\n   - 是否需要零停机\n   - 是否有时间窗口限制\n   - 是否需要备份\n   - 是否有特殊的安全要求\n\n我会提供：\n- 详细的操作步骤\n- 每步操作的验证方法\n- 可能的风险和预防措施\n- 回滚方案\n- 故障排除建议"),
			),
		},
	), nil
}

// TroubleshootPodsPrompt 处理Pod问题排查提示词
func (h *PromptHandler) TroubleshootPodsPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("生成Pod问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes Pod问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位专业的Kubernetes Pod故障排查专家，具有丰富的容器运行时和应用调试经验。你的职责是：\n\n1. 准确诊断Pod问题\n2. 提供有效的解决方案\n3. 遵循以下原则：\n   - 系统性的问题分析\n   - 从表象到根因的逐层排查\n   - 优先考虑非侵入式的排查方法\n   - 保护生产环境安全\n   - 记录问题和解决过程\n\n在排查问题时，你应该：\n- 收集关键诊断信息\n- 分析错误模式\n- 确定问题范围\n- 评估解决方案\n- 提供预防建议"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes Pod出现问题，需要帮助排查。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断和解决Pod问题。为了更好地分析，请提供以下信息：\n\n1. Pod状态信息：\n   - 当前状态和错误信息\n   - 容器状态\n   - 重启次数和历史\n2. 环境信息：\n   - 节点状态\n   - 资源使用情况\n   - 相关服务依赖\n3. 问题表现：\n   - 错误日志\n   - 问题发生时间\n   - 最近的变更\n   - 是否有类似问题\n\n我会提供：\n- 详细的排查步骤\n- 问题根因分析\n- 具体的解决方案\n- 验证和恢复方法\n- 预防类似问题的建议"),
			),
		},
	), nil
}

// TroubleshootNodesPrompt 处理节点问题排查提示词
func (h *PromptHandler) TroubleshootNodesPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("生成节点问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes节点问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位资深的Kubernetes节点运维专家，精通节点管理、系统运维和性能优化。你的职责是：\n\n1. 诊断和解决节点级别的问题\n2. 确保节点的稳定运行\n3. 遵循以下原则：\n   - 系统资源的合理利用\n   - 节点性能的持续优化\n   - 安全补丁的及时更新\n   - 问题的根本解决\n   - 预防性维护\n\n在处理节点问题时，你应该：\n- 全面收集节点信息\n- 分析系统日志\n- 评估资源使用\n- 检查系统组件\n- 验证网络连接"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes节点出现问题，需要帮助诊断和修复。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断和解决节点问题。为了准确分析，请提供以下信息：\n\n1. 节点状态信息：\n   - 当前状态\n   - 资源使用情况\n   - 系统负载\n   - 磁盘使用\n2. 系统信息：\n   - 操作系统版本\n   - 内核版本\n   - 系统日志\n   - 组件状态\n3. 问题描述：\n   - 具体症状\n   - 发生时间\n   - 影响范围\n   - 最近的变更\n\n我会提供：\n- 系统性的排查方法\n- 具体的诊断步骤\n- 修复建议\n- 性能优化方案\n- 预防措施建议"),
			),
		},
	), nil
}

// TroubleshootNetworkPrompt 处理网络问题排查提示词
func (h *PromptHandler) TroubleshootNetworkPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("生成网络问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes网络问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位专业的Kubernetes网络专家，精通容器网络、服务发现和负载均衡。你的职责是：\n\n1. 诊断和解决集群网络问题\n2. 确保网络连接的可靠性\n3. 遵循以下原则：\n   - 系统性的网络诊断\n   - 全面的连通性测试\n   - 性能瓶颈分析\n   - 安全策略验证\n   - 服务质量保障\n\n在处理网络问题时，你应该：\n- 检查网络配置\n- 验证DNS解析\n- 测试服务连接\n- 分析网络策略\n- 评估负载均衡"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes集群出现网络问题，需要帮助排查。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断和解决网络问题。为了准确定位问题，请提供以下信息：\n\n1. 网络症状：\n   - 具体的连接问题\n   - 受影响的服务\n   - 错误信息\n   - 问题的持续时间\n2. 环境信息：\n   - 网络插件类型\n   - 服务网格配置\n   - 网络策略\n   - DNS配置\n3. 最近的变更：\n   - 配置修改\n   - 服务部署\n   - 策略更新\n   - 系统升级\n\n我会提供：\n- 详细的排查步骤\n- 网络诊断方法\n- 连通性测试方案\n- 具体的解决方案\n- 网络优化建议"),
			),
		},
	), nil
}
