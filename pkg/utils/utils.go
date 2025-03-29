package utils

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ParseGVK 解析API版本和Kind并返回GroupVersionKind
func ParseGVK(apiVersion string, kind string) schema.GroupVersionKind {
	parts := strings.Split(apiVersion, "/")
	var group, version string
	if len(parts) == 1 {
		group = ""
		version = parts[0]
	} else {
		group = parts[0]
		version = parts[1]
	}

	return schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}
