package utils

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// MetricsOption defines the function type for metrics query options
type MetricsOption func(*MetricsOptions)

// MetricsOptions stores configuration options for metrics queries
type MetricsOptions struct {
	// Sort type
	SortType models.SortType
	// Result limit count
	Limit int
	// Filter function, returns true to keep the item
	NodeFilter func(models.NodeMetricInfo) bool
	// Pod filter function, returns true to keep the item
	PodFilter func(models.PodMetricInfo) bool
	// Resource type (cpu, memory, storage, pods)
	ResourceType string
	// Whether to include detailed information
	IncludeDetail bool
	// Unit type (raw, percent, human)
	UnitType string
	// Kubernetes field selector
	FieldSelector string
	// Kubernetes label selector
	LabelSelector string
}

// WithSortType sets the sort type
func WithSortType(sortType models.SortType) MetricsOption {
	return func(options *MetricsOptions) {
		options.SortType = sortType
	}
}

// WithSortByString sets the sort type from a string
func WithSortByString(sortByStr string) MetricsOption {
	return func(options *MetricsOptions) {
		options.SortType = ParseSortType(sortByStr)
	}
}

// WithLimit sets the result count limit
func WithLimit(limit int) MetricsOption {
	return func(options *MetricsOptions) {
		options.Limit = limit
	}
}

// WithNodeFilter sets the node filter function
func WithNodeFilter(filter func(models.NodeMetricInfo) bool) MetricsOption {
	return func(options *MetricsOptions) {
		options.NodeFilter = filter
	}
}

// WithPodFilter sets the pod filter function
func WithPodFilter(filter func(models.PodMetricInfo) bool) MetricsOption {
	return func(options *MetricsOptions) {
		options.PodFilter = filter
	}
}

// WithNodeNameFilter creates a filter function based on node name
func WithNodeNameFilter(nodeName string) MetricsOption {
	return WithNodeFilter(func(node models.NodeMetricInfo) bool {
		return node.Name == nodeName
	})
}

// WithNamespaceFilter creates a pod filter function based on namespace
func WithNamespaceFilter(namespace string) MetricsOption {
	return WithPodFilter(func(pod models.PodMetricInfo) bool {
		return pod.Namespace == namespace
	})
}

// WithPodNameFilter creates a filter function based on pod name
func WithPodNameFilter(podName string) MetricsOption {
	return WithPodFilter(func(pod models.PodMetricInfo) bool {
		return pod.Name == podName
	})
}

// WithFieldSelector sets the Kubernetes field selector
func WithFieldSelector(fieldSelector string) MetricsOption {
	return func(options *MetricsOptions) {
		options.FieldSelector = fieldSelector
	}
}

// WithLabelSelector sets the Kubernetes label selector
func WithLabelSelector(labelSelector string) MetricsOption {
	return func(options *MetricsOptions) {
		options.LabelSelector = labelSelector
	}
}

// WithResourceFilter sets the resource filter function
func WithResourceFilter(resourceType string) MetricsOption {
	return func(options *MetricsOptions) {
		options.ResourceType = resourceType
	}
}

// WithIncludeDetail sets whether to include detailed information
func WithIncludeDetail(include bool) MetricsOption {
	return func(options *MetricsOptions) {
		options.IncludeDetail = include
	}
}

// WithUnitType sets the unit type (e.g., raw value, percentage, etc.)
func WithUnitType(unitType string) MetricsOption {
	return func(options *MetricsOptions) {
		options.UnitType = unitType
	}
}

// ResourceMetricsOptions stores special options for resource metrics queries
type ResourceMetricsOptions struct {
	// Namespace
	Namespace string
	// Whether to include detailed information
	IncludeDetail bool
	// Resource type (cpu, memory, storage, pods)
	ResourceType string
	// Unit type (raw, percent, human)
	UnitType string
}

// GetNodesMetrics retrieves metrics for all nodes
func GetNodesMetrics(ctx context.Context, client kubernetes.Client, opts ...MetricsOption) ([]models.NodeMetricInfo, error) {
	// Initialize default options
	options := &MetricsOptions{
		SortType: models.SortByCPU,
		Limit:    0,
	}

	// Apply option functions
	for _, opt := range opts {
		opt(options)
	}

	// Prepare query options
	listOptions := metav1.ListOptions{}
	if options.FieldSelector != "" {
		listOptions.FieldSelector = options.FieldSelector
	}
	if options.LabelSelector != "" {
		listOptions.LabelSelector = options.LabelSelector
	}

	// Get node metrics
	nodeMetrics, err := client.GetMetricsClient().MetricsV1beta1().NodeMetricses().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Get node information
	nodes, err := client.ClientSet().CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get node information: %w", err)
	}

	// Build node allocatable resources map
	nodeAllocatable := make(map[string]corev1.ResourceList)
	for _, node := range nodes.Items {
		nodeAllocatable[node.Name] = node.Status.Allocatable
	}

	// Build node metrics information
	var result []models.NodeMetricInfo
	for _, metric := range nodeMetrics.Items {
		allocatable, exists := nodeAllocatable[metric.Name]
		if !exists {
			continue
		}

		nodeMetric := models.BuildNodeMetricInfoFromK8s(metric, allocatable)

		// Apply filters
		if options.NodeFilter != nil && !options.NodeFilter(nodeMetric) {
			continue
		}

		result = append(result, nodeMetric)
	}

	// Sort by specified type
	SortNodeMetrics(result, options.SortType)

	// Limit result count
	if options.Limit > 0 && options.Limit < len(result) {
		result = result[:options.Limit]
	}

	return result, nil
}

// SortNodeMetrics sorts node metrics according to the specified sort type
func SortNodeMetrics(metrics []models.NodeMetricInfo, sortType models.SortType) {
	// If no sort type is specified, default to sorting by CPU usage
	if sortType == "" {
		sortType = models.SortByCPU
	}

	sort.Slice(metrics, func(i, j int) bool {
		switch sortType {
		case models.SortByCPU:
			return metrics[i].CPUUsage > metrics[j].CPUUsage
		case models.SortByMemory:
			return metrics[i].MemoryUsage > metrics[j].MemoryUsage
		case models.SortByCPUPercent:
			return metrics[i].CPUPercent > metrics[j].CPUPercent
		case models.SortByMemoryPercent:
			return metrics[i].MemoryPercent > metrics[j].MemoryPercent
		case models.SortByName:
			return metrics[i].Name < metrics[j].Name
		default:
			// Default to sorting by CPU usage
			return metrics[i].CPUUsage > metrics[j].CPUUsage
		}
	})
}

// GetNodeMetric retrieves metrics for a specific node
func GetNodeMetric(ctx context.Context, client kubernetes.Client, nodeName string) (*models.NodeMetricInfo, error) {
	// Get node metrics
	nodeMetric, err := client.GetMetricsClient().MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for node %s: %w", nodeName, err)
	}

	// Get node information
	node, err := client.ClientSet().CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get information for node %s: %w", nodeName, err)
	}

	metricInfo := models.BuildNodeMetricInfoFromK8s(*nodeMetric, node.Status.Allocatable)
	return &metricInfo, nil
}

// GetPodsMetrics retrieves Pod metrics
func GetPodsMetrics(ctx context.Context, client kubernetes.Client, namespace string, opts ...MetricsOption) ([]models.PodMetricInfo, error) {
	// Initialize default options
	options := &MetricsOptions{
		SortType: models.SortByCPU,
		Limit:    0,
	}

	// Apply option functions
	for _, opt := range opts {
		opt(options)
	}

	// Prepare query options
	listOptions := metav1.ListOptions{}
	if options.FieldSelector != "" {
		listOptions.FieldSelector = options.FieldSelector
	}
	if options.LabelSelector != "" {
		listOptions.LabelSelector = options.LabelSelector
	}

	// Get Pod metrics
	var podMetrics *metricsv1beta1.PodMetricsList
	var err error

	if namespace != "" {
		podMetrics, err = client.GetMetricsClient().MetricsV1beta1().PodMetricses(namespace).List(ctx, listOptions)
	} else {
		podMetrics, err = client.GetMetricsClient().MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(ctx, listOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get Pod metrics: %w", err)
	}

	// Build Pod metrics information
	var result []models.PodMetricInfo
	for _, metric := range podMetrics.Items {
		podMetric := models.BuildPodMetricInfoFromK8s(metric)

		// Apply filters
		if options.PodFilter != nil && !options.PodFilter(podMetric) {
			continue
		}

		result = append(result, podMetric)
	}

	// Sort by specified type
	SortPodMetrics(result, options.SortType)

	// Limit result count
	if options.Limit > 0 && options.Limit < len(result) {
		result = result[:options.Limit]
	}

	return result, nil
}

// SortPodMetrics sorts Pod metrics according to the specified sort type
func SortPodMetrics(metrics []models.PodMetricInfo, sortType models.SortType) {
	// If no sort type is specified, default to sorting by CPU usage
	if sortType == "" {
		sortType = models.SortByCPU
	}

	sort.Slice(metrics, func(i, j int) bool {
		switch sortType {
		case models.SortByCPU:
			return metrics[i].TotalCPU > metrics[j].TotalCPU
		case models.SortByMemory:
			return metrics[i].TotalMemory > metrics[j].TotalMemory
		case models.SortByName:
			// Sort by namespace first, then by name when namespaces are the same
			if metrics[i].Namespace == metrics[j].Namespace {
				return metrics[i].Name < metrics[j].Name
			}
			return metrics[i].Namespace < metrics[j].Namespace
		default:
			// Default to sorting by CPU usage
			return metrics[i].TotalCPU > metrics[j].TotalCPU
		}
	})
}

// GetClusterResourceMetrics retrieves overall cluster resource usage
func GetClusterResourceMetrics(ctx context.Context, client kubernetes.Client, namespace string, opts ...MetricsOption) (*models.ClusterResourceMetrics, error) {
	// Initialize default options
	options := &MetricsOptions{
		ResourceType:  "all",
		IncludeDetail: true,
		UnitType:      "raw",
	}

	// Apply option functions
	for _, opt := range opts {
		opt(options)
	}

	// Prepare query options
	listOptions := metav1.ListOptions{}
	if options.FieldSelector != "" {
		listOptions.FieldSelector = options.FieldSelector
	}
	if options.LabelSelector != "" {
		listOptions.LabelSelector = options.LabelSelector
	}

	// Get node list
	nodes, err := client.ClientSet().CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get node list: %w", err)
	}

	metrics := &models.ClusterResourceMetrics{}

	// Calculate cluster total capacity and allocatable resources
	for _, node := range nodes.Items {
		metrics.CPUCapacity += node.Status.Capacity.Cpu().MilliValue()
		metrics.CPUAllocatable += node.Status.Allocatable.Cpu().MilliValue()

		metrics.MemoryCapacity += node.Status.Capacity.Memory().Value() / (1024 * 1024)
		metrics.MemoryAllocatable += node.Status.Allocatable.Memory().Value() / (1024 * 1024)

		// Try to get storage capacity - note that StorageEphemeral may not exist
		storage := node.Status.Capacity.StorageEphemeral()
		if !storage.IsZero() {
			metrics.StorageCapacity += storage.Value() / (1024 * 1024 * 1024)
		}

		// Try to get allocatable storage - note that StorageEphemeral may not exist
		storage = node.Status.Allocatable.StorageEphemeral()
		if !storage.IsZero() {
			metrics.StorageAllocatable += storage.Value() / (1024 * 1024 * 1024)
		}

		metrics.PodCapacity += node.Status.Capacity.Pods().Value()
	}

	// Get current resource usage
	nodeMetrics, err := client.GetMetricsClient().MetricsV1beta1().NodeMetricses().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	for _, metric := range nodeMetrics.Items {
		metrics.CPUUsage += metric.Usage.Cpu().MilliValue()
		metrics.MemoryUsage += metric.Usage.Memory().Value() / (1024 * 1024)
	}

	// Calculate usage percentages
	if metrics.CPUAllocatable > 0 {
		metrics.CPUPercent = float64(metrics.CPUUsage) / float64(metrics.CPUAllocatable) * 100
	}

	if metrics.MemoryAllocatable > 0 {
		metrics.MemoryPercent = float64(metrics.MemoryUsage) / float64(metrics.MemoryAllocatable) * 100
	}

	// Get current Pod count
	var podOptions metav1.ListOptions = listOptions

	if namespace != "" {
		pods, err := client.ClientSet().CoreV1().Pods(namespace).List(ctx, podOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to get Pod list for namespace %s: %w", namespace, err)
		}
		metrics.RunningPods = len(pods.Items)
		metrics.Namespace = namespace
	} else {
		pods, err := client.ClientSet().CoreV1().Pods(metav1.NamespaceAll).List(ctx, podOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to get Pod list: %w", err)
		}
		metrics.RunningPods = len(pods.Items)
	}

	// Calculate Pod usage percentage
	if metrics.PodCapacity > 0 {
		metrics.PodPercent = float64(metrics.RunningPods) / float64(metrics.PodCapacity) * 100
	}

	// Set additional information
	metrics.ResourceType = options.ResourceType
	metrics.IncludeDetail = options.IncludeDetail
	metrics.UnitType = options.UnitType

	return metrics, nil
}

// ParseSortType parses sort type from a string
func ParseSortType(sortTypeStr string) models.SortType {
	// Convert to lowercase and remove extra spaces
	sortTypeStr = strings.ToLower(strings.TrimSpace(sortTypeStr))

	switch sortTypeStr {
	case "cpu", "cpuusage":
		return models.SortByCPU
	case "memory", "mem", "memoryusage":
		return models.SortByMemory
	case "cpu_percent", "cpupercent":
		return models.SortByCPUPercent
	case "memory_percent", "mempercent", "memorypercent":
		return models.SortByMemoryPercent
	case "name":
		return models.SortByName
	default:
		// Default to sorting by CPU usage
		return models.SortByCPU
	}
}

// FormatResourceValue formats resource values into human-readable form
func FormatResourceValue(resourceName string, value int64) string {
	switch resourceName {
	case "cpu":
		return fmt.Sprintf("%dm", value)
	case "memory":
		return fmt.Sprintf("%dMi", value)
	case "storage":
		return fmt.Sprintf("%dGi", value)
	default:
		return fmt.Sprintf("%d", value)
	}
}

// ParseQuantity parses resource quantity string into int64 value
func ParseQuantity(quantityStr string) (int64, error) {
	quantity, err := resource.ParseQuantity(quantityStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse resource quantity %s: %w", quantityStr, err)
	}
	return quantity.Value(), nil
}
