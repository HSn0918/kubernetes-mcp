package models

import "time"

// NodeInfo 定义节点信息结构
type NodeInfo struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	KubeletVersion    string            `json:"kubeletVersion"`
	OSImage           string            `json:"osImage"`
	KernelVersion     string            `json:"kernelVersion"`
	Architecture      string            `json:"architecture"`
	InternalIP        string            `json:"internalIP,omitempty"`
	ExternalIP        string            `json:"externalIP,omitempty"`
	Roles             []string          `json:"roles,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Taints            []Taint           `json:"taints,omitempty"`
	AllocatableMemory string            `json:"allocatableMemory,omitempty"`
	AllocatableCPU    string            `json:"allocatableCPU,omitempty"`
	AllocatablePods   string            `json:"allocatablePods,omitempty"`
	CreationTime      time.Time         `json:"creationTime"`
}

// Taint 定义节点污点结构
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"`
}

// NodeListResponse 定义节点列表响应结构
type NodeListResponse struct {
	Count       int        `json:"count"`
	Nodes       []NodeInfo `json:"nodes"`
	RetrievedAt time.Time  `json:"retrievedAt"`
}

// NamespaceInfo 定义命名空间信息结构
type NamespaceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
}

// NamespaceListResponse 定义命名空间列表响应结构
type NamespaceListResponse struct {
	Count       int             `json:"count"`
	Namespaces  []NamespaceInfo `json:"namespaces"`
	RetrievedAt time.Time       `json:"retrievedAt"`
}

// ResourceInfo 定义通用资源信息结构
type ResourceInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace,omitempty"`
	Kind         string            `json:"kind"`
	APIVersion   string            `json:"apiVersion"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
}

// ResourceListResponse 定义通用资源列表响应结构
type ResourceListResponse struct {
	Count       int            `json:"count"`
	Kind        string         `json:"kind"`
	APIVersion  string         `json:"apiVersion"`
	Namespace   string         `json:"namespace,omitempty"`
	Resources   []ResourceInfo `json:"resources"`
	RetrievedAt time.Time      `json:"retrievedAt"`
}
