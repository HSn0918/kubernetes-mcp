package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
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
	base.Handler
}

// Ensure interface implementation
var _ interfaces.ToolHandler = (*MetricsHandler)(nil)

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(client kubernetes.Client) interfaces.ToolHandler {
	return &MetricsHandler{
		Handler: base.NewHandler(client, interfaces.ClusterScope, interfaces.Metrics),
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
		return utils.NewErrorToolResult(fmt.Sprintf("unknown metrics method: %s", request.Method)), nil
	}
}

// Register registers metric tools to the MCP server
func (h *MetricsHandler) Register(server *server.MCPServer) {
	h.Log.Info("Registering metrics handlers")
	// Register node metrics tool
	server.AddTool(mcp.NewTool(GET_NODE_METRICS,
		mcp.WithDescription("获取Kubernetes节点资源使用指标。提供节点级别的CPU、内存、磁盘等资源使用情况，支持多种排序方式和过滤条件。适用于节点性能监控、容量规划、资源分配优化等场景。可用于识别资源瓶颈和性能热点。"),
		mcp.WithString("nodeName",
			mcp.Description("节点名称（可选）。不指定时获取所有节点的指标。支持精确匹配，用于监控特定节点的资源使用情况。"),
		),
		mcp.WithString("sortBy",
			mcp.Description("排序方式，支持以下选项：\n- cpu：按CPU使用量排序\n- memory：按内存使用量排序\n- cpu_percent：按CPU使用率百分比排序\n- memory_percent：按内存使用率百分比排序\n- name：按节点名称字母顺序排序"),
			mcp.DefaultString("cpu"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按节点属性进行过滤。例如：'metadata.name=node-1'。支持多个条件，使用逗号分隔。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按节点标签进行过滤。例如：'kubernetes.io/role=master'。支持多个标签，使用逗号分隔。"),
		),
	), h.GetNodeMetrics)

	// Register pod metrics tool
	server.AddTool(mcp.NewTool(GET_POD_METRICS,
		mcp.WithDescription("获取Kubernetes Pod资源使用指标。监控Pod级别的CPU、内存使用情况，支持namespace过滤、名称搜索和多种排序方式。适用于应用性能监控、资源使用分析、容量规划等场景。可用于优化应用资源配置和问题诊断。"),
		mcp.WithString("namespace",
			mcp.Description("命名空间（可选）。不指定时获取所有命名空间的Pod指标。用于监控特定业务域的资源使用情况。"),
		),
		mcp.WithString("podName",
			mcp.Description("Pod名称（可选）。不指定时获取所有Pod的指标。支持精确匹配，用于监控特定应用实例的资源使用情况。"),
		),
		mcp.WithString("sortBy",
			mcp.Description("排序方式，支持以下选项：\n- cpu：按CPU使用量排序\n- memory：按内存使用量排序\n- name：按Pod名称字母顺序排序\n用于快速识别资源消耗较高的Pod。"),
			mcp.DefaultString("cpu"),
		),
		mcp.WithNumber("limit",
			mcp.Description("结果数量限制。默认返回资源使用最高的10个Pod。较大的限制值可能影响查询性能。"),
			mcp.DefaultNumber(10),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按Pod属性进行过滤。例如：'status.phase=Running'。可用于筛选特定状态的Pod。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按Pod标签进行过滤。例如：'app=nginx,tier=frontend'。用于监控特定应用或组件的资源使用情况。"),
		),
	), h.GetPodMetrics)

	// Register resource metrics tool
	server.AddTool(mcp.NewTool(GET_RESOURCE_METRICS,
		mcp.WithDescription("获取Kubernetes集群整体资源使用情况。提供集群级别的CPU、内存、存储和Pod数量统计，支持按命名空间和标签过滤。适用于集群容量规划、资源使用趋势分析、成本优化等场景。帮助了解资源使用效率和分布情况。"),
		mcp.WithString("resource",
			mcp.Description("资源类型，支持以下选项：\n- cpu：CPU使用情况\n- memory：内存使用情况\n- storage：存储使用情况\n- pods：Pod数量统计\n选择要分析的具体资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("命名空间（可选）。不指定时统计所有命名空间的资源使用情况。用于分析特定业务域的资源消耗。"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按资源属性进行过滤。例如：'status.phase=Running'。帮助关注特定状态的资源使用情况。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按资源标签进行过滤。例如：'app=nginx,tier=frontend'。用于分析特定应用或组件的资源使用情况。"),
		),
	), h.GetResourceMetrics)

	// Register top consumers tool
	server.AddTool(mcp.NewTool(GET_TOP_CONSUMERS,
		mcp.WithDescription("获取资源消耗最高的Pods列表。识别集群中CPU或内存使用率最高的Pod，支持namespace过滤和自定义返回数量。适用于性能热点分析、资源优化、成本控制等场景。帮助快速定位资源密集型应用。"),
		mcp.WithString("resource",
			mcp.Description("资源类型，支持以下选项：\n- cpu：按CPU使用量排序\n- memory：按内存使用量排序\n选择要分析的资源类型。"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("命名空间（可选）。不指定时分析所有命名空间的Pod。用于关注特定业务域的资源消耗情况。"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回结果数量限制。默认返回前10个资源消耗最高的Pod。较大的限制值可能影响查询性能。"),
			mcp.DefaultNumber(10),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Kubernetes字段选择器，用于按Pod属性进行过滤。例如：'status.phase=Running'。帮助筛选特定状态的Pod。"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Kubernetes标签选择器，用于按Pod标签进行过滤。例如：'app=nginx,tier=frontend'。用于分析特定应用或组件的资源消耗情况。"),
		),
	), h.GetTopConsumers)

	// 注册集群资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("CLUSTER_RESOURCE_USAGE",
		mcp.WithPromptDescription("分析Kubernetes集群资源使用情况，包括CPU、内存、存储和Pod数量的使用统计。提供资源使用趋势、分布情况和优化建议。帮助进行容量规划和资源优化。"),
		mcp.WithArgument("resource_type",
			mcp.ArgumentDescription("资源类型，支持：\n- cpu：CPU使用分析\n- memory：内存使用分析\n- storage：存储使用分析\n- pods：Pod分布分析\n选择要分析的资源类型。"),
			mcp.RequiredArgument(),
		),
	), h.ClusterResourceUsagePrompt)

	// 注册节点资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("NODE_RESOURCE_USAGE",
		mcp.WithPromptDescription("分析Kubernetes节点资源使用情况，包括CPU、内存使用率，系统负载等指标。提供节点级别的资源使用分析、性能评估和优化建议。帮助优化节点资源分配和负载均衡。"),
		mcp.WithArgument("node_name",
			mcp.ArgumentDescription("节点名称（可选）。不指定时分析所有节点。用于深入分析特定节点的资源使用情况。"),
		),
		mcp.WithArgument("sort_by",
			mcp.ArgumentDescription("排序方式：\n- cpu：按CPU使用量排序\n- memory：按内存使用量排序\n- cpu_percent：按CPU使用率排序\n- memory_percent：按内存使用率排序\n- name：按节点名称排序"),
		),
	), h.NodeResourceUsagePrompt)

	// 注册Pod资源使用情况提示词
	server.AddPrompt(mcp.NewPrompt("POD_RESOURCE_USAGE",
		mcp.WithPromptDescription("分析Kubernetes Pod资源使用情况，包括实时资源消耗、历史趋势和异常检测。提供容器级别的资源使用分析、性能评估和优化建议。帮助优化应用资源配置和性能调优。"),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("命名空间（可选）。不指定时分析所有命名空间的Pod。用于关注特定业务域的资源使用情况。"),
		),
		mcp.WithArgument("pod_name",
			mcp.ArgumentDescription("Pod名称（可选）。不指定时分析所有Pod。用于深入分析特定应用实例的资源使用情况。"),
		),
		mcp.WithArgument("sort_by",
			mcp.ArgumentDescription("排序方式：\n- cpu：按CPU使用量排序\n- memory：按内存使用量排序\n- name：按Pod名称排序\n用于快速识别资源消耗异常的Pod。"),
		),
	), h.PodResourceUsagePrompt)
}

// GetNodeMetrics retrieves node resource usage metrics
func (h *MetricsHandler) GetNodeMetrics(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
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
			return utils.NewErrorToolResult(fmt.Sprintf("Failed to get node metric: %v", err)), nil
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
			return utils.NewErrorToolResult(fmt.Sprintf("JSON formatting failed: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("Failed to get nodes metrics: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("JSON formatting failed: %v", err)), nil
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
	arguments := request.GetArguments()
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
		return utils.NewErrorToolResult(fmt.Sprintf("Failed to get pod metrics: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("JSON formatting failed: %v", err)), nil
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
	arguments := request.GetArguments()
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
		return utils.NewErrorToolResult(fmt.Sprintf("Failed to get cluster resource metrics: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("JSON formatting failed: %v", err)), nil
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
	arguments := request.GetArguments()
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
		return utils.NewErrorToolResult(fmt.Sprintf("unsupported resource type: %s, supported types are: cpu, memory", resourceType)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("Failed to get pod metrics: %v", err)), nil
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
		return utils.NewErrorToolResult(fmt.Sprintf("JSON formatting failed: %v", err)), nil
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
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
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
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
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
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
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
