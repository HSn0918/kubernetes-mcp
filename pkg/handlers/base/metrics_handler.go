package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Define metrics related tool constants
const (
	GET_NODE_METRICS     = "GET_NODE_METRICS"
	GET_POD_METRICS      = "GET_POD_METRICS"
	GET_RESOURCE_METRICS = "GET_RESOURCE_METRICS"
	GET_TOP_CONSUMERS    = "GET_TOP_CONSUMERS"
)

// MetricsHandler handles Kubernetes metrics related functions
type MetricsHandler struct {
	Handler
}

// Ensure interface implementation
var _ interfaces.ToolHandler = (*MetricsHandler)(nil)

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(client client.KubernetesClient) interfaces.ToolHandler {
	return &MetricsHandler{
		Handler: NewBaseHandler(client, interfaces.ClusterScope, interfaces.CoreAPIGroup),
	}
}

// Handle calls the appropriate handler function based on the request method
func (h *MetricsHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Log.Info("Handle called for metrics handler, method: ", request.Method)

	switch request.Method {
	case GET_NODE_METRICS:
		return h.GetNodeMetrics(ctx, request)
	case GET_POD_METRICS:
		return h.GetPodMetrics(ctx, request)
	case GET_RESOURCE_METRICS:
		return h.GetResourceMetrics(ctx, request)
	case GET_TOP_CONSUMERS:
		return h.GetTopConsumers(ctx, request)
	default:
		return nil, fmt.Errorf("unknown metrics method: %s", request.Method)
	}
}

// Register registers metric tools to the MCP server
func (h *MetricsHandler) Register(server *server.MCPServer) {
	h.Log.Info("Registering metrics handlers")
	// Register node metrics tool
	server.AddTool(mcp.NewTool(GET_NODE_METRICS,
		mcp.WithDescription("Get Kubernetes node metrics"),
		mcp.WithString("nodeName",
			mcp.Description("Node name (optional, retrieves all nodes if not specified)"),
		),
		mcp.WithString("sortBy",
			mcp.Description("Sort method (cpu, memory, cpu_percent, memory_percent, name)"),
			mcp.DefaultString("cpu"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes field selector (e.g. 'metadata.name=node-1')"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes label selector (e.g. 'kubernetes.io/role=master')"),
		),
	), h.GetNodeMetrics)

	// Register pod metrics tool
	server.AddTool(mcp.NewTool(GET_POD_METRICS,
		mcp.WithDescription("Get Kubernetes pod metrics"),
		mcp.WithString("namespace",
			mcp.Description("Namespace (optional, retrieves all namespaces if not specified)"),
		),
		mcp.WithString("podName",
			mcp.Description("Pod name (optional, retrieves all pods if not specified)"),
		),
		mcp.WithString("sortBy",
			mcp.Description("Sort method (cpu, memory, name)"),
			mcp.DefaultString("cpu"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Result count limit"),
			mcp.DefaultNumber(10),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes field selector (e.g. 'status.phase=Running')"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes label selector (e.g. 'app=nginx,tier=frontend')"),
		),
	), h.GetPodMetrics)

	// Register resource metrics tool
	server.AddTool(mcp.NewTool(GET_RESOURCE_METRICS,
		mcp.WithDescription("Get Kubernetes overall resource usage"),
		mcp.WithString("resource",
			mcp.Description("Resource type (cpu, memory, storage, pods)"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (optional, retrieves all namespaces if not specified)"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes field selector (e.g. 'status.phase=Running')"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes label selector (e.g. 'app=nginx,tier=frontend')"),
		),
	), h.GetResourceMetrics)

	// Register top consumers tool
	server.AddTool(mcp.NewTool(GET_TOP_CONSUMERS,
		mcp.WithDescription("Get Pods with highest resource consumption"),
		mcp.WithString("resource",
			mcp.Description("Resource type (cpu, memory)"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (optional, retrieves all namespaces if not specified)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Result count limit"),
			mcp.DefaultNumber(10),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes field selector (e.g. 'status.phase=Running')"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes label selector (e.g. 'app=nginx,tier=frontend')"),
		),
	), h.GetTopConsumers)

	// 注册集群资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("CLUSTER_RESOURCE_USAGE",
		mcp.WithPromptDescription("获取Kubernetes集群资源使用情况"),
		mcp.WithArgument("resource_type",
			mcp.ArgumentDescription("资源类型 (cpu, memory, storage, pods)"),
			mcp.RequiredArgument(),
		),
	), h.ClusterResourceUsagePrompt)

	// 注册节点资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("NODE_RESOURCE_USAGE",
		mcp.WithPromptDescription("获取Kubernetes节点资源使用情况"),
		mcp.WithArgument("node_name",
			mcp.ArgumentDescription("节点名称（可选）"),
		),
		mcp.WithArgument("sort_by",
			mcp.ArgumentDescription("排序方式 (cpu, memory, cpu_percent, memory_percent, name)"),
		),
	), h.NodeResourceUsagePrompt)

	// 注册Pod资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("POD_RESOURCE_USAGE",
		mcp.WithPromptDescription("获取Kubernetes Pod资源使用情况"),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("命名空间（可选）"),
		),
		mcp.WithArgument("pod_name",
			mcp.ArgumentDescription("Pod名称（可选）"),
		),
		mcp.WithArgument("sort_by",
			mcp.ArgumentDescription("排序方式 (cpu, memory, name)"),
		),
	), h.PodResourceUsagePrompt)
}

// GetNodeMetrics retrieves node resource usage metrics
func (h *MetricsHandler) GetNodeMetrics(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	nodeName, _ := arguments["nodeName"].(string)
	sortByStr, _ := arguments["sortBy"].(string)
	fieldSelector, _ := arguments["fieldSelector"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)

	h.Log.Info("Getting node metrics",
		"nodeName", nodeName,
		"sortBy", sortByStr,
		"fieldSelector", fieldSelector,
		"labelSelector", labelSelector,
	)

	var nodeMetrics []models.NodeMetricInfo
	var err error

	// If node name is specified, get metrics for that node only
	if nodeName != "" {
		nodeMetric, err := utils.GetNodeMetric(ctx, h.Client, nodeName)
		if err != nil {
			return nil, err
		}

		// Create NodeResponse object
		result := models.NodeResponse{
			Name:              nodeMetric.Name,
			CPUUsage:          nodeMetric.CPUUsage,
			CPUAllocatable:    nodeMetric.CPUAllocatable,
			CPUPercent:        nodeMetric.CPUPercent,
			MemoryUsage:       nodeMetric.MemoryUsage,
			MemoryAllocatable: nodeMetric.MemoryAllocatable,
			MemoryPercent:     nodeMetric.MemoryPercent,
			Timestamp:         nodeMetric.Timestamp,
			UpdatedAgo:        utils.FormatTimeAgo(nodeMetric.Timestamp),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("JSON formatting failed: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(jsonData),
				},
			},
		}, nil
	}

	// Prepare options for getting metrics for all nodes
	options := []utils.MetricsOption{utils.WithSortByString(sortByStr)}

	// Add field selector if provided
	if fieldSelector != "" {
		options = append(options, utils.WithFieldSelector(fieldSelector))
	}

	// Add label selector if provided
	if labelSelector != "" {
		options = append(options, utils.WithLabelSelector(labelSelector))
	}

	// Get metrics for all nodes using functional options pattern
	nodeMetrics, err = utils.GetNodesMetrics(
		ctx,
		h.Client,
		options...,
	)

	if err != nil {
		return nil, err
	}

	// Create NodesListResponse object
	result := models.NodesListResponse{
		Nodes:      make([]models.NodeResponse, 0, len(nodeMetrics)),
		SortBy:     string(utils.ParseSortType(sortByStr)),
		TotalCount: len(nodeMetrics),
	}

	for _, metric := range nodeMetrics {
		result.Nodes = append(result.Nodes, models.NodeResponse{
			Name:              metric.Name,
			CPUUsage:          metric.CPUUsage,
			CPUAllocatable:    metric.CPUAllocatable,
			CPUPercent:        metric.CPUPercent,
			MemoryUsage:       metric.MemoryUsage,
			MemoryAllocatable: metric.MemoryAllocatable,
			MemoryPercent:     metric.MemoryPercent,
			Timestamp:         metric.Timestamp,
			UpdatedAgo:        utils.FormatTimeAgo(metric.Timestamp),
		})
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON formatting failed: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// GetPodMetrics retrieves Pod resource usage metrics
func (h *MetricsHandler) GetPodMetrics(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	namespace, _ := arguments["namespace"].(string)
	podName, _ := arguments["podName"].(string)
	sortByStr, _ := arguments["sortBy"].(string)
	limit, _ := arguments["limit"].(float64)
	fieldSelector, _ := arguments["fieldSelector"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)

	h.Log.Info("Getting pod metrics",
		"namespace", namespace,
		"podName", podName,
		"sortBy", sortByStr,
		"limit", limit,
		"fieldSelector", fieldSelector,
		"labelSelector", labelSelector,
	)

	// Prepare options
	var options []utils.MetricsOption
	options = append(options, utils.WithSortByString(sortByStr))
	options = append(options, utils.WithLimit(int(limit)))

	// If pod name is specified, add pod name filter
	if podName != "" {
		options = append(options, utils.WithPodNameFilter(podName))
	}

	// Add field selector if provided
	if fieldSelector != "" {
		options = append(options, utils.WithFieldSelector(fieldSelector))
	}

	// Add label selector if provided
	if labelSelector != "" {
		options = append(options, utils.WithLabelSelector(labelSelector))
	}

	// Get Pod metrics using functional options pattern
	podMetrics, err := utils.GetPodsMetrics(ctx, h.Client, namespace, options...)
	if err != nil {
		return nil, err
	}

	// Create PodsListResponse object
	result := models.PodsListResponse{
		Pods:          make([]models.PodResponse, 0, len(podMetrics)),
		SortBy:        string(utils.ParseSortType(sortByStr)),
		TotalCount:    len(podMetrics),
		Namespace:     namespace,
		Limit:         int(limit),
		IncludeDetail: podName != "", // Include details if pod name is specified
	}

	for _, pod := range podMetrics {
		podResp := models.PodResponse{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			TotalCPU:    pod.TotalCPU,
			TotalMemory: pod.TotalMemory,
			Timestamp:   pod.Timestamp,
			UpdatedAgo:  utils.FormatTimeAgo(pod.Timestamp),
		}

		// If pod name is specified, include container details
		if podName != "" && pod.Name == podName {
			podResp.Containers = make([]models.ContainerResponse, 0, len(pod.Containers))
			for _, container := range pod.Containers {
				podResp.Containers = append(podResp.Containers, models.ContainerResponse{
					Name:        container.Name,
					CPUUsage:    container.CPUUsage,
					MemoryUsage: container.MemoryUsage,
				})
			}
		}

		result.Pods = append(result.Pods, podResp)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON formatting failed: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// GetResourceMetrics retrieves overall resource usage
func (h *MetricsHandler) GetResourceMetrics(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	resourceType, _ := arguments["resource"].(string)
	namespace, _ := arguments["namespace"].(string)
	fieldSelector, _ := arguments["fieldSelector"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)

	h.Log.Info("Getting resource metrics",
		"resourceType", resourceType,
		"namespace", namespace,
		"fieldSelector", fieldSelector,
		"labelSelector", labelSelector,
	)

	// Prepare options
	var options []utils.MetricsOption
	options = append(options, utils.WithResourceFilter(resourceType))

	// Add human-readable format option if needed
	options = append(options, utils.WithUnitType("human"))

	// Include detailed information by default
	options = append(options, utils.WithIncludeDetail(true))

	// Add field selector if provided
	if fieldSelector != "" {
		options = append(options, utils.WithFieldSelector(fieldSelector))
	}

	// Add label selector if provided
	if labelSelector != "" {
		options = append(options, utils.WithLabelSelector(labelSelector))
	}

	// Get cluster resource metrics using functional options pattern
	metrics, err := utils.GetClusterResourceMetrics(ctx, h.Client, namespace, options...)
	if err != nil {
		return nil, err
	}

	// Create ResourceMetricsResponse object
	result := models.ResourceMetricsResponse{
		ResourceType: resourceType,
		UnitType:     metrics.UnitType,
	}

	// Fill fields based on resource type
	switch resourceType {
	case "cpu":
		result.CPUCapacity = metrics.CPUCapacity
		result.CPUAllocatable = metrics.CPUAllocatable
		result.CPUUsage = metrics.CPUUsage
		result.CPUPercent = metrics.CPUPercent
		result.CPUAvailable = metrics.CPUAllocatable - metrics.CPUUsage

	case "memory":
		result.MemoryCapacity = metrics.MemoryCapacity
		result.MemoryAllocatable = metrics.MemoryAllocatable
		result.MemoryUsage = metrics.MemoryUsage
		result.MemoryPercent = metrics.MemoryPercent
		result.MemoryAvailable = metrics.MemoryAllocatable - metrics.MemoryUsage

	case "storage":
		result.StorageCapacity = metrics.StorageCapacity
		result.StorageAllocatable = metrics.StorageAllocatable

	case "pods":
		result.PodCapacity = metrics.PodCapacity
		result.RunningPods = metrics.RunningPods
		result.PodPercent = metrics.PodPercent
		result.PodsAvailable = metrics.PodCapacity - int64(metrics.RunningPods)

	default:
		// Include information for all resource types
		result.ResourceType = "all"

		result.CPUCapacity = metrics.CPUCapacity
		result.CPUAllocatable = metrics.CPUAllocatable
		result.CPUUsage = metrics.CPUUsage
		result.CPUPercent = metrics.CPUPercent
		result.CPUAvailable = metrics.CPUAllocatable - metrics.CPUUsage

		result.MemoryCapacity = metrics.MemoryCapacity
		result.MemoryAllocatable = metrics.MemoryAllocatable
		result.MemoryUsage = metrics.MemoryUsage
		result.MemoryPercent = metrics.MemoryPercent
		result.MemoryAvailable = metrics.MemoryAllocatable - metrics.MemoryUsage

		result.StorageCapacity = metrics.StorageCapacity
		result.StorageAllocatable = metrics.StorageAllocatable

		result.PodCapacity = metrics.PodCapacity
		result.RunningPods = metrics.RunningPods
		result.PodPercent = metrics.PodPercent
		result.PodsAvailable = metrics.PodCapacity - int64(metrics.RunningPods)
	}

	// Add namespace information if specified
	if namespace != "" {
		result.Namespace = namespace
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON formatting failed: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// GetTopConsumers retrieves pods with highest resource consumption
func (h *MetricsHandler) GetTopConsumers(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	resourceType, _ := arguments["resource"].(string)
	namespace, _ := arguments["namespace"].(string)
	limit, _ := arguments["limit"].(float64)
	fieldSelector, _ := arguments["fieldSelector"].(string)
	labelSelector, _ := arguments["labelSelector"].(string)

	h.Log.Info("Getting top consumers",
		"resourceType", resourceType,
		"namespace", namespace,
		"limit", limit,
		"fieldSelector", fieldSelector,
		"labelSelector", labelSelector,
	)

	// Validate resource type
	if resourceType != "cpu" && resourceType != "memory" {
		return nil, fmt.Errorf("unsupported resource type: %s, supported types are: cpu, memory", resourceType)
	}

	// Select sort type based on resource type
	var sortType models.SortType
	if resourceType == "cpu" {
		sortType = models.SortByCPU
	} else {
		sortType = models.SortByMemory
	}

	// Prepare options
	options := []utils.MetricsOption{
		utils.WithSortType(sortType),
		utils.WithLimit(int(limit)),
	}

	// Add field selector if provided
	if fieldSelector != "" {
		options = append(options, utils.WithFieldSelector(fieldSelector))
	}

	// Add label selector if provided
	if labelSelector != "" {
		options = append(options, utils.WithLabelSelector(labelSelector))
	}

	// Get Pod metrics sorted by resource usage using functional options pattern
	podMetrics, err := utils.GetPodsMetrics(
		ctx,
		h.Client,
		namespace,
		options...,
	)
	if err != nil {
		return nil, err
	}

	// Create TopConsumersListResponse object
	result := models.TopConsumersListResponse{
		Consumers:    make([]models.TopConsumerResponse, 0, len(podMetrics)),
		ResourceType: resourceType,
		Limit:        int(limit),
		Namespace:    namespace,
		TotalCount:   len(podMetrics),
	}

	for _, pod := range podMetrics {
		usageValue := pod.TotalCPU
		if resourceType == "memory" {
			usageValue = pod.TotalMemory
		}

		result.Consumers = append(result.Consumers, models.TopConsumerResponse{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Usage:      usageValue,
			Timestamp:  pod.Timestamp,
			UpdatedAgo: utils.FormatTimeAgo(pod.Timestamp),
		})
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON formatting failed: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// ClusterResourceUsagePrompt 处理集群资源使用情况提示词
func (h *MetricsHandler) ClusterResourceUsagePrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("处理集群资源使用情况提示词")

	// 序列化模板为JSON格式
	template := models.ClusterResourcePrompt
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建promptText并加入JSON内容
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes集群资源使用情况提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	// 创建标准的GetPromptResult
	return mcp.NewGetPromptResult(
		"Kubernetes集群资源使用情况",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是Kubernetes集群管理员，提供准确的集群资源使用情况分析。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("请分析Kubernetes集群的资源使用情况，包括CPU、内存、存储和Pod数量。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会为你提供集群资源使用情况的详细分析，包括资源使用百分比和可用资源状态。"),
			),
		},
	), nil
}

// NodeResourceUsagePrompt 处理节点资源使用情况提示词
func (h *MetricsHandler) NodeResourceUsagePrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("处理节点资源使用情况提示词")

	// 序列化模板为JSON格式
	template := models.NodeResourcePrompt
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建promptText并加入JSON内容
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes节点资源使用情况提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	// 创建标准的GetPromptResult
	return mcp.NewGetPromptResult(
		"Kubernetes节点资源使用情况",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是Kubernetes集群管理员，提供准确的节点资源使用情况分析。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("请分析Kubernetes集群中各节点的资源使用情况，帮我找出负载高的节点。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会为你分析各节点的CPU和内存使用情况，并帮你识别负载较高或资源紧张的节点。"),
			),
		},
	), nil
}

// PodResourceUsagePrompt 处理Pod资源使用情况提示词
func (h *MetricsHandler) PodResourceUsagePrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	h.Log.Info("处理Pod资源使用情况提示词")

	// 序列化模板为JSON格式
	template := models.PodResourcePrompt
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建promptText并加入JSON内容
	var promptText strings.Builder
	promptText.WriteString("=== Kubernetes Pod资源使用情况提示词 ===\n\n")
	promptText.WriteString(string(jsonData))

	// 创建标准的GetPromptResult
	return mcp.NewGetPromptResult(
		"Kubernetes Pod资源使用情况",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				"system",
				mcp.NewTextContent("你是Kubernetes集群管理员，提供准确的Pod资源使用情况分析。"),
			),
			mcp.NewPromptMessage(
				"user",
				mcp.NewTextContent("请分析Kubernetes集群中各Pod的资源使用情况，帮我找出资源消耗较高的Pod。"),
			),
			mcp.NewPromptMessage(
				"assistant",
				mcp.NewTextContent("我会为你分析各Pod的CPU和内存使用情况，并帮你识别资源消耗较高的Pod。"),
			),
		},
	), nil
}
