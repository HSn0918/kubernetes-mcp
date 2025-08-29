package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ExplainResource 解释资源结构
func (h *UtilityHandler) ExplainResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	field, _ := arguments["field"].(string)
	recursive, _ := arguments["recursive"].(bool)

	h.Log.Info("Explaining resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"field", field,
		"recursive", recursive,
	)

	// 构建参数
	group, version := parseGroup(apiVersion), parseVersion(apiVersion)

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Resource Structure for %s (%s):\n\n", kind, apiVersion))

	// 从discovery客户端获取资源定义
	_, resources, err := h.Client.GetDiscoveryClient().ServerGroupsAndResources()
	if err != nil {
		if !discovery.IsGroupDiscoveryFailedError(err) {
			h.Log.Error("Failed to get API resources", "error", err)
			return nil, fmt.Errorf("failed to get API resources: %w", err)
		}
		h.Log.Warn("Partial API discovery error", "error", err)
	}

	// 查找特定的资源定义
	var targetResource *metav1.APIResource
	targetGroupVersion := ""
	if group == "" {
		targetGroupVersion = version
	} else {
		targetGroupVersion = group + "/" + version
	}

	for _, resList := range resources {
		if resList.GroupVersion == targetGroupVersion {
			for _, res := range resList.APIResources {
				if strings.EqualFold(res.Kind, kind) {
					targetResource = &res
					break
				}
			}
			if targetResource != nil {
				break
			}
		}
	}

	if targetResource == nil {
		result.WriteString(fmt.Sprintf("Resource %s with apiVersion %s not found in the cluster.\n", kind, apiVersion))
	} else {
		// 显示资源基本信息
		result.WriteString(fmt.Sprintf("KIND:         %s\n", targetResource.Kind))
		result.WriteString(fmt.Sprintf("API VERSION:  %s\n", apiVersion))
		result.WriteString(fmt.Sprintf("RESOURCE:     %s\n", targetResource.Name))
		result.WriteString(fmt.Sprintf("SCOPE:        %s\n", getScopeText(targetResource.Namespaced)))
		result.WriteString(fmt.Sprintf("VERBS:        %s\n", strings.Join(targetResource.Verbs, ", ")))

		if len(targetResource.ShortNames) > 0 {
			result.WriteString(fmt.Sprintf("SHORTNAMES:   %s\n", strings.Join(targetResource.ShortNames, ", ")))
		}

		// 如果指定了字段，显示该字段的详细信息
		if field != "" {
			result.WriteString(fmt.Sprintf("\nFIELD:        %s\n", field))
		}

		// 使用OpenAPI模式解释字段结构
		result.WriteString("\nDESCRIPTION:\n")

		// 提供一些常见字段的说明
		if field == "" || field == "metadata" {
			result.WriteString("  metadata - 标准的Kubernetes对象元数据\n")
			if recursive {
				result.WriteString("    name        - 对象的名称，在命名空间内必须唯一\n")
				result.WriteString("    namespace   - 对象所属的命名空间\n")
				result.WriteString("    labels      - 键值对标签，用于组织和分类对象\n")
				result.WriteString("    annotations - 键值对注释，用于存储非识别性元数据\n")
			}
		}

		if field == "" || field == "spec" {
			result.WriteString("  spec - 期望状态的规格说明\n")
			// 根据不同资源类型提供更具体的spec字段说明
			if strings.EqualFold(kind, "Pod") {
				if recursive {
					result.WriteString("    containers  - Pod中的容器列表\n")
					result.WriteString("    volumes     - Pod可以挂载的卷定义\n")
					result.WriteString("    nodeSelector - 限制Pod调度到匹配标签的节点上\n")
				}
			} else if strings.EqualFold(kind, "Deployment") {
				if recursive {
					result.WriteString("    replicas    - 期望运行的Pod副本数\n")
					result.WriteString("    selector    - 标签选择器，用于标识Pod\n")
					result.WriteString("    template    - Pod模板，定义要创建的Pod\n")
					result.WriteString("    strategy    - 部署策略，控制Pod更新方式\n")
				}
			} else if strings.EqualFold(kind, "Service") {
				if recursive {
					result.WriteString("    selector    - 标签选择器，选择服务后端Pod\n")
					result.WriteString("    ports       - 服务暴露的端口列表\n")
					result.WriteString("    type        - 服务类型 (ClusterIP, NodePort, LoadBalancer, ExternalName)\n")
				}
			}
		}

		if field == "" || field == "status" {
			result.WriteString("  status - 当前状态信息\n")
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

// getScopeText 返回资源作用域的文本描述
func getScopeText(namespaced bool) string {
	if namespaced {
		return "Namespaced"
	}
	return "Cluster"
}

// ApplyManifest 应用资源清单
func (h *UtilityHandler) ApplyManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	yamlStr, _ := arguments["yaml"].(string)
	dryRun, _ := arguments["dryRun"].(bool)
	fieldManager, _ := arguments["fieldManager"].(string)

	h.Log.Info("Applying manifest",
		"dryRun", dryRun,
		"fieldManager", fieldManager,
	)

	if yamlStr == "" {
		return nil, fmt.Errorf("yaml manifest is required")
	}

	// 构建响应
	var result strings.Builder
	if dryRun {
		result.WriteString("Dry Run: Resources that would be applied:\n\n")
	} else {
		result.WriteString("Applied Resources:\n\n")
	}

	// 将YAML拆分为多个文档
	docs := strings.Split(yamlStr, "---")
	appliedCount := 0
	errorCount := 0

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 解析YAML为非结构化对象
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			h.Log.Error("Failed to parse YAML document",
				"document", i+1,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: %v\n", i+1, err))
			errorCount++
			continue
		}

		// 获取资源类型和名称
		kind := obj.GetKind()
		apiVersion := obj.GetAPIVersion()
		name := obj.GetName()
		namespace := obj.GetNamespace()

		if kind == "" || apiVersion == "" {
			h.Log.Error("Document is missing kind or apiVersion",
				"document", i+1,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: missing kind or apiVersion\n", i+1))
			errorCount++
			continue
		}

		if name == "" {
			h.Log.Error("Document is missing metadata.name",
				"document", i+1,
				"kind", kind,
				"apiVersion", apiVersion,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: missing metadata.name\n", i+1))
			errorCount++
			continue
		}

		h.Log.Info("Processing resource",
			"document", i+1,
			"kind", kind,
			"apiVersion", apiVersion,
			"name", name,
			"namespace", namespace,
		)

		// 设置ServerSideApply选项
		var options metav1.PatchOptions
		if fieldManager != "" {
			options.FieldManager = fieldManager
		} else {
			options.FieldManager = "kubernetes-mcp"
		}

		if dryRun {
			options.DryRun = []string{"All"}
		}

		// 确定资源的组、版本和资源类型
		group, version := parseGroup(apiVersion), parseVersion(apiVersion)
		gvr, err := h.Client.GetDiscoveryClient().ServerResourcesForGroupVersion(apiVersion)
		if err != nil {
			h.Log.Error("Failed to get resource for group version",
				"apiVersion", apiVersion,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error: Failed to get resource for apiVersion %s: %v\n", apiVersion, err))
			errorCount++
			continue
		}

		// 查找资源名称
		var resourceName string
		for _, r := range gvr.APIResources {
			if strings.EqualFold(r.Kind, kind) {
				resourceName = r.Name
				break
			}
		}

		if resourceName == "" {
			h.Log.Error("Resource not found",
				"kind", kind,
				"apiVersion", apiVersion,
			)
			result.WriteString(fmt.Sprintf("Error: Resource not found for kind %s with apiVersion %s\n", kind, apiVersion))
			errorCount++
			continue
		}

		// 使用动态客户端应用资源
		dynamicClient := h.Client.GetDynamicClient()
		var dr dynamic.ResourceInterface

		// 确定是命名空间资源还是集群资源
		isNamespaced := false
		for _, r := range gvr.APIResources {
			if strings.EqualFold(r.Kind, kind) && r.Namespaced {
				isNamespaced = true
				break
			}
		}

		// 获取适当的动态资源接口
		if isNamespaced {
			ns := namespace
			if ns == "" {
				ns = "default"
			}
			dr = dynamicClient.Resource(schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resourceName,
			}).Namespace(ns)
		} else {
			dr = dynamicClient.Resource(schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resourceName,
			})
		}

		// 转换为JSON以应用
		data, err := json.Marshal(obj)
		if err != nil {
			h.Log.Error("Failed to marshal object to JSON",
				"kind", kind,
				"name", name,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error: Failed to marshal %s/%s: %v\n", kind, name, err))
			errorCount++
			continue
		}

		// 使用服务器端应用
		_, err = dr.Patch(ctx, name, types.ApplyPatchType, data, options)
		if err != nil {
			h.Log.Error("Failed to apply resource",
				"kind", kind,
				"name", name,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error: Failed to apply %s/%s: %v\n", kind, name, err))
			errorCount++
			continue
		}

		// 记录成功
		if namespace != "" {
			result.WriteString(fmt.Sprintf("Success: Applied %s/%s in namespace %s\n", kind, name, namespace))
		} else {
			result.WriteString(fmt.Sprintf("Success: Applied %s/%s (cluster-scoped)\n", kind, name))
		}
		appliedCount++
	}

	// 添加摘要
	result.WriteString(fmt.Sprintf("\nSummary: %d resource(s) applied, %d error(s)\n", appliedCount, errorCount))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// ValidateManifest 验证资源清单
func (h *UtilityHandler) ValidateManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Validating manifest")

	if yamlStr == "" {
		return nil, fmt.Errorf("yaml manifest is required")
	}

	// 构建响应
	var result strings.Builder
	result.WriteString("Validation Results:\n\n")

	// 将YAML拆分为多个文档
	docs := strings.Split(yamlStr, "---")
	validCount := 0
	errorCount := 0

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 解析YAML为非结构化对象
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			h.Log.Error("Failed to parse YAML document",
				"document", i+1,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: YAML parsing failed - %v\n", i+1, err))
			errorCount++
			continue
		}

		// 获取资源类型和名称
		kind := obj.GetKind()
		apiVersion := obj.GetAPIVersion()
		name := obj.GetName()
		namespace := obj.GetNamespace()

		// 验证基本字段
		if kind == "" || apiVersion == "" {
			h.Log.Error("Document is missing kind or apiVersion",
				"document", i+1,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: missing kind or apiVersion\n", i+1))
			errorCount++
			continue
		}

		if name == "" {
			h.Log.Error("Document is missing metadata.name",
				"document", i+1,
				"kind", kind,
				"apiVersion", apiVersion,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: missing metadata.name\n", i+1))
			errorCount++
			continue
		}

		// 检查API资源是否存在
		gvr, err := h.Client.GetDiscoveryClient().ServerResourcesForGroupVersion(apiVersion)
		if err != nil {
			h.Log.Error("Failed to get resource for group version",
				"apiVersion", apiVersion,
				"error", err,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: apiVersion '%s' not found in the cluster\n", i+1, apiVersion))
			errorCount++
			continue
		}

		// 查找资源类型
		resourceFound := false
		for _, r := range gvr.APIResources {
			if strings.EqualFold(r.Kind, kind) {
				resourceFound = true
				break
			}
		}

		if !resourceFound {
			h.Log.Error("Resource not found",
				"kind", kind,
				"apiVersion", apiVersion,
			)
			result.WriteString(fmt.Sprintf("Error in document %d: kind '%s' with apiVersion '%s' not found in the cluster\n", i+1, kind, apiVersion))
			errorCount++
			continue
		}

		// 验证通过，记录
		if namespace != "" {
			result.WriteString(fmt.Sprintf("Valid: %s/%s in namespace %s (document %d)\n", kind, name, namespace, i+1))
		} else {
			result.WriteString(fmt.Sprintf("Valid: %s/%s (cluster-scoped) (document %d)\n", kind, name, i+1))
		}
		validCount++
	}

	// 添加摘要
	result.WriteString(fmt.Sprintf("\nSummary: %d valid, %d invalid out of %d documents\n", validCount, errorCount, validCount+errorCount))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// DiffManifest 比较资源清单与集群中的资源
func (h *UtilityHandler) DiffManifest(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Diffing manifest")

	if yamlStr == "" {
		return nil, fmt.Errorf("yaml manifest is required")
	}

	// 构建响应
	var result strings.Builder
	result.WriteString("Diff Results:\n\n")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlStr), &obj.Object); err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 获取资源信息
	kind := obj.GetKind()
	apiVersion := obj.GetAPIVersion()
	name := obj.GetName()
	namespace := obj.GetNamespace()

	if kind == "" || apiVersion == "" || name == "" {
		return nil, fmt.Errorf("YAML must include kind, apiVersion, and metadata.name")
	}

	// 获取集群中的现有资源
	liveObj := &unstructured.Unstructured{}
	liveObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   parseGroup(apiVersion),
		Version: parseVersion(apiVersion),
		Kind:    kind,
	})

	// 确定资源的组、版本和资源类型
	group, version := parseGroup(apiVersion), parseVersion(apiVersion)
	gvr, err := h.Client.GetDiscoveryClient().ServerResourcesForGroupVersion(apiVersion)
	if err != nil {
		h.Log.Error("Failed to get resource for group version",
			"apiVersion", apiVersion,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get resource definition: %w", err)
	}

	// 查找资源名称
	var resourceName string
	var namespaced bool
	for _, r := range gvr.APIResources {
		if strings.EqualFold(r.Kind, kind) {
			resourceName = r.Name
			namespaced = r.Namespaced
			break
		}
	}

	if resourceName == "" {
		return nil, fmt.Errorf("resource kind %s with apiVersion %s not found in the cluster", kind, apiVersion)
	}

	// 使用动态客户端获取现有资源
	var dynamicResource dynamic.ResourceInterface
	if namespaced {
		ns := namespace
		if ns == "" {
			ns = "default" // 使用默认命名空间
		}
		dynamicResource = h.Client.GetDynamicClient().Resource(schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resourceName,
		}).Namespace(ns)
	} else {
		dynamicResource = h.Client.GetDynamicClient().Resource(schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resourceName,
		})
	}

	// 获取现有资源
	existingObj, err := dynamicResource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		h.Log.Error("Failed to get existing resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		result.WriteString(fmt.Sprintf("Resource %s/%s does not exist in the cluster. This would be a new resource.\n", kind, name))
		// 显示将要创建的资源概要
		result.WriteString("\nNew resource to be created:\n")
		result.WriteString(fmt.Sprintf("Kind:       %s\n", kind))
		result.WriteString(fmt.Sprintf("API Version: %s\n", apiVersion))
		result.WriteString(fmt.Sprintf("Name:       %s\n", name))
		if namespace != "" {
			result.WriteString(fmt.Sprintf("Namespace:  %s\n", namespace))
		} else {
			result.WriteString("Namespace:  <cluster-scoped>\n")
		}

		// 显示标签和注释
		labels := obj.GetLabels()
		if len(labels) > 0 {
			result.WriteString("\nLabels:\n")
			for k, v := range labels {
				result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}

		annotations := obj.GetAnnotations()
		if len(annotations) > 0 {
			result.WriteString("\nAnnotations:\n")
			for k, v := range annotations {
				result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
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

	// 存在的资源，比较差异
	result.WriteString(fmt.Sprintf("Comparing %s/%s in %s:\n\n", kind, name, namespace))

	// 移除比较时不需要的字段（如状态，资源版本等）
	cleanObject(obj)
	cleanObject(existingObj)

	// 比较字段差异
	result.WriteString("Field differences:\n")
	diffCount := 0

	// 转成JSON便于比较
	newJSON, _ := json.MarshalIndent(obj.Object, "", "  ")
	existingJSON, _ := json.MarshalIndent(existingObj.Object, "", "  ")

	if string(newJSON) == string(existingJSON) {
		result.WriteString("  No differences found. Resources are identical.\n")
	} else {
		// 比较特定的关键字段
		fieldsToCompare := map[string]string{
			"apiVersion": "API Version",
			"kind":       "Kind",
		}

		// 添加可能存在的规格字段
		spec, found, _ := unstructured.NestedMap(obj.Object, "spec")
		if found {
			for k := range spec {
				fieldsToCompare[fmt.Sprintf("spec.%s", k)] = fmt.Sprintf("Spec.%s", k)
			}
		}

		// 添加可能存在的元数据字段
		metadata, found, _ := unstructured.NestedMap(obj.Object, "metadata")
		if found {
			// 过滤一些不需要比较的元数据字段
			metadataFieldsToSkip := map[string]bool{
				"resourceVersion":   true,
				"uid":               true,
				"selfLink":          true,
				"generation":        true,
				"creationTimestamp": true,
				"managedFields":     true,
			}

			for k := range metadata {
				if !metadataFieldsToSkip[k] {
					fieldsToCompare[fmt.Sprintf("metadata.%s", k)] = fmt.Sprintf("Metadata.%s", k)
				}
			}
		}

		// 比较字段
		for path, displayName := range fieldsToCompare {
			parts := strings.Split(path, ".")
			var newValue, existingValue interface{}
			var newFound, existingFound bool

			// 获取路径对应的值
			if len(parts) == 1 {
				newValue, newFound = obj.Object[parts[0]]
				existingValue, existingFound = existingObj.Object[parts[0]]
			} else if len(parts) == 2 {
				newMap, found, _ := unstructured.NestedMap(obj.Object, parts[0])
				if found {
					newValue, newFound = newMap[parts[1]]
				}

				existingMap, found, _ := unstructured.NestedMap(existingObj.Object, parts[0])
				if found {
					existingValue, existingFound = existingMap[parts[1]]
				}
			}

			// 比较值是否不同
			if !reflect.DeepEqual(newValue, existingValue) || newFound != existingFound {
				diffCount++
				if !newFound && existingFound {
					result.WriteString(fmt.Sprintf("  - %s: would be removed (currently: %v)\n", displayName, existingValue))
				} else if newFound && !existingFound {
					result.WriteString(fmt.Sprintf("  + %s: would be added (%v)\n", displayName, newValue))
				} else {
					result.WriteString(fmt.Sprintf("  ~ %s: would change from %v to %v\n", displayName, existingValue, newValue))
				}
			}
		}

		if diffCount == 0 {
			// 如果没有检测到具体字段差异，但JSON不同，则提供一般性差异信息
			result.WriteString("  Differences detected, but may be in fields not specifically compared.\n")
			result.WriteString("  Consider using kubectl diff or a similar tool for a detailed comparison.\n")
		}
	}

	// 总结
	if diffCount > 0 {
		result.WriteString(fmt.Sprintf("\nSummary: Found %d differences between manifest and live resource.\n", diffCount))
	} else {
		result.WriteString("\nSummary: No significant differences found.\n")
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

// cleanObject 清理对象，移除不相关的比较字段
func cleanObject(obj *unstructured.Unstructured) {
	// 删除status
	unstructured.RemoveNestedField(obj.Object, "status")

	// 删除元数据中的自动生成字段
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "generation")
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
}

// GetEvents 获取资源的事件
func (h *UtilityHandler) GetEvents(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间
	namespace := namespaceArg
	if namespace == "" {
		namespace = "default"
	}

	h.Log.Info("Getting resource events",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	if kind == "" || apiVersion == "" || name == "" {
		return nil, fmt.Errorf("missing required parameters: kind, apiVersion, and name")
	}

	// 构建完整的资源名称
	resourceName := fmt.Sprintf("%s/%s", strings.ToLower(kind), name)

	// 创建响应构建器
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Events for %s/%s in namespace %s:\n\n", kind, name, namespace))

	// 获取所有事件
	eventsList := &corev1.EventList{}
	err := h.Client.List(ctx, eventsList, &ctrlclient.ListOptions{
		Namespace: namespace,
	})

	if err != nil {
		h.Log.Error("Failed to list events", "error", err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// 过滤与指定资源相关的事件
	var relatedEvents []corev1.Event
	for _, event := range eventsList.Items {
		if event.InvolvedObject.Kind == kind && event.InvolvedObject.Name == name {
			relatedEvents = append(relatedEvents, event)
		}
	}

	// 按照时间排序
	sort.Slice(relatedEvents, func(i, j int) bool {
		return relatedEvents[i].LastTimestamp.After(relatedEvents[j].LastTimestamp.Time)
	})

	// 如果没有找到事件
	if len(relatedEvents) == 0 {
		result.WriteString(fmt.Sprintf("No events found for %s '%s' in namespace '%s'\n", kind, name, namespace))
		result.WriteString("\nPossible reasons:\n")
		result.WriteString(" - The resource is new and hasn't generated any events yet\n")
		result.WriteString(" - The resource is operating normally without issues\n")
		result.WriteString(" - The resource does not exist in the specified namespace\n")
		result.WriteString(" - Events older than the retention period have been cleaned up\n")
	} else {
		// 写入标题
		result.WriteString(fmt.Sprintf("Found %d events:\n\n", len(relatedEvents)))
		result.WriteString(fmt.Sprintf("%-25s %-10s %-15s %-20s %s\n", "LAST SEEN", "TYPE", "REASON", "OBJECT", "MESSAGE"))
		result.WriteString(strings.Repeat("-", 100) + "\n")

		// 写入事件
		for _, event := range relatedEvents {
			// 格式化时间
			lastSeen := formatTimeAgo(event.LastTimestamp.Time)

			// 截断过长的消息
			message := event.Message
			if len(message) > 1024 {
				message = message[:1021] + "..."
			}

			// 写入事件信息
			result.WriteString(fmt.Sprintf("%-25s %-10s %-15s %-20s %s\n",
				lastSeen,
				event.Type,
				event.Reason,
				resourceName,
				message,
			))
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
