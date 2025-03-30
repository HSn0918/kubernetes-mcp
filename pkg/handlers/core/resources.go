package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	GET_POD_LOGS = "kubernetes.getPodLogs"
)

// ResourceHandlerImpl 核心资源处理程序实现
type ResourceHandlerImpl struct {
	base.Handler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的核心资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	return &ResourceHandlerImpl{
		Handler: base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.CoreAPIGroup),
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case base.LIST_RESOURCES:
		return h.ListResources(ctx, request)
	case base.GET_RESOURCE:
		return h.GetResource(ctx, request)
	case base.CREATE_RESOURCE:
		return h.CreateResource(ctx, request)
	case base.UPDATE_RESOURCE:
		return h.UpdateResource(ctx, request)
	case base.DELETE_RESOURCE:
		return h.DeleteResource(ctx, request)
	case GET_POD_LOGS:
		return h.GetPodLogs(ctx, request)
	default:
		return nil, fmt.Errorf("unknown resource method: %s", request.Method)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	h.Log.Info("Registering core resource handlers",
		"scope", h.Scope,
		"apiGroup", h.Group,
	)

	// 注册列出资源工具
	server.AddTool(mcp.NewTool(base.LIST_RESOURCES,
		mcp.WithDescription("List Kubernetes resources (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.ListResources)

	// 注册获取资源工具
	server.AddTool(mcp.NewTool(base.GET_RESOURCE,
		mcp.WithDescription("Get a specific Kubernetes resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
		),
		mcp.WithString("name",
			mcp.Description("Name of the resource"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.GetResource)

	// 注册创建资源工具
	server.AddTool(mcp.NewTool(base.CREATE_RESOURCE,
		mcp.WithDescription("Create a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.CreateResource)

	// 注册更新资源工具
	server.AddTool(mcp.NewTool(base.UPDATE_RESOURCE,
		mcp.WithDescription("Update a Kubernetes resource from YAML"),
		mcp.WithString("yaml",
			mcp.Description("YAML manifest of the resource"),
			mcp.Required(),
		),
	), h.UpdateResource)

	// 注册删除资源工具
	server.AddTool(mcp.NewTool(base.DELETE_RESOURCE,
		mcp.WithDescription("Delete a Kubernetes resource (Namespace-scoped)"),
		mcp.WithString("kind",
			mcp.Description("Kind of resource (Pod, Service, ConfigMap, etc.)"),
			mcp.Required(),
		),
		mcp.WithString("apiVersion",
			mcp.Description("API Version (v1)"),
			mcp.DefaultString("v1"),
		),
		mcp.WithString("name",
			mcp.Description("Name of the resource"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
	), h.DeleteResource)

	// 注册获取Pod日志工具
	server.AddTool(mcp.NewTool(GET_POD_LOGS,
		mcp.WithDescription("Get logs from a Pod"),
		mcp.WithString("name",
			mcp.Description("Name of the Pod"),
			mcp.Required(),
		),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace"),
			mcp.DefaultString("default"),
		),
		mcp.WithString("container",
			mcp.Description("Container name (if Pod has multiple containers)"),
		),
		mcp.WithNumber("tailLines",
			mcp.Description("Number of lines to show from the end of the logs (default 500)"),
			mcp.DefaultNumber(500),
		),
		mcp.WithBoolean("previous",
			mcp.Description("Whether to get logs from previous terminated container instance"),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("timestamps",
			mcp.Description("Include timestamps on each line"),
			mcp.DefaultBool(true),
		),
	), h.GetPodLogs)
}

// ListResources 实现接口方法
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Listing resources",
		"kind", kind,
		"apiVersion", apiVersion,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建列表对象
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    kind + "List",
	})

	// 列出资源
	err := h.Client.List(ctx, list, &clientpkg.ListOptions{Namespace: namespace})
	if err != nil {
		h.Log.Error("Failed to list resources",
			"kind", kind,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list resources: %v", err)
	}

	// 构建响应
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d %s resources in namespace %s:\n\n", len(list.Items), kind, namespace))

	for _, item := range list.Items {
		result.WriteString(fmt.Sprintf("Name: %s\n", item.GetName()))
	}

	h.Log.Info("Resources listed successfully",
		"kind", kind,
		"namespace", namespace,
		"count", len(list.Items),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result.String(),
			},
		},
	}, nil
}

// GetResource 实现接口方法
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Getting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建对象
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	// 获取资源
	err := h.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		h.Log.Error("Failed to get resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("core resource not found (Kind: %s, Name: %s, Namespace: %s)", kind, name, namespace)
		}
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	// 转换为YAML
	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		h.Log.Error("Failed to marshal resource to YAML",
			"kind", kind,
			"name", name,
			"error", err,
		)
		return nil, fmt.Errorf("failed to marshal to YAML: %v", err)
	}

	h.Log.Info("Resource retrieved successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(yamlData),
			},
		},
	}, nil
}

// CreateResource 实现接口方法
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Creating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}
	if obj.GetNamespace() == "" {
		obj.SetNamespace("default")
		h.Log.Debug("Empty namespace, using default namespace")
	}
	h.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 创建资源
	err = h.Client.Create(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to create resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	h.Log.Info("Resource created successfully",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created %s/%s in namespace %s",
					obj.GetKind(), obj.GetName(), obj.GetNamespace()),
			},
		},
	}, nil
}

// UpdateResource 实现接口方法
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	yamlStr, _ := arguments["yaml"].(string)

	h.Log.Info("Updating resource from YAML")

	// 解析YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlStr), &obj.Object)
	if err != nil {
		h.Log.Error("Failed to parse YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	h.Log.Debug("Parsed resource from YAML",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	// 更新资源
	err = h.Client.Update(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to update resource",
			"kind", obj.GetKind(),
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to update resource: %v", err)
	}

	h.Log.Info("Resource updated successfully",
		"kind", obj.GetKind(),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated %s/%s in namespace %s",
					obj.GetKind(), obj.GetName(), obj.GetNamespace()),
			},
		},
	}, nil
}

// DeleteResource 实现接口方法
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	kind, _ := arguments["kind"].(string)
	apiVersion, _ := arguments["apiVersion"].(string)
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)

	h.Log.Info("Deleting resource",
		"kind", kind,
		"apiVersion", apiVersion,
		"name", name,
		"namespace", namespace,
	)

	// 解析GroupVersionKind
	gvk := utils.ParseGVK(apiVersion, kind)

	// 创建对象
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	// 删除资源
	err := h.Client.Delete(ctx, obj)
	if err != nil {
		h.Log.Error("Failed to delete resource",
			"kind", kind,
			"name", name,
			"namespace", namespace,
			"error", err,
		)
		return nil, fmt.Errorf("failed to delete resource: %v", err)
	}

	h.Log.Info("Resource deleted successfully",
		"kind", kind,
		"name", name,
		"namespace", namespace,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted %s/%s from namespace %s", kind, name, namespace),
			},
		},
	}, nil
}

const (
	// 如果用户未指定 tailLines，并且日志行数超过此值，则默认显示最后这么多行
	defaultDisplayTailLines = 500
	MAX_LOG_BYTES_LIMIT     = 1024 * 1024 * 2
)

// GetPodLogs 获取Pod日志 (优化日志输出和格式)
func (h *ResourceHandlerImpl) GetPodLogs(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// --- 参数提取 (保持不变) ---
	arguments := request.Params.Arguments
	name, _ := arguments["name"].(string)
	namespace, _ := arguments["namespace"].(string)
	container, _ := arguments["container"].(string)
	tailLinesVal, _ := arguments["tailLines"]
	// 更安全地处理可能的类型，例如 float64 (JSON 数字) -> int
	var tailLines int
	if tlf, ok := tailLinesVal.(float64); ok {
		tailLines = int(tlf)
	} else if tli, ok := tailLinesVal.(int); ok {
		tailLines = tli
	} // 可以添加更多类型处理或错误处理
	previous, _ := arguments["previous"].(bool)
	timestamps, _ := arguments["timestamps"].(bool)

	reqLogger := h.Log.With("pod", name, "namespace", namespace, "container", container)
	reqLogger.Info("Starting pod logs request", "options", map[string]interface{}{
		"tailLines":  tailLines,
		"previous":   previous,
		"timestamps": timestamps,
	})

	// --- 设置日志选项 ---
	podLogOptions := &corev1.PodLogOptions{
		Container:  container,
		Previous:   previous,
		Timestamps: timestamps,
	}
	// **注意**: 如果用户指定了 tailLines，我们请求 Kubernetes 返回这些行。
	// 如果用户没指定 (tailLines <= 0)，我们请求所有日志 (稍后在格式化时可能截断显示)。
	// 也可以在这里就设置一个默认的 Kubernetes 请求 TailLines (例如 1000)，以限制初始拉取的数据量。
	// 这里我们保持原样：用户指定多少就请求多少，不指定就请求全部。
	if tailLines > 0 {
		tailLinesInt64 := int64(tailLines)
		podLogOptions.TailLines = &tailLinesInt64
	}

	// --- 获取和读取日志流 (逻辑基本不变，错误处理稍作调整) ---
	logRESTRequest := h.Client.ClientSet().CoreV1().Pods(namespace).GetLogs(name, podLogOptions)
	podLogsStream, err := logRESTRequest.Stream(ctx)
	if err != nil {
		reqLogger.Error("Failed to get pod logs stream", "error", err)
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Pod '%s' not found in namespace '%s'", name, namespace)
		}
		return nil, fmt.Errorf("failed to stream pod logs for pod %s: %w", name, err)
	}
	defer podLogsStream.Close()

	// 读取日志内容 (仍然需要读入内存以进行后续处理和格式化)
	buf := new(bytes.Buffer)
	// 考虑限制读取的最大字节数，防止 OOM
	_, err = io.CopyN(buf, podLogsStream, MAX_LOG_BYTES_LIMIT)
	// _, err = io.Copy(buf, podLogsStream)
	if err != nil && err != io.EOF { // EOF 不是真正的错误，只是流结束了
		reqLogger.Error("Failed to read pod logs stream fully", "error", err)
		// 即使读取部分失败，可能仍希望显示已读取的内容，或者返回错误
		// return nil, fmt.Errorf("failed to read pod logs stream for pod %s: %w", name, err)
		// 这里我们继续处理已读取的内容
	}

	logsContent := buf.String()
	logLengthBytes := len(logsContent)
	logLines := strings.Split(logsContent, "\n")
	// 移除可能因 Split 产生的末尾空字符串
	if len(logLines) > 0 && logLines[len(logLines)-1] == "" {
		logLines = logLines[:len(logLines)-1]
	}
	actualLineCount := len(logLines)

	// --- 处理日志截断显示 ---
	displayLogs := logsContent
	truncated := false
	displayLineCount := actualLineCount
	if tailLines <= 0 && actualLineCount > defaultDisplayTailLines {
		// 用户未指定 tailLines，但日志过多，截断显示最后 N 行
		startIndex := actualLineCount - defaultDisplayTailLines
		displayLogs = strings.Join(logLines[startIndex:], "\n")
		truncated = true
		displayLineCount = defaultDisplayTailLines
	} else if tailLines > 0 && actualLineCount > tailLines {
		// K8s 返回的行数可能比请求的多一点点，或者因为 split 多了空行，这里确保最多显示请求的行数
		startIndex := actualLineCount - tailLines
		if startIndex < 0 {
			startIndex = 0
		} // 避免负数索引
		displayLogs = strings.Join(logLines[startIndex:], "\n")
		displayLineCount = tailLines
	}

	// --- 构建摘要信息 ---
	summaryDetails := fmt.Sprintf("Pod: %s | Namespace: %s", name, namespace)
	if container != "" {
		summaryDetails += fmt.Sprintf(" | Container: %s", container)
	}
	summaryDetails += fmt.Sprintf(" | Options: previous=%t, timestamps=%t", previous, timestamps)
	if tailLines > 0 {
		summaryDetails += fmt.Sprintf(", tailLines=%d", tailLines)
	} else {
		summaryDetails += ", tailLines=All"
	}
	summaryDetails += fmt.Sprintf(" | Displaying %d lines", displayLineCount)
	if truncated {
		summaryDetails += fmt.Sprintf(" (truncated from %d lines, showing last %d)", actualLineCount, defaultDisplayTailLines)
	}

	// --- 格式化最终输出 ---
	formattedOutput := formatPodLogsImproved(summaryDetails, displayLogs)

	reqLogger.Info("Pod logs retrieved successfully", "bytes", logLengthBytes, "linesRetrieved", actualLineCount, "linesDisplayed", displayLineCount)

	// --- 返回统一的响应 ---
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: formattedOutput, // 将所有内容合并到一个 TextContent
			},
		},
	}, nil
}

// formatPodLogsImproved 改进的日志格式化函数
func formatPodLogsImproved(summaryDetails string, logs string) string {
	var sb strings.Builder
	separator := "----------------------------------------------------------------------" // 分隔线

	// 1. 打印头部和摘要信息
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(" KUBERNETES POD LOGS\n") // 标题
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(summaryDetails) // 打印包含所有选项的摘要
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n\n") // 空两行

	// 2. 打印日志内容（带缩进）
	if strings.TrimSpace(logs) == "" {
		sb.WriteString("  --- No logs available with the specified options. ---\n") // 清晰的空日志提示并缩进
	} else {
		scanner := bufio.NewScanner(strings.NewReader(logs))
		for scanner.Scan() {
			sb.WriteString("  ") // 每行日志前加两个空格缩进
			sb.WriteString(scanner.Text())
			sb.WriteString("\n")
		}
		// 可以选择检查 scanner.Err()
	}

	// 3. 打印尾部
	sb.WriteString("\n") // 日志内容后的空行
	sb.WriteString(separator)
	sb.WriteString("\n--- End of Logs ---\n")
	sb.WriteString(separator)
	sb.WriteString("\n")

	return sb.String()
}
