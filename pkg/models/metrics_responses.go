package models

import "time"

// NodeResponse represents the API response for node metrics
type NodeResponse struct {
	Name              string    `json:"name"`
	CPUUsage          int64     `json:"cpuUsage"`
	CPUAllocatable    int64     `json:"cpuAllocatable"`
	CPUPercent        float64   `json:"cpuPercent"`
	MemoryUsage       int64     `json:"memoryUsage"`
	MemoryAllocatable int64     `json:"memoryAllocatable"`
	MemoryPercent     float64   `json:"memoryPercent"`
	Timestamp         time.Time `json:"timestamp"`
	UpdatedAgo        string    `json:"updatedAgo"`
}

// NodesListResponse represents the API response for a list of node metrics
type NodesListResponse struct {
	Nodes      []NodeResponse `json:"nodes"`
	SortBy     string         `json:"sortBy"`
	TotalCount int            `json:"totalCount"`
}

// ContainerResponse represents the API response for container metrics
type ContainerResponse struct {
	Name        string `json:"name"`
	CPUUsage    int64  `json:"cpuUsage"`
	MemoryUsage int64  `json:"memoryUsage"`
}

// PodResponse represents the API response for pod metrics
type PodResponse struct {
	Name        string              `json:"name"`
	Namespace   string              `json:"namespace"`
	TotalCPU    int64               `json:"totalCpu"`
	TotalMemory int64               `json:"totalMemory"`
	Timestamp   time.Time           `json:"timestamp"`
	UpdatedAgo  string              `json:"updatedAgo"`
	Containers  []ContainerResponse `json:"containers,omitempty"`
}

// PodsListResponse represents the API response for a list of pod metrics
type PodsListResponse struct {
	Pods          []PodResponse `json:"pods"`
	SortBy        string        `json:"sortBy"`
	TotalCount    int           `json:"totalCount"`
	Namespace     string        `json:"namespace,omitempty"`
	Limit         int           `json:"limit"`
	IncludeDetail bool          `json:"includeDetail"`
}

// ResourceMetricsResponse represents the API response for resource metrics
type ResourceMetricsResponse struct {
	ResourceType   string  `json:"resourceType"`
	CPUCapacity    int64   `json:"cpuCapacity,omitempty"`
	CPUAllocatable int64   `json:"cpuAllocatable,omitempty"`
	CPUUsage       int64   `json:"cpuUsage,omitempty"`
	CPUPercent     float64 `json:"cpuPercent,omitempty"`
	CPUAvailable   int64   `json:"cpuAvailable,omitempty"`

	MemoryCapacity    int64   `json:"memoryCapacity,omitempty"`
	MemoryAllocatable int64   `json:"memoryAllocatable,omitempty"`
	MemoryUsage       int64   `json:"memoryUsage,omitempty"`
	MemoryPercent     float64 `json:"memoryPercent,omitempty"`
	MemoryAvailable   int64   `json:"memoryAvailable,omitempty"`

	StorageCapacity    int64 `json:"storageCapacity,omitempty"`
	StorageAllocatable int64 `json:"storageAllocatable,omitempty"`

	PodCapacity   int64   `json:"podCapacity,omitempty"`
	RunningPods   int     `json:"runningPods,omitempty"`
	PodPercent    float64 `json:"podPercent,omitempty"`
	PodsAvailable int64   `json:"podsAvailable,omitempty"`

	Namespace string `json:"namespace,omitempty"`
	UnitType  string `json:"unitType"`
}

// TopConsumerResponse represents the API response for top resource consumers
type TopConsumerResponse struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Usage      int64     `json:"usage"`
	Timestamp  time.Time `json:"timestamp"`
	UpdatedAgo string    `json:"updatedAgo"`
}

// TopConsumersListResponse represents the API response for a list of top resource consumers
type TopConsumersListResponse struct {
	Consumers    []TopConsumerResponse `json:"consumers"`
	ResourceType string                `json:"resourceType"`
	Limit        int                   `json:"limit"`
	Namespace    string                `json:"namespace,omitempty"`
	TotalCount   int                   `json:"totalCount"`
}
