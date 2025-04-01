package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	Handler
}

// 确保实现了接口
var _ interfaces.ToolHandler = (*PromptHandler)(nil)

// NewPromptHandler 创建新的提示词处理程序
func NewPromptHandler(client client.KubernetesClient) *PromptHandler {
	return &PromptHandler{
		Handler: NewBaseHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
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
		mcp.WithPromptDescription("生成Kubernetes YAML资源清单"),
		mcp.WithArgument("resource_type",
			mcp.ArgumentDescription("资源类型 (例如 Deployment, Service, ConfigMap)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("name",
			mcp.ArgumentDescription("资源名称"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("命名空间"),
		),
	), h.KubernetesYAMLPrompt)

	// 同时将YAML提示词作为工具注册
	s.AddTool(mcp.NewTool(KUBERNETES_YAML_PROMPT,
		mcp.WithDescription("生成Kubernetes YAML资源清单"),
		mcp.WithString("resource_type",
			mcp.Description("资源类型 (例如 Deployment, Service, ConfigMap)"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("资源名称"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("命名空间"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleKubernetesYAMLPrompt(ctx, request)
	})

	// Kubernetes查询提示词
	s.AddPrompt(mcp.NewPrompt(KUBERNETES_QUERY_PROMPT,
		mcp.WithPromptDescription("Kubernetes操作指导"),
		mcp.WithArgument("task",
			mcp.ArgumentDescription("需要执行的任务描述"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("context",
			mcp.ArgumentDescription("额外的上下文信息"),
		),
	), h.KubernetesQueryPrompt)

	// 同时将查询提示词作为工具注册
	s.AddTool(mcp.NewTool(KUBERNETES_QUERY_PROMPT,
		mcp.WithDescription("Kubernetes操作指导"),
		mcp.WithString("task",
			mcp.Description("需要执行的任务描述"),
			mcp.Required(),
		),
		mcp.WithString("context",
			mcp.Description("额外的上下文信息"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleKubernetesQueryPrompt(ctx, request)
	})

	// Pod问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_PODS_PROMPT,
		mcp.WithPromptDescription("Kubernetes Pod问题排查"),
		mcp.WithArgument("pod_status",
			mcp.ArgumentDescription("Pod状态 (例如 CrashLoopBackOff, Pending, ImagePullBackOff)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("pod_logs",
			mcp.ArgumentDescription("Pod日志内容"),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("命名空间"),
		),
	), h.TroubleshootPodsPrompt)

	// 同时将Pod问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_PODS_PROMPT,
		mcp.WithDescription("Kubernetes Pod问题排查"),
		mcp.WithString("pod_status",
			mcp.Description("Pod状态 (例如 CrashLoopBackOff, Pending, ImagePullBackOff)"),
			mcp.Required(),
		),
		mcp.WithString("pod_logs",
			mcp.Description("Pod日志内容"),
		),
		mcp.WithString("namespace",
			mcp.Description("命名空间"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootPodsPrompt(ctx, request)
	})

	// 节点问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_NODES_PROMPT,
		mcp.WithPromptDescription("Kubernetes节点问题排查"),
		mcp.WithArgument("node_status",
			mcp.ArgumentDescription("节点状态 (例如 Ready, NotReady, MemoryPressure, DiskPressure)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("node_conditions",
			mcp.ArgumentDescription("节点状况详情"),
		),
	), h.TroubleshootNodesPrompt)

	// 同时将节点问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_NODES_PROMPT,
		mcp.WithDescription("Kubernetes节点问题排查"),
		mcp.WithString("node_status",
			mcp.Description("节点状态 (例如Ready, NotReady, MemoryPressure, DiskPressure)"),
			mcp.Required(),
		),
		mcp.WithString("node_conditions",
			mcp.Description("节点状况详情"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootNodesPrompt(ctx, request)
	})

	// 网络问题排查提示词
	s.AddPrompt(mcp.NewPrompt(TROUBLESHOOT_NET_PROMPT,
		mcp.WithPromptDescription("Kubernetes网络问题排查"),
		mcp.WithArgument("problem_type",
			mcp.ArgumentDescription("问题类型 (例如 服务不可达, DNS问题, Ingress不工作)"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("service_name",
			mcp.ArgumentDescription("相关服务名称"),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("命名空间"),
		),
	), h.TroubleshootNetworkPrompt)

	// 同时将网络问题排查提示词作为工具注册
	s.AddTool(mcp.NewTool(TROUBLESHOOT_NET_PROMPT,
		mcp.WithDescription("Kubernetes网络问题排查"),
		mcp.WithString("problem_type",
			mcp.Description("问题类型 (例如 服务不可达, DNS问题, Ingress不工作)"),
			mcp.Required(),
		),
		mcp.WithString("service_name",
			mcp.Description("相关服务名称"),
		),
		mcp.WithString("namespace",
			mcp.Description("命名空间"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return h.handleTroubleshootNetworkPrompt(ctx, request)
	})
}

// KubernetesYAMLPrompt 处理 Kubernetes YAML 生成提示词
func (h *PromptHandler) KubernetesYAMLPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 简化处理逻辑
	h.Log.Info("生成Kubernetes YAML提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes YAML 生成",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位Kubernetes专家。请根据用户的需求，生成符合最佳实践的YAML资源清单。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我需要一个Kubernetes资源清单。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你创建一个符合Kubernetes最佳实践的YAML配置。"),
			),
		},
	), nil
}

// KubernetesQueryPrompt 处理 Kubernetes 查询提示词
func (h *PromptHandler) KubernetesQueryPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 简化处理逻辑
	h.Log.Info("生成Kubernetes操作指导提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes操作指导",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位Kubernetes操作专家。请提供准确的指导和操作步骤，帮助用户完成各种Kubernetes管理任务。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我需要在Kubernetes集群中执行某些操作，请提供指导。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你提供详细的操作步骤，以下是操作指南："),
			),
		},
	), nil
}

// TroubleshootPodsPrompt 处理Pod问题排查提示词
func (h *PromptHandler) TroubleshootPodsPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 简化处理逻辑
	h.Log.Info("生成Pod问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes Pod问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位Kubernetes故障排查专家。分析Pod问题原因并提供解决方案。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes Pod出现问题，请帮我排查。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断Pod问题并提供解决方案。以下是常见Pod问题的排查步骤："),
			),
		},
	), nil
}

// TroubleshootNodesPrompt 处理节点问题排查提示词
func (h *PromptHandler) TroubleshootNodesPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 简化处理逻辑
	h.Log.Info("生成节点问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes节点问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位Kubernetes节点管理专家。分析节点问题并提供修复方案。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes集群节点出现问题，请帮我排查。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断节点问题并提供解决方案。以下是节点问题的排查步骤："),
			),
		},
	), nil
}

// TroubleshootNetworkPrompt 处理网络问题排查提示词
func (h *PromptHandler) TroubleshootNetworkPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 简化处理逻辑
	h.Log.Info("生成网络问题排查提示词")

	return mcp.NewGetPromptResult(
		"Kubernetes网络问题排查",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是一位Kubernetes网络专家。分析网络连接问题并提供解决方案。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("我的Kubernetes集群网络出现问题，请帮我排查。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会帮你诊断网络问题并提供解决方案。以下是网络故障的排查步骤："),
			),
		},
	), nil
}
