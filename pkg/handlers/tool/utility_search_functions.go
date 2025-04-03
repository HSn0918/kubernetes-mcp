package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/mark3labs/mcp-go/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

// SearchResources 搜索资源
func (h *UtilityHandler) SearchResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	query, _ := arguments["query"].(string)
	namespacesStr, _ := arguments["namespaces"].(string)
	kindsStr, _ := arguments["kinds"].(string)
	matchLabels, _ := arguments["matchLabels"].(bool)
	matchAnnotations, _ := arguments["matchAnnotations"].(bool)

	h.Log.Info("Searching resources",
		"query", query,
		"namespaces", namespacesStr,
		"kinds", kindsStr,
		"matchLabels", matchLabels,
		"matchAnnotations", matchAnnotations,
	)

	// 解析命名空间列表
	var namespaces []string
	if namespacesStr != "" {
		namespaces = strings.Split(namespacesStr, ",")
		for i := range namespaces {
			namespaces[i] = strings.TrimSpace(namespaces[i])
		}
	}

	// 解析资源类型列表
	var kinds []string
	if kindsStr != "" {
		kinds = strings.Split(kindsStr, ",")
		for i := range kinds {
			kinds[i] = strings.TrimSpace(kinds[i])
		}
	}

	// 如果没有指定命名空间，获取所有命名空间
	if len(namespaces) == 0 || (len(namespaces) == 1 && namespaces[0] == "all") {
		nsList := &corev1.NamespaceList{}
		err := h.Client.List(ctx, nsList)
		if err != nil {
			h.Log.Error("Failed to list namespaces", "error", err)
			return nil, fmt.Errorf("failed to list namespaces: %w", err)
		}
		namespaces = make([]string, 0, len(nsList.Items))
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// 获取API资源列表
	_, resourcesList, err := h.Client.GetDiscoveryClient().ServerGroupsAndResources()
	if err != nil {
		// 处理部分发现错误，继续使用已获取的资源
		if !discovery.IsGroupDiscoveryFailedError(err) {
			h.Log.Error("Failed to get API resources", "error", err)
			return nil, fmt.Errorf("failed to get API resources: %w", err)
		}
		h.Log.Warn("Partial API discovery error", "error", err)
	}

	// 根据请求筛选需要搜索的资源类型
	matchingResourcesList := make(map[string][]metav1.APIResource)
	for _, resList := range resourcesList {
		for _, res := range resList.APIResources {
			// 跳过子资源
			if strings.Contains(res.Name, "/") {
				continue
			}
			// 检查是否有list权限
			if !hasListVerb(res.Verbs) {
				continue
			}
			// 如果指定了kinds，则只搜索指定的kinds
			if len(kinds) > 0 {
				found := false
				for _, k := range kinds {
					if strings.EqualFold(res.Kind, k) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			matchingResourcesList[resList.GroupVersion] = append(matchingResourcesList[resList.GroupVersion], res)
		}
	}

	// 使用models.SearchResult替代本地定义的结构体
	var results []models.SearchResult

	// 遍历所有资源类型和命名空间，查找匹配的资源
	totalSearched := 0
	for groupVersion, resources := range matchingResourcesList {
		for _, resource := range resources {
			// 检查资源作用域
			isNamespaced := resource.Namespaced

			// 对于非命名空间资源，只搜索全局范围
			if !isNamespaced {
				rs, err := searchResourcesInNamespace(ctx, h, groupVersion, resource, query, "", matchLabels, matchAnnotations)
				if err != nil {
					h.Log.Error("Failed to search resources", "error", err, "groupVersion", groupVersion, "resource", resource.Name)
					continue
				}
				// 添加到结果中
				for _, r := range rs {
					results = append(results, models.SearchResult{
						Kind:         r.Kind,
						APIVersion:   r.APIVersion,
						Name:         r.Name,
						Namespace:    r.Namespace,
						Labels:       r.Labels,
						Annotations:  r.Annotations,
						MatchedBy:    r.MatchedBy,
						MatchedValue: r.MatchedValue,
						CreationTime: r.CreationTime,
					})
				}
				totalSearched++
				continue
			}

			// 对于命名空间资源，在所有指定的命名空间中搜索
			for _, ns := range namespaces {
				rs, err := searchResourcesInNamespace(ctx, h, groupVersion, resource, query, ns, matchLabels, matchAnnotations)
				if err != nil {
					h.Log.Error("Failed to search resources", "error", err, "namespace", ns, "groupVersion", groupVersion, "resource", resource.Name)
					continue
				}
				// 添加到结果中
				for _, r := range rs {
					results = append(results, models.SearchResult{
						Kind:         r.Kind,
						APIVersion:   r.APIVersion,
						Name:         r.Name,
						Namespace:    r.Namespace,
						Labels:       r.Labels,
						Annotations:  r.Annotations,
						MatchedBy:    r.MatchedBy,
						MatchedValue: r.MatchedValue,
						CreationTime: r.CreationTime,
					})
				}
				totalSearched++
			}
		}
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search Results for '%s':\n\n", query))
	result.WriteString(fmt.Sprintf("Found %d matching resources across %d resource types\n\n", len(results), totalSearched))

	// 按照种类和名称排序
	sort.Slice(results, func(i, j int) bool {
		if results[i].Kind != results[j].Kind {
			return results[i].Kind < results[j].Kind
		}
		if results[i].Namespace != results[j].Namespace {
			return results[i].Namespace < results[j].Namespace
		}
		return results[i].Name < results[j].Name
	})

	// 按照资源类型分组显示结果
	currentKind := ""
	for _, res := range results {
		if res.Kind != currentKind {
			if currentKind != "" {
				result.WriteString("\n")
			}
			result.WriteString(fmt.Sprintf("== %s ==\n", res.Kind))
			currentKind = res.Kind
		}

		if res.Namespace != "" {
			result.WriteString(fmt.Sprintf("- %s (namespace: %s)", res.Name, res.Namespace))
		} else {
			result.WriteString(fmt.Sprintf("- %s (cluster-scoped)", res.Name))
		}

		result.WriteString(fmt.Sprintf(", matched by: %s", res.MatchedBy))
		result.WriteString("\n")
	}

	if len(results) == 0 {
		result.WriteString("No resources found matching the query.\n")
	}

	// 创建完整的搜索结果模型
	searchResults := models.SearchResults{
		Items:       results,
		SearchQuery: query,
		TotalCount:  len(results),
		TypesCount:  totalSearched,
	}

	// 序列化为JSON
	resultsJSON, err := json.Marshal(searchResults)
	if err != nil {
		h.Log.Error("Failed to marshal search results", "error", err)
		// 继续执行，只返回文本格式
	} else {
		// 添加JSON格式数据
		result.WriteString("\nJSON格式数据:\n")
		result.WriteString(string(resultsJSON))
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
