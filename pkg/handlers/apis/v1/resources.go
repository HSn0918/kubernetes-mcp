package v1

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1api "k8s.io/api/core/v1"
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
	handler     base.Handler
	baseHandler interfaces.BaseResourceHandler
}

// 确保实现了接口
var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

// NewResourceHandler 创建新的核心资源处理程序
func NewResourceHandler(client client.KubernetesClient) interfaces.ResourceHandler {
	baseHandler := base.NewBaseHandler(client, interfaces.NamespaceScope, interfaces.CoreAPIGroup)
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
	default:
		// 其他方法使用父类的处理方法
		return h.baseHandler.Handle(ctx, request)
	}
}

// Register 实现接口方法
func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	// 注册父类的工具
	h.baseHandler.Register(server)

	// 注册特定于Core的工具
	server.AddTool(mcp.NewTool(GET_POD_LOGS,
		mcp.WithDescription("获取Pod日志"),
	), h.GetPodLogs)
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

// GetPodLogs 获取Pod日志
func (h *ResourceHandlerImpl) GetPodLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	podName, _ := arguments["podName"].(string)
	namespace, _ := arguments["namespace"].(string)
	containerName, _ := arguments["containerName"].(string)
	tailLines, _ := arguments["tailLines"].(float64)
	previous, _ := arguments["previous"].(bool)

	h.handler.Log.Info("Getting pod logs",
		"pod", podName,
		"namespace", namespace,
		"container", containerName,
		"tailLines", tailLines,
		"previous", previous,
	)

	// 设置日志选项
	podLogOpts := &corev1api.PodLogOptions{
		Container: containerName,
		Previous:  previous,
	}
	if tailLines > 0 {
		lines := int64(tailLines)
		podLogOpts.TailLines = &lines
	}

	// 获取Pod日志
	req := h.handler.Client.ClientSet().CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		h.handler.Log.Error("Failed to get pod logs",
			"pod", podName,
			"namespace", namespace,
			"container", containerName,
			"error", err,
		)
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("pod or container not found (Pod: %s, Container: %s, Namespace: %s)",
				podName, containerName, namespace)
		}
		return nil, fmt.Errorf("failed to get pod logs: %v", err)
	}
	defer podLogs.Close()

	// 读取日志内容
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		h.handler.Log.Error("Failed to read pod logs",
			"pod", podName,
			"namespace", namespace,
			"container", containerName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to read pod logs: %v", err)
	}

	// 处理日志内容
	var lines []string
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		h.handler.Log.Error("Failed to scan pod logs",
			"pod", podName,
			"namespace", namespace,
			"container", containerName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to scan pod logs: %v", err)
	}

	// 构建响应
	var resultText string
	if len(lines) == 0 {
		resultText = fmt.Sprintf("No logs available for pod %s in namespace %s", podName, namespace)
	} else {
		containerStr := ""
		if containerName != "" {
			containerStr = fmt.Sprintf(", container %s", containerName)
		}
		resultText = fmt.Sprintf("Logs for pod %s%s in namespace %s:\n\n%s",
			podName,
			containerStr,
			namespace,
			strings.Join(lines, "\n"),
		)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}
