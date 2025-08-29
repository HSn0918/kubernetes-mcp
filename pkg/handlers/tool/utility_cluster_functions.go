package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

// GetClusterInfo 获取集群信息
func (h *UtilityHandler) GetClusterInfo(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	h.Log.Info("Getting cluster info")

	// 构建响应
	var result strings.Builder
	result.WriteString("Kubernetes Cluster Information:\n\n")

	// 获取服务器版本信息
	versionInfo, err := h.Client.GetDiscoveryClient().ServerVersion()
	if err != nil {
		h.Log.Error("Failed to get server version", "error", err)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to get server version: %v", err)), nil
	}

	// 添加版本信息
	result.WriteString(fmt.Sprintf("Version:      %s\n", versionInfo.GitVersion))
	result.WriteString(fmt.Sprintf("Build Date:   %s\n", versionInfo.BuildDate))
	result.WriteString(fmt.Sprintf("Go Version:   %s\n", versionInfo.GoVersion))
	result.WriteString(fmt.Sprintf("Platform:     %s\n", versionInfo.Platform))
	result.WriteString(fmt.Sprintf("Git Commit:   %s\n", versionInfo.GitCommit))
	result.WriteString(fmt.Sprintf("Git TreeState: %s\n", versionInfo.GitTreeState))
	result.WriteString(fmt.Sprintf("Compiler:     %s\n", versionInfo.Compiler))

	// 获取当前命名空间
	currentNamespace, err := h.Client.GetCurrentNamespace()
	if err == nil && currentNamespace != "" {
		result.WriteString(fmt.Sprintf("\nCurrent Namespace: %s\n", currentNamespace))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// GetAPIResources 获取API资源列表
func (h *UtilityHandler) GetAPIResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	group, _ := arguments["group"].(string)

	h.Log.Info("Getting API resources", "group", group)

	// 构建响应
	var result strings.Builder
	result.WriteString("API Resources:\n\n")

	// 获取API资源
	var resourcesList []*metav1.APIResourceList
	var err error

	// 根据是否指定了group来获取资源
	if group == "" {
		// 获取所有API组的资源
		_, resourcesList, err = h.Client.GetDiscoveryClient().ServerGroupsAndResources()
		if err != nil {
			// 处理部分发现错误，继续使用已获取的资源
			if !discovery.IsGroupDiscoveryFailedError(err) {
				h.Log.Error("Failed to get API resources", "error", err)
				return utils.NewErrorToolResult(fmt.Sprintf("failed to get API resources: %v", err)), nil
			}
			h.Log.Warn("Partial API discovery error", "error", err)
		}
	} else {
		// 获取特定组的资源列表
		apiGroup, err := h.Client.GetDiscoveryClient().ServerResourcesForGroupVersion(group)
		if err != nil {
			h.Log.Error("Failed to get API resources for group", "group", group, "error", err)
			return utils.NewErrorToolResult(fmt.Sprintf("failed to get API resources for group %s: %v", group, err)), nil
		}
		resourcesList = []*metav1.APIResourceList{apiGroup}
	}

	// 格式化输出
	if len(resourcesList) == 0 {
		result.WriteString("No API resources found\n")
	} else {
		// 对API组进行排序
		sort.Slice(resourcesList, func(i, j int) bool {
			return resourcesList[i].GroupVersion < resourcesList[j].GroupVersion
		})

		// 遍历每个API组
		for _, apiResourceList := range resourcesList {
			gv := apiResourceList.GroupVersion
			result.WriteString(fmt.Sprintf("GROUP VERSION: %s\n", gv))

			// 对资源进行排序
			resources := apiResourceList.APIResources
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Name < resources[j].Name
			})

			// 遍历每个资源
			for _, resource := range resources {
				// 跳过子资源
				if strings.Contains(resource.Name, "/") {
					continue
				}

				namespaced := "namespaced"
				if !resource.Namespaced {
					namespaced = "cluster-wide"
				}

				verbs := strings.Join(resource.Verbs, ",")
				result.WriteString(fmt.Sprintf("  %-40s %-15s %-30s\n", resource.Name, namespaced, verbs))
			}
			result.WriteString("\n")
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}
