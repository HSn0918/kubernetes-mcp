package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// formatTimeAgo 格式化事件的时间，显示为相对时间
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		seconds := int(diff.Seconds())
		return fmt.Sprintf("%d second%s ago", seconds, pluralSuffix(seconds))
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralSuffix(minutes))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralSuffix(hours))
	} else {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralSuffix(days))
	}
}

// pluralSuffix 根据数量返回复数后缀
func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// hasListVerb 检查资源是否有列表权限
func hasListVerb(verbs []string) bool {
	for _, verb := range verbs {
		if verb == "list" {
			return true
		}
	}
	return false
}

// parseGroup 从GroupVersion字符串解析Group
func parseGroup(groupVersion string) string {
	parts := strings.Split(groupVersion, "/")
	if len(parts) == 1 {
		return ""
	}
	return parts[0]
}

// parseVersion 从GroupVersion字符串解析Version
func parseVersion(groupVersion string) string {
	parts := strings.Split(groupVersion, "/")
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[1]
}

// searchResourcesInNamespace 在特定命名空间中搜索指定资源类型
func searchResourcesInNamespace(
	ctx context.Context,
	h *UtilityHandler,
	groupVersion string,
	resource metav1.APIResource,
	query string,
	namespace string,
	matchLabels bool,
	matchAnnotations bool,
) ([]models.SearchResult, error) {
	// 创建列表对象
	obj := &unstructured.UnstructuredList{}

	// 列出资源
	dynamicList, err := h.Client.GetDynamicClient().Resource(
		schema.GroupVersionResource{
			Group:    parseGroup(groupVersion),
			Version:  parseVersion(groupVersion),
			Resource: resource.Name,
		}).Namespace(namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	obj.Items = dynamicList.Items

	queryLower := strings.ToLower(query)
	var results []models.SearchResult

	// 遍历资源，检查是否匹配查询
	for _, item := range obj.Items {
		// 匹配名称
		name := item.GetName()
		if strings.Contains(strings.ToLower(name), queryLower) {
			results = append(results, models.SearchResult{
				Kind:         resource.Kind,
				APIVersion:   groupVersion,
				Name:         name,
				Namespace:    namespace,
				MatchedBy:    "name",
				MatchedValue: name,
			})
			continue
		}

		// 匹配标签
		if matchLabels {
			labels := item.GetLabels()
			for k, v := range labels {
				labelMatch := strings.Contains(strings.ToLower(k), queryLower) ||
					strings.Contains(strings.ToLower(v), queryLower)
				if labelMatch {
					results = append(results, models.SearchResult{
						Kind:         resource.Kind,
						APIVersion:   groupVersion,
						Name:         name,
						Namespace:    namespace,
						Labels:       fmt.Sprintf("%v", labels),
						MatchedBy:    "label",
						MatchedValue: fmt.Sprintf("%s=%s", k, v),
					})
					break
				}
			}
		}

		// 匹配注解
		if matchAnnotations {
			annotations := item.GetAnnotations()
			for k, v := range annotations {
				annotationMatch := strings.Contains(strings.ToLower(k), queryLower) ||
					strings.Contains(strings.ToLower(v), queryLower)
				if annotationMatch {
					results = append(results, models.SearchResult{
						Kind:         resource.Kind,
						APIVersion:   groupVersion,
						Name:         name,
						Namespace:    namespace,
						Annotations:  fmt.Sprintf("%v", annotations),
						MatchedBy:    "annotation",
						MatchedValue: fmt.Sprintf("%s=%s", k, v),
					})
					break
				}
			}
		}
	}

	return results, nil
}
