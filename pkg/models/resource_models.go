package models

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

// ResourceDescription 表示资源的详细描述信息
type ResourceDescription struct {
	// 基本信息
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Kind       string    `json:"kind"`
	APIVersion string    `json:"apiVersion"`
	CreatedAt  time.Time `json:"createdAt"`

	// 元数据
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion string            `json:"resourceVersion"`
	UID             string            `json:"uid"`

	// 规格和状态
	Spec   map[string]interface{} `json:"spec,omitempty"`
	Status map[string]interface{} `json:"status,omitempty"`

	// 检索时间
	RetrievedAt time.Time `json:"retrievedAt"`
}

// NewResourceDescriptionFromUnstructured 从 unstructured.Unstructured 创建 ResourceDescription
func NewResourceDescriptionFromUnstructured(obj *unstructured.Unstructured) ResourceDescription {
	desc := ResourceDescription{
		Name:            obj.GetName(),
		Namespace:       obj.GetNamespace(),
		Kind:            obj.GetKind(),
		APIVersion:      obj.GetAPIVersion(),
		CreatedAt:       obj.GetCreationTimestamp().Time,
		ResourceVersion: obj.GetResourceVersion(),
		UID:             string(obj.GetUID()),
		RetrievedAt:     time.Now(),
	}

	// 添加标签
	if labels := obj.GetLabels(); len(labels) > 0 {
		desc.Labels = labels
	}

	// 添加注解
	if annotations := obj.GetAnnotations(); len(annotations) > 0 {
		desc.Annotations = annotations
	}

	// 获取spec和status
	unstructContent := obj.UnstructuredContent()
	if spec, found, _ := unstructured.NestedMap(unstructContent, "spec"); found && spec != nil {
		desc.Spec = spec
	}
	if status, found, _ := unstructured.NestedMap(unstructContent, "status"); found && status != nil {
		desc.Status = status
	}

	return desc
}
