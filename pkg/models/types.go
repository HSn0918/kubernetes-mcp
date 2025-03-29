package models

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceSummary 资源摘要信息
type ResourceSummary struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Kind       string            `json:"kind"`
	APIVersion string            `json:"apiVersion"`
	Labels     map[string]string `json:"labels,omitempty"`
	Created    time.Time         `json:"created"`
}

// ResourceListResult 资源列表结果
type ResourceListResult struct {
	Items      []ResourceSummary `json:"items"`
	TotalCount int               `json:"totalCount"`
}

// NamespaceSummary 命名空间摘要信息
type NamespaceSummary struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Active bool   `json:"active"`
}

// NamespaceListResult 命名空间列表结果
type NamespaceListResult struct {
	Items      []NamespaceSummary `json:"items"`
	TotalCount int                `json:"totalCount"`
}

// GroupVersionKindName 资源的完整标识
type GroupVersionKindName struct {
	GVK  schema.GroupVersionKind
	Name string
}

// ResourceRequest 资源请求参数
type ResourceRequest struct {
	Group     string
	Version   string
	Kind      string
	Name      string
	Namespace string
}

// NewResourceRequest 从参数创建资源请求
func NewResourceRequest(apiVersion, kind, name, namespace string) ResourceRequest {
	group, version := parseAPIVersion(apiVersion)
	return ResourceRequest{
		Group:     group,
		Version:   version,
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}
}

// ToGVK 转换为GroupVersionKind
func (r ResourceRequest) ToGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

// ToObjectKey 转换为Object Key
func (r ResourceRequest) ToObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

// 解析API版本为组和版本
func parseAPIVersion(apiVersion string) (string, string) {
	parts := strings.Split(apiVersion, "/")
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}
