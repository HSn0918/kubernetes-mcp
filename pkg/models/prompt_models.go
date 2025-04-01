package models

// PromptTemplate 定义提示词模板结构
type PromptTemplate struct {
	Title    string          `json:"title"`
	Messages []PromptMessage `json:"messages"`
}

// PromptMessage 定义提示词消息结构
type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClusterResourcePrompt 集群资源使用情况提示词模板
var ClusterResourcePrompt = PromptTemplate{
	Title: "Kubernetes集群资源使用情况",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是Kubernetes集群管理员，提供准确的集群资源使用情况分析。",
		},
		{
			Role:    "user",
			Content: "请分析Kubernetes集群的资源使用情况，包括CPU、内存、存储和Pod数量。",
		},
		{
			Role:    "assistant",
			Content: "我会为你提供集群资源使用情况的详细分析，包括资源使用百分比和可用资源状态。",
		},
	},
}

// NodeResourcePrompt 节点资源使用情况提示词模板
var NodeResourcePrompt = PromptTemplate{
	Title: "Kubernetes节点资源使用情况",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是Kubernetes集群管理员，提供准确的节点资源使用情况分析。",
		},
		{
			Role:    "user",
			Content: "请分析Kubernetes集群中各节点的资源使用情况，帮我找出负载高的节点。",
		},
		{
			Role:    "assistant",
			Content: "我会为你分析各节点的CPU和内存使用情况，并帮你识别负载较高或资源紧张的节点。",
		},
	},
}

// PodResourcePrompt Pod资源使用情况提示词模板
var PodResourcePrompt = PromptTemplate{
	Title: "Kubernetes Pod资源使用情况",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是Kubernetes集群管理员，提供准确的Pod资源使用情况分析。",
		},
		{
			Role:    "user",
			Content: "请分析Kubernetes集群中各Pod的资源使用情况，帮我找出资源消耗较高的Pod。",
		},
		{
			Role:    "assistant",
			Content: "我会为你分析各Pod的CPU和内存使用情况，并帮你识别资源消耗较高的Pod。",
		},
	},
}

// KubernetesYAMLPrompt Kubernetes YAML生成提示词模板
var KubernetesYAMLPrompt = PromptTemplate{
	Title: "Kubernetes YAML 生成",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是一位Kubernetes专家。请根据用户的需求，生成符合最佳实践的YAML资源清单。",
		},
		{
			Role:    "user",
			Content: "我需要一个Kubernetes资源清单。",
		},
		{
			Role:    "assistant",
			Content: "我会帮你创建一个符合Kubernetes最佳实践的YAML配置。",
		},
	},
}

// KubernetesQueryPrompt Kubernetes操作指导提示词模板
var KubernetesQueryPrompt = PromptTemplate{
	Title: "Kubernetes操作指导",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是一位Kubernetes操作专家。请提供准确的指导和操作步骤，帮助用户完成各种Kubernetes管理任务。",
		},
		{
			Role:    "user",
			Content: "我需要在Kubernetes集群中执行某些操作，请提供指导。",
		},
		{
			Role:    "assistant",
			Content: "我会帮你提供详细的操作步骤，以下是操作指南：",
		},
	},
}

// TroubleshootPodsPrompt Pod问题排查提示词模板
var TroubleshootPodsPrompt = PromptTemplate{
	Title: "Kubernetes Pod问题排查",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是一位Kubernetes故障排查专家。分析Pod问题原因并提供解决方案。",
		},
		{
			Role:    "user",
			Content: "我的Kubernetes Pod出现问题，请帮我排查。",
		},
		{
			Role:    "assistant",
			Content: "我会帮你诊断Pod问题并提供解决方案。以下是常见Pod问题的排查步骤：",
		},
	},
}

// TroubleshootNodesPrompt 节点问题排查提示词模板
var TroubleshootNodesPrompt = PromptTemplate{
	Title: "Kubernetes节点问题排查",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是一位Kubernetes节点管理专家。分析节点问题并提供修复方案。",
		},
		{
			Role:    "user",
			Content: "我的Kubernetes集群节点出现问题，请帮我排查。",
		},
		{
			Role:    "assistant",
			Content: "我会帮你诊断节点问题并提供解决方案。以下是节点问题的排查步骤：",
		},
	},
}

// TroubleshootNetworkPrompt 网络问题排查提示词模板
var TroubleshootNetworkPrompt = PromptTemplate{
	Title: "Kubernetes网络问题排查",
	Messages: []PromptMessage{
		{
			Role:    "system",
			Content: "你是一位Kubernetes网络专家。分析网络连接问题并提供解决方案。",
		},
		{
			Role:    "user",
			Content: "我的Kubernetes集群网络出现问题，请帮我排查。",
		},
		{
			Role:    "assistant",
			Content: "我会帮你诊断网络问题并提供解决方案。以下是网络故障的排查步骤：",
		},
	},
}
