package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
)

const (
	GET_POD_LOGS = "kubernetes.getPodLogs"
)

// ResourceHandlerImpl 核心资源处理程序实现
type ResourceHandlerImpl struct {
	base.ResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的核心资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.CoreAPIGroup)
	return &ResourceHandlerImpl{
		ResourceHandler: base.NewResourceHandler(baseHandler, "CORE"),
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case GET_POD_LOGS:
		return h.GetPodLogs(ctx, request)
	default:
		// 其他方法使用父类的处理方法
		return h.ResourceHandler.Handle(ctx, request)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 先注册基础资源处理工具
	h.ResourceHandler.Register(server)

	// 额外注册Pod日志工具
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

const (
	// 如果用户未指定 tailLines，并且日志行数超过此值，则默认显示最后这么多行
	defaultDisplayTailLines = 500
	MAX_LOG_BYTES_LIMIT     = 1024 * 1024 * 2
)

// GetPodLogs 获取Pod日志
func (h *ResourceHandlerImpl) GetPodLogs(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// --- 参数提取 ---
	arguments := request.Params.Arguments
	name, _ := arguments["name"].(string)
	namespaceArg, _ := arguments["namespace"].(string)

	// 获取命名空间，使用合适的默认值
	namespace := h.ResourceHandler.GetNamespaceWithDefault(namespaceArg)

	container, _ := arguments["container"].(string)
	tailLinesVal, _ := arguments["tailLines"]
	// 更安全地处理可能的类型，例如 float64 (JSON 数字) -> int
	var tailLines int
	if tlf, ok := tailLinesVal.(float64); ok {
		tailLines = int(tlf)
	} else if tli, ok := tailLinesVal.(int); ok {
		tailLines = tli
	}
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
	if tailLines > 0 {
		tailLinesInt64 := int64(tailLines)
		podLogOptions.TailLines = &tailLinesInt64
	}

	// --- 获取和读取日志流 ---
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

	// 读取日志内容
	buf := new(bytes.Buffer)
	_, err = io.CopyN(buf, podLogsStream, MAX_LOG_BYTES_LIMIT)
	if err != nil && err != io.EOF {
		reqLogger.Error("Failed to read pod logs stream fully", "error", err)
	}

	logsContent := buf.String()
	logLengthBytes := len(logsContent)
	logLines := strings.Split(logsContent, "\n")
	if len(logLines) > 0 && logLines[len(logLines)-1] == "" {
		logLines = logLines[:len(logLines)-1]
	}
	actualLineCount := len(logLines)

	// --- 处理日志截断显示 ---
	displayLogs := logsContent
	truncated := false
	displayLineCount := actualLineCount
	if tailLines <= 0 && actualLineCount > defaultDisplayTailLines {
		startIndex := actualLineCount - defaultDisplayTailLines
		displayLogs = strings.Join(logLines[startIndex:], "\n")
		truncated = true
		displayLineCount = defaultDisplayTailLines
	} else if tailLines > 0 && actualLineCount > tailLines {
		startIndex := actualLineCount - tailLines
		if startIndex < 0 {
			startIndex = 0
		}
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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: formattedOutput,
			},
		},
	}, nil
}

// formatPodLogsImproved 改进的日志格式化函数
func formatPodLogsImproved(summaryDetails string, logs string) string {
	var sb strings.Builder
	separator := "----------------------------------------------------------------------"

	// 1. 打印头部和摘要信息
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(" KUBERNETES POD LOGS\n")
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(summaryDetails)
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n\n")

	// 2. 打印日志内容（带缩进）
	if strings.TrimSpace(logs) == "" {
		sb.WriteString("  --- No logs available with the specified options. ---\n")
	} else {
		scanner := bufio.NewScanner(strings.NewReader(logs))
		for scanner.Scan() {
			sb.WriteString("  ")
			sb.WriteString(scanner.Text())
			sb.WriteString("\n")
		}
	}

	// 3. 打印尾部
	sb.WriteString("\n")
	sb.WriteString(separator)
	sb.WriteString("\n--- End of Logs ---\n")
	sb.WriteString(separator)
	sb.WriteString("\n")

	return sb.String()
}
