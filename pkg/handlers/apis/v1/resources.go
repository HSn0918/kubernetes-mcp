package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/models"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

const (
	GET_POD_LOGS     = "GET_POD_LOGS"
	ANALYZE_POD_LOGS = "ANALYZE_POD_LOGS"
)

// ResourceHandlerImpl 核心资源处理程序实现
type ResourceHandlerImpl struct {
	handler     base.Handler
	baseHandler interfaces.BaseResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的核心资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewHandler(client, interfaces.NamespaceScope, interfaces.CoreAPIGroup)
	baseResourceHandler := base.NewResourceHandlerPtr(baseHandler, "CORE")
	return &ResourceHandlerImpl{
		handler:     baseHandler,
		baseHandler: baseResourceHandler,
	}
}

// Handle 实现接口方法
func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 根据工具名称分派到具体的处理方法
	switch request.Method {
	case GET_POD_LOGS:
		return h.GetPodLogs(ctx, request)
	case ANALYZE_POD_LOGS:
		return h.AnalyzePodLogs(ctx, request)
	default:
		// 其他方法使用父类的处理方法
		return h.baseHandler.Handle(ctx, request)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 注册父类的工具
	h.baseHandler.Register(server)

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

	// 注册Pod日志分析工具
	server.AddTool(mcp.NewTool(ANALYZE_POD_LOGS,
		mcp.WithDescription("Analyze logs from a Pod and provide insights"),
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
			mcp.Description("Number of lines to analyze from the end of the logs"),
			mcp.DefaultNumber(1000),
		),
		mcp.WithBoolean("previous",
			mcp.Description("Whether to analyze logs from previous terminated container instance"),
			mcp.DefaultBool(false),
		),
		mcp.WithString("errorPattern",
			mcp.Description("Custom regex pattern to identify errors (default looks for common error keywords)"),
		),
		mcp.WithString("prompt",
			mcp.Description("Additional prompt for customizing the analysis focus"),
		),
	), h.AnalyzePodLogs)
}

// GetScope 实现ToolHandler接口
func (h *ResourceHandlerImpl) GetScope() interfaces.ResourceScope {
	return h.handler.GetScope()
}

// GetAPIGroup 实现ToolHandler接口
func (h *ResourceHandlerImpl) GetAPIGroup() interfaces.APIGroup {
	return h.handler.GetAPIGroup()
}

// ListResources 实现ResourceHandler接口
func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.ListResources(ctx, request)
}

// GetResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.GetResource(ctx, request)
}

// DescribeResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) DescribeResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DescribeResource(ctx, request)
}

// CreateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.CreateResource(ctx, request)
}

// UpdateResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.UpdateResource(ctx, request)
}

// DeleteResource 实现ResourceHandler接口
func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DeleteResource(ctx, request)
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

	// Type assertion with proper error handling
	nameVal, ok := arguments["name"]
	if !ok || nameVal == nil {
		return utils.NewErrorToolResult("Pod name is required"), nil
	}
	name := nameVal.(string)

	namespaceArg, _ := arguments["namespace"].(string) // namespace is optional with default

	// 获取命名空间，使用合适的默认值
	namespace := h.baseHandler.GetNamespaceWithDefault(namespaceArg)

	container, _ := arguments["container"].(string) // container is optional
	tailLinesVal := arguments["tailLines"]          // tailLines is handled specially below
	previous, _ := arguments["previous"].(bool)
	timestamps, _ := arguments["timestamps"].(bool)

	reqLogger := h.handler.Log.With("pod", name, "namespace", namespace, "container", container)
	reqLogger.Info("Starting pod logs request", "options", map[string]interface{}{
		"tailLines":  tailLinesVal,
		"previous":   previous,
		"timestamps": timestamps,
	})

	// --- 设置日志选项 ---
	podLogOptions := &corev1.PodLogOptions{
		Container:  container,
		Previous:   previous,
		Timestamps: timestamps,
	}

	// 处理tailLines参数
	var tailLines int
	if tailLinesVal != nil {
		// 转换tailLines为int类型
		if tlf, ok := tailLinesVal.(float64); ok {
			tailLines = int(tlf)
		} else if tli, ok := tailLinesVal.(int); ok {
			tailLines = tli
		} else {
			tailLines = 0 // 如果无法转换，视为不限制
		}
	} else {
		tailLines = 0 // 不限制
	}

	if tailLines > 0 {
		tailLinesInt64 := int64(tailLines)
		podLogOptions.TailLines = &tailLinesInt64
	}

	// --- 获取和读取日志流 ---
	logRESTRequest := h.handler.Client.ClientSet().CoreV1().Pods(namespace).GetLogs(name, podLogOptions)
	podLogsStream, err := logRESTRequest.Stream(ctx)
	if err != nil {
		reqLogger.Error("Failed to get pod logs stream", "error", err)
		if errors.IsNotFound(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("Pod '%s' not found in namespace '%s'", name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to stream pod logs for pod %s: %v", name, err)), nil
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

	// 已在上面声明了tailLines变量，这里不需要重新声明
	if tailLinesVal != nil {
		// 转换tailLines为int类型
		if tlf, ok := tailLinesVal.(float64); ok {
			tailLines = int(tlf)
		} else if tli, ok := tailLinesVal.(int); ok {
			tailLines = tli
		} else {
			tailLines = 0 // 如果无法转换，视为不限制
		}
	} else {
		tailLines = 0 // 不限制
	}

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

	// --- 构建JSON响应 ---
	logResponse := models.PodLogsResponse{
		Pod:          name,
		Namespace:    namespace,
		Container:    container,
		Previous:     previous,
		Timestamps:   timestamps,
		TailLines:    tailLines,
		LineCount:    displayLineCount,
		TotalLines:   actualLineCount,
		Truncated:    truncated,
		LogSize:      uint64(logLengthBytes),
		LogSizeHuman: humanize.Bytes(uint64(logLengthBytes)),
		Logs:         displayLogs,
		RetrievedAt:  time.Now(),
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(logResponse, "", "  ")
	if err != nil {
		return utils.NewErrorToolResult(fmt.Sprintf("JSON序列化失败: %v", err)), nil
	}

	reqLogger.Info("Pod logs retrieved successfully",
		"bytes", humanize.Bytes(uint64(logLengthBytes)),
		"linesRetrieved", humanize.Comma(int64(actualLineCount)),
		"linesDisplayed", humanize.Comma(int64(displayLineCount)))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

// AnalyzePodLogs 分析Pod日志并提供洞察
func (h *ResourceHandlerImpl) AnalyzePodLogs(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// --- 参数提取 ---
	arguments := request.Params.Arguments

	// Type assertion with proper error handling
	nameVal, ok := arguments["name"]
	if !ok || nameVal == nil {
		return utils.NewErrorToolResult("Pod name is required"), nil
	}
	name := nameVal.(string)

	namespaceArg, _ := arguments["namespace"].(string) // namespace is optional with default

	// 获取命名空间，使用合适的默认值
	namespace := h.baseHandler.GetNamespaceWithDefault(namespaceArg)

	container, _ := arguments["container"].(string) // container is optional
	tailLinesVal := arguments["tailLines"]          // tailLines is handled specially below

	// 处理tailLines参数
	var tailLines int
	if tailLinesVal != nil {
		// 转换tailLines为int类型
		if tlf, ok := tailLinesVal.(float64); ok {
			tailLines = int(tlf)
		} else if tli, ok := tailLinesVal.(int); ok {
			tailLines = tli
		} else {
			tailLines = 1000 // 如果无法转换，使用默认值1000
		}
	} else {
		tailLines = 1000 // 默认分析1000行
	}

	previous, _ := arguments["previous"].(bool)
	customErrorPattern, _ := arguments["errorPattern"].(string)
	prompt, _ := arguments["prompt"].(string)

	reqLogger := h.handler.Log.With("pod", name, "namespace", namespace, "container", container)
	reqLogger.Info("Starting pod logs analysis", "options", map[string]interface{}{
		"tailLines":    tailLines,
		"previous":     previous,
		"errorPattern": customErrorPattern,
		"prompt":       prompt,
	})

	// --- 设置日志选项 ---
	podLogOptions := &corev1.PodLogOptions{
		Container:  container,
		Previous:   previous,
		Timestamps: true, // 分析需要时间戳
	}
	if tailLines > 0 {
		tailLinesInt64 := int64(tailLines)
		podLogOptions.TailLines = &tailLinesInt64
	}

	// --- 获取和读取日志流 ---
	logRESTRequest := h.handler.Client.ClientSet().CoreV1().Pods(namespace).GetLogs(name, podLogOptions)
	podLogsStream, err := logRESTRequest.Stream(ctx)
	if err != nil {
		reqLogger.Error("Failed to get pod logs stream for analysis", "error", err)
		if errors.IsNotFound(err) {
			return utils.NewErrorToolResult(fmt.Sprintf("Pod '%s' not found in namespace '%s'", name, namespace)), nil
		}
		return utils.NewErrorToolResult(fmt.Sprintf("failed to stream pod logs for analysis, pod %s: %v", name, err)), nil
	}
	defer podLogsStream.Close()

	// 读取日志内容
	buf := new(bytes.Buffer)
	_, err = io.CopyN(buf, podLogsStream, MAX_LOG_BYTES_LIMIT)
	if err != nil && err != io.EOF {
		reqLogger.Error("Failed to read pod logs stream fully for analysis", "error", err)
	}

	logsContent := buf.String()
	logLines := strings.Split(logsContent, "\n")
	if len(logLines) > 0 && logLines[len(logLines)-1] == "" {
		logLines = logLines[:len(logLines)-1]
	}
	actualLineCount := len(logLines)

	// --- 日志分析 ---
	analyzer := utils.NewLogAnalyzer()

	// 如果提供了自定义错误模式，则设置它
	if customErrorPattern != "" {
		analyzer = utils.NewLogAnalyzerWithPattern(customErrorPattern)
	}

	analysis := analyzer.AnalyzeLogsWithPrompt(logLines, prompt)

	// --- 使用转换函数创建JSON响应 ---
	analysisResponse := models.NewLogAnalysisResponseFromResult(
		name, namespace, container,
		analysis,
		actualLineCount,
		previous,
		customErrorPattern,
		prompt,
	)

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(analysisResponse, "", "  ")
	if err != nil {
		return utils.NewErrorToolResult(fmt.Sprintf("JSON序列化失败: %v", err)), nil
	}

	reqLogger.Info("Pod logs analysis completed", "linesAnalyzed", actualLineCount)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}
