package tool

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

type searchQueryMode string

const (
	searchModeGeneric    searchQueryMode = "generic"
	searchModeName       searchQueryMode = "name"
	searchModeLabel      searchQueryMode = "label"
	searchModeAnnotation searchQueryMode = "annotation"
)

type parsedSearchQuery struct {
	mode     searchQueryMode
	raw      string
	pattern  string
	key      string
	value    string
	wildcard bool
}

func parseSearchQuery(raw string) (parsedSearchQuery, error) {
	query := strings.TrimSpace(raw)
	if query == "" {
		return parsedSearchQuery{
			mode:     searchModeName,
			raw:      raw,
			pattern:  "*",
			wildcard: true,
		}, nil
	}

	lower := strings.ToLower(query)
	switch {
	case strings.HasPrefix(lower, "name="):
		pattern := strings.TrimSpace(query[len("name="):])
		if pattern == "" {
			pattern = "*"
		}
		return parsedSearchQuery{
			mode:     searchModeName,
			raw:      raw,
			pattern:  pattern,
			wildcard: hasWildcard(pattern),
		}, nil

	case strings.HasPrefix(lower, "label="):
		key, value, err := parseMapQuery(strings.TrimSpace(query[len("label="):]))
		if err != nil {
			return parsedSearchQuery{}, err
		}
		return parsedSearchQuery{
			mode:  searchModeLabel,
			raw:   raw,
			key:   key,
			value: value,
		}, nil

	case strings.HasPrefix(lower, "annotation="):
		key, value, err := parseMapQuery(strings.TrimSpace(query[len("annotation="):]))
		if err != nil {
			return parsedSearchQuery{}, err
		}
		return parsedSearchQuery{
			mode:  searchModeAnnotation,
			raw:   raw,
			key:   key,
			value: value,
		}, nil

	default:
		return parsedSearchQuery{
			mode:     searchModeGeneric,
			raw:      raw,
			pattern:  query,
			wildcard: hasWildcard(query),
		}, nil
	}
}

func parseMapQuery(raw string) (string, string, error) {
	if raw == "" {
		return "", "", fmt.Errorf("query must not be empty")
	}

	parts := strings.SplitN(raw, ":", 2)
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", fmt.Errorf("query key must not be empty")
	}

	value := "*"
	if len(parts) == 2 {
		v := strings.TrimSpace(parts[1])
		if v != "" {
			value = v
		}
	}
	return key, value, nil
}

func hasWildcard(pattern string) bool {
	return strings.ContainsAny(pattern, "*?")
}

func wildcardMatch(pattern, value string) bool {
	p := strings.ToLower(strings.TrimSpace(pattern))
	v := strings.ToLower(value)
	if p == "" || p == "*" {
		return true
	}

	re := "^" + regexp.QuoteMeta(p) + "$"
	re = strings.ReplaceAll(re, `\*`, ".*")
	re = strings.ReplaceAll(re, `\?`, ".")
	matched, err := regexp.MatchString(re, v)
	if err != nil {
		return false
	}
	return matched
}

func stringMatch(pattern, candidate string, wildcard bool) bool {
	if wildcard {
		return wildcardMatch(pattern, candidate)
	}
	return strings.Contains(strings.ToLower(candidate), strings.ToLower(pattern))
}

func mapMatch(m map[string]string, keyPattern, valuePattern string) (bool, string) {
	keyWildcard := hasWildcard(keyPattern)
	valueWildcard := hasWildcard(valuePattern)
	for k, v := range m {
		if !stringMatch(keyPattern, k, keyWildcard) {
			continue
		}
		if !stringMatch(valuePattern, v, valueWildcard) {
			continue
		}
		return true, fmt.Sprintf("%s=%s", k, v)
	}
	return false, ""
}

// searchResourcesInNamespace 在特定命名空间中搜索指定资源类型
func searchResourcesInNamespace(
	ctx context.Context,
	h *UtilityHandler,
	groupVersion string,
	resource metav1.APIResource,
	searchQuery parsedSearchQuery,
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

	var results []models.SearchResult

	// 遍历资源，检查是否匹配查询
	for _, item := range obj.Items {
		name := item.GetName()
		switch searchQuery.mode {
		case searchModeName:
			if !stringMatch(searchQuery.pattern, name, searchQuery.wildcard) {
				continue
			}
			results = append(results, models.SearchResult{
				Kind:         resource.Kind,
				APIVersion:   groupVersion,
				Name:         name,
				Namespace:    namespace,
				MatchedBy:    "name",
				MatchedValue: name,
			})
			continue

		case searchModeLabel:
			labels := item.GetLabels()
			if ok, matchedValue := mapMatch(labels, searchQuery.key, searchQuery.value); ok {
				results = append(results, models.SearchResult{
					Kind:         resource.Kind,
					APIVersion:   groupVersion,
					Name:         name,
					Namespace:    namespace,
					Labels:       fmt.Sprintf("%v", labels),
					MatchedBy:    "label",
					MatchedValue: matchedValue,
				})
			}
			continue

		case searchModeAnnotation:
			annotations := item.GetAnnotations()
			if ok, matchedValue := mapMatch(annotations, searchQuery.key, searchQuery.value); ok {
				results = append(results, models.SearchResult{
					Kind:         resource.Kind,
					APIVersion:   groupVersion,
					Name:         name,
					Namespace:    namespace,
					Annotations:  fmt.Sprintf("%v", annotations),
					MatchedBy:    "annotation",
					MatchedValue: matchedValue,
				})
			}
			continue
		}

		if stringMatch(searchQuery.pattern, name, searchQuery.wildcard) {
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

		if matchLabels {
			labels := item.GetLabels()
			for k, v := range labels {
				if stringMatch(searchQuery.pattern, k, searchQuery.wildcard) ||
					stringMatch(searchQuery.pattern, v, searchQuery.wildcard) {
					results = append(results, models.SearchResult{
						Kind:         resource.Kind,
						APIVersion:   groupVersion,
						Name:         name,
						Namespace:    namespace,
						Labels:       fmt.Sprintf("%v", labels),
						MatchedBy:    "label",
						MatchedValue: fmt.Sprintf("%s=%s", k, v),
					})
					goto nextItem
				}
			}
		}

		if matchAnnotations {
			annotations := item.GetAnnotations()
			for k, v := range annotations {
				if stringMatch(searchQuery.pattern, k, searchQuery.wildcard) ||
					stringMatch(searchQuery.pattern, v, searchQuery.wildcard) {
					results = append(results, models.SearchResult{
						Kind:         resource.Kind,
						APIVersion:   groupVersion,
						Name:         name,
						Namespace:    namespace,
						Annotations:  fmt.Sprintf("%v", annotations),
						MatchedBy:    "annotation",
						MatchedValue: fmt.Sprintf("%s=%s", k, v),
					})
					goto nextItem
				}
			}
		}

	nextItem:
	}

	return results, nil
}
