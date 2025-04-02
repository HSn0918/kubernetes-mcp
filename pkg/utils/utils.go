package utils

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewErrorToolResult 创建一个表示错误的CallToolResult
// 这将IsError设置为true，并将提供的错误消息添加到Content中，而不是返回error对象
func NewErrorToolResult(errMsg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: errMsg,
			},
		},
		IsError: true,
	}
}

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
