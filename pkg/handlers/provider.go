package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

type HandlerProviderImpl struct {
	handlers []interfaces.ToolHandler
}

var _ interfaces.HandlerProvider = &HandlerProviderImpl{}

func (p *HandlerProviderImpl) GetHandlers() []interfaces.ToolHandler {
	return p.handlers
}

func (p *HandlerProviderImpl) RegisterAllHandlers(server *server.MCPServer) {
	log := logger.GetLogger()
	for _, handler := range p.handlers {
		handler.Register(server)
	}
	log.Info("All handlers registered")
}

func NewHandlerProvider() interfaces.HandlerProvider {
	k8sClient := kubernetes.GetClient()
	factory := NewHandlerFactory(k8sClient)
	return &HandlerProviderImpl{handlers: buildHandlers(factory)}
}

func buildHandlers(factory interfaces.HandlerFactory) []interfaces.ToolHandler {
	handlers := []interfaces.ToolHandler{
		factory.CreateNamespaceHandler(),
		factory.CreateNodeHandler(),
		factory.CreateCoreHandler(),
		// Register one generic resource toolset instead of one toolset per API group.
		factory.CreateResourceHandler(interfaces.NamespaceScope, interfaces.ResourceAPIGroup, "K8S"),
	}

	return append(
		handlers,
		factory.CreateUtilityHandler(),
		factory.CreatePromptHandler(),
		factory.CreateMetricsHandler(),
	)
}
