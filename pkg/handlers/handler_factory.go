package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	appsv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/apps/v1"
	corev1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/core/v1"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	metricshandler "github.com/hsn0918/kubernetes-mcp/pkg/handlers/metrics"
	prompthandler "github.com/hsn0918/kubernetes-mcp/pkg/handlers/prompt"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/tool"
)

type HandlerFactoryImpl struct {
	client kubernetes.Client
}

var _ interfaces.HandlerFactory = &HandlerFactoryImpl{}

func NewHandlerFactory(client kubernetes.Client) interfaces.HandlerFactory {
	return &HandlerFactoryImpl{client: client}
}

func (f *HandlerFactoryImpl) CreateResourceHandler(scope interfaces.ResourceScope, group interfaces.APIGroup, prefix string) interfaces.ResourceHandler {
	h := base.NewHandler(f.client, scope, group)
	return base.NewResourceHandlerPtr(h, prefix)
}

func (f *HandlerFactoryImpl) CreateCoreHandler() interfaces.ResourceHandler {
	return corev1.NewResourceHandler(f.client)
}

func (f *HandlerFactoryImpl) CreateAppsHandler() interfaces.ResourceHandler {
	return appsv1.NewResourceHandler(f.client)
}

func (f *HandlerFactoryImpl) CreateNamespaceHandler() interfaces.NamespaceHandler {
	return corev1.NewNamespaceHandler(f.client)
}

func (f *HandlerFactoryImpl) CreateNodeHandler() interfaces.ToolHandler {
	return corev1.NewNodeHandler(f.client)
}

func (f *HandlerFactoryImpl) CreateUtilityHandler() interfaces.ToolHandler {
	return tool.NewUtilityHandler(f.client)
}

func (f *HandlerFactoryImpl) CreatePromptHandler() interfaces.ToolHandler {
	return prompthandler.NewPromptHandler(f.client)
}

func (f *HandlerFactoryImpl) CreateMetricsHandler() interfaces.ToolHandler {
	return metricshandler.NewMetricsHandler(f.client)
}
