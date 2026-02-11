package v1

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/hsn0918/kubernetes-mcp/pkg/utils"
)

const defaultAppsAPIVersion = "apps/v1"

type ResourceHandlerImpl struct {
	handler     base.Handler
	baseHandler interfaces.BaseResourceHandler
	listMethod  string
}

var _ interfaces.ResourceHandler = &ResourceHandlerImpl{}

func NewResourceHandler(client kubernetes.Client) interfaces.ResourceHandler {
	h := base.NewHandler(client, interfaces.NamespaceScope, interfaces.AppsAPIGroup)
	baseResourceHandler := base.NewResourceHandlerPtr(h, "APPS")
	return &ResourceHandlerImpl{
		handler:     h,
		baseHandler: baseResourceHandler,
		listMethod:  fmt.Sprintf("LIST_%s_RESOURCES", baseResourceHandler.GetResourcePrefix()),
	}
}

func (h *ResourceHandlerImpl) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if request.Method == h.listMethod {
		return h.ListResources(ctx, request)
	}
	return h.baseHandler.Handle(ctx, request)
}

func (h *ResourceHandlerImpl) Register(server *server.MCPServer) {
	h.baseHandler.Register(server)
}

func (h *ResourceHandlerImpl) GetScope() interfaces.ResourceScope {
	return h.handler.GetScope()
}

func (h *ResourceHandlerImpl) GetAPIGroup() interfaces.APIGroup {
	return h.handler.GetAPIGroup()
}

func (h *ResourceHandlerImpl) ListResources(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	kind, apiVersion, namespace, labelSelector, showLabels, errResult := h.parseListArguments(request)
	if errResult != nil {
		return errResult, nil
	}

	h.handler.Log.Info("Listing Apps resources",
		logger.String("kind", kind),
		logger.String("apiVersion", apiVersion),
		logger.String("namespace", namespace),
		logger.String("labelSelector", labelSelector),
	)

	gvk := utils.ParseGVK(apiVersion, kind)
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    kind + "List",
	})

	listOptions := &clientpkg.ListOptions{Namespace: namespace}
	if labelSelector != "" {
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			return utils.NewErrorToolResult(fmt.Sprintf("failed to parse label selector: %v", err)), nil
		}
		listOptions.LabelSelector = selector
	}

	if err := h.handler.Client.List(ctx, list, listOptions); err != nil {
		h.handler.Log.Error("Failed to list Apps resources",
			logger.String("kind", kind),
			logger.String("namespace", namespace),
			logger.Any("error", err),
		)
		return utils.NewErrorToolResult(fmt.Sprintf("failed to list Apps resources: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d %s resources in namespace %s", len(list.Items), kind, namespace))
	if labelSelector != "" {
		result.WriteString(fmt.Sprintf(" with label selector '%s'", labelSelector))
	}
	result.WriteString(":\n\n")

	for _, item := range list.Items {
		result.WriteString(fmt.Sprintf("- %s\n", item.GetName()))
		writeAppsKindDetails(&result, kind, &item)
		if showLabels {
			writeSortedLabels(&result, item.GetLabels())
		}
		result.WriteString("\n")
	}

	h.handler.Log.Info("Apps resources listed successfully",
		logger.String("kind", kind),
		logger.String("namespace", namespace),
		logger.Int("count", len(list.Items)),
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

func (h *ResourceHandlerImpl) GetResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.GetResource(ctx, request)
}

func (h *ResourceHandlerImpl) DescribeResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DescribeResource(ctx, request)
}

func (h *ResourceHandlerImpl) CreateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.CreateResource(ctx, request)
}

func (h *ResourceHandlerImpl) UpdateResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.UpdateResource(ctx, request)
}

func (h *ResourceHandlerImpl) DeleteResource(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	return h.baseHandler.DeleteResource(ctx, request)
}

func (h *ResourceHandlerImpl) parseListArguments(
	request mcp.CallToolRequest,
) (
	kind string,
	apiVersion string,
	namespace string,
	labelSelector string,
	showLabels bool,
	errResult *mcp.CallToolResult,
) {
	arguments := request.GetArguments()

	kind, _ = arguments["kind"].(string)
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "", "", "", "", false, utils.NewErrorToolResult("kind is required")
	}

	apiVersion, _ = arguments["apiVersion"].(string)
	apiVersion = strings.TrimSpace(apiVersion)
	if apiVersion == "" {
		apiVersion = defaultAppsAPIVersion
	}

	namespaceArg, _ := arguments["namespace"].(string)
	namespace = h.baseHandler.GetNamespaceWithDefault(strings.TrimSpace(namespaceArg))

	labelSelector, _ = arguments["labelSelector"].(string)
	labelSelector = strings.TrimSpace(labelSelector)

	showLabels, _ = arguments["showLabels"].(bool)
	return kind, apiVersion, namespace, labelSelector, showLabels, nil
}

func writeAppsKindDetails(result *strings.Builder, kind string, item *unstructured.Unstructured) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "deployment":
		writeNestedInt64(result, "Replicas", item, "spec", "replicas")
		writeNestedInt64(result, "Available", item, "status", "availableReplicas")
		writeNestedInt64(result, "Ready", item, "status", "readyReplicas")
	case "statefulset":
		writeNestedInt64(result, "Replicas", item, "spec", "replicas")
		writeNestedInt64(result, "Ready", item, "status", "readyReplicas")
	case "daemonset":
		writeNestedInt64(result, "Ready", item, "status", "numberReady")
		writeNestedInt64(result, "Desired", item, "status", "desiredNumberScheduled")
	}
}

func writeNestedInt64(result *strings.Builder, label string, item *unstructured.Unstructured, fields ...string) {
	if value, exists, _ := unstructured.NestedInt64(item.Object, fields...); exists {
		result.WriteString(fmt.Sprintf("  %s: %d\n", label, value))
	}
}

func writeSortedLabels(result *strings.Builder, labelMap map[string]string) {
	if len(labelMap) == 0 {
		return
	}
	keys := make([]string, 0, len(labelMap))
	for key := range labelMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result.WriteString("  Labels:\n")
	for _, key := range keys {
		result.WriteString(fmt.Sprintf("    %s: %s\n", key, labelMap[key]))
	}
}
