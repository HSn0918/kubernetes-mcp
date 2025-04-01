package models

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// SortType defines the type for sorting metrics
type SortType string

const (
	// SortByCPU sorts by CPU usage
	SortByCPU SortType = "cpu"
	// SortByMemory sorts by memory usage
	SortByMemory SortType = "memory"
	// SortByCPUPercent sorts by CPU usage percentage
	SortByCPUPercent SortType = "cpu_percent"
	// SortByMemoryPercent sorts by memory usage percentage
	SortByMemoryPercent SortType = "memory_percent"
	// SortByName sorts by resource name
	SortByName SortType = "name"
)

// FilterOptions defines options for filtering resources
type FilterOptions struct {
	// FieldSelector is the Kubernetes field selector
	FieldSelector string
	// LabelSelector is the Kubernetes label selector
	LabelSelector string
	// Limit is the maximum number of results to return
	Limit int
	// SortBy specifies the field to sort by
	SortBy SortType
	// Namespace to filter resources by
	Namespace string
	// ResourceName to filter by specific resource name
	ResourceName string
}

// NodeMetricInfo holds node resource metrics
type NodeMetricInfo struct {
	// Node name
	Name string
	// CPU usage in millicores
	CPUUsage int64
	// CPU allocatable in millicores
	CPUAllocatable int64
	// CPU usage percentage
	CPUPercent float64
	// Memory usage in MB
	MemoryUsage int64
	// Memory allocatable in MB
	MemoryAllocatable int64
	// Memory usage percentage
	MemoryPercent float64
	// Metric timestamp
	Timestamp time.Time
}

// PodMetricInfo holds pod resource metrics
type PodMetricInfo struct {
	// Pod name
	Name string
	// Pod namespace
	Namespace string
	// Total CPU usage in millicores
	TotalCPU int64
	// Total memory usage in MB
	TotalMemory int64
	// Container metrics
	Containers []ContainerMetricInfo
	// Metric timestamp
	Timestamp time.Time
}

// ContainerMetricInfo holds container resource metrics
type ContainerMetricInfo struct {
	// Container name
	Name string
	// CPU usage in millicores
	CPUUsage int64
	// Memory usage in MB
	MemoryUsage int64
}

// ClusterResourceMetrics holds overall cluster resource usage
type ClusterResourceMetrics struct {
	// CPU capacity in millicores
	CPUCapacity int64
	// CPU allocatable in millicores
	CPUAllocatable int64
	// Current CPU usage in millicores
	CPUUsage int64
	// CPU usage percentage
	CPUPercent float64

	// Memory capacity in MB
	MemoryCapacity int64
	// Memory allocatable in MB
	MemoryAllocatable int64
	// Current memory usage in MB
	MemoryUsage int64
	// Memory usage percentage
	MemoryPercent float64

	// Storage capacity in GB
	StorageCapacity int64
	// Storage allocatable in GB
	StorageAllocatable int64

	// Maximum pods
	PodCapacity int64
	// Current running pods count
	RunningPods int
	// Pod usage percentage
	PodPercent float64

	// Namespace filter (if specified)
	Namespace string
	// Whether to include detailed information
	IncludeDetail bool
	// Resource type being queried
	ResourceType string
	// Unit type (raw, percent, human)
	UnitType string
}

// BuildNodeMetricInfoFromK8s constructs NodeMetricInfo from Kubernetes API data
func BuildNodeMetricInfoFromK8s(nodeMetric metricsv1beta1.NodeMetrics, allocatable corev1.ResourceList) NodeMetricInfo {
	cpuUsage := nodeMetric.Usage.Cpu().MilliValue()
	cpuAllocatable := allocatable.Cpu().MilliValue()
	memoryUsage := nodeMetric.Usage.Memory().Value() / (1024 * 1024) // Convert to MB
	memoryAllocatable := allocatable.Memory().Value() / (1024 * 1024)

	// Calculate usage percentages
	cpuPercent := float64(0)
	if cpuAllocatable > 0 {
		cpuPercent = float64(cpuUsage) / float64(cpuAllocatable) * 100
	}

	memoryPercent := float64(0)
	if memoryAllocatable > 0 {
		memoryPercent = float64(memoryUsage) / float64(memoryAllocatable) * 100
	}

	return NodeMetricInfo{
		Name:              nodeMetric.Name,
		CPUUsage:          cpuUsage,
		CPUAllocatable:    cpuAllocatable,
		CPUPercent:        cpuPercent,
		MemoryUsage:       memoryUsage,
		MemoryAllocatable: memoryAllocatable,
		MemoryPercent:     memoryPercent,
		Timestamp:         nodeMetric.Timestamp.Time,
	}
}

// BuildPodMetricInfoFromK8s constructs PodMetricInfo from Kubernetes API data
func BuildPodMetricInfoFromK8s(podMetric metricsv1beta1.PodMetrics) PodMetricInfo {
	result := PodMetricInfo{
		Name:       podMetric.Name,
		Namespace:  podMetric.Namespace,
		Containers: make([]ContainerMetricInfo, 0, len(podMetric.Containers)),
		Timestamp:  podMetric.Timestamp.Time,
	}

	// Aggregate container metrics
	for _, container := range podMetric.Containers {
		containerCPU := container.Usage.Cpu().MilliValue()
		containerMemory := container.Usage.Memory().Value() / (1024 * 1024) // Convert to MB

		result.TotalCPU += containerCPU
		result.TotalMemory += containerMemory

		result.Containers = append(result.Containers, ContainerMetricInfo{
			Name:        container.Name,
			CPUUsage:    containerCPU,
			MemoryUsage: containerMemory,
		})
	}

	return result
}
