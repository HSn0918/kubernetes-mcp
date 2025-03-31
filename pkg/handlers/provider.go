package handlers

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

// HandlerProviderImpl 实现HandlerProvider接口
type HandlerProviderImpl struct {
	handlers []interfaces.ToolHandler
}

// 确保实现了接口
var _ interfaces.HandlerProvider = &HandlerProviderImpl{}

// GetHandlers 实现接口方法
func (p *HandlerProviderImpl) GetHandlers() []interfaces.ToolHandler {
	return p.handlers
}

// RegisterAllHandlers 实现接口方法
func (p *HandlerProviderImpl) RegisterAllHandlers(server *server.MCPServer) {
	log := logger.GetLogger()

	// 注册所有处理程序
	for _, handler := range p.handlers {
		handler.Register(server)
	}

	log.Info("All handlers registered")
}

// NewHandlerProvider 创建新的处理程序提供者
func NewHandlerProvider() interfaces.HandlerProvider {
	k8sClient := client.GetClient()

	// 使用工厂创建所有处理程序
	factory := NewHandlerFactory(k8sClient)

	// 按照API组和Version组织处理程序
	handlers := []interfaces.ToolHandler{
		// 集群级别资源
		factory.CreateNamespaceHandler(), // 集群作用域, v1 (core)
		factory.CreateNodeHandler(),      // 集群作用域, v1 (core)

		// 核心API组 (v1)
		factory.CreateCoreHandler(),

		// apps API组 (apps/v1)
		factory.CreateAppsHandler(),

		// batch API组 (batch/v1)
		factory.CreateBatchHandler(),

		// networking API组 (networking.k8s.io/v1)
		factory.CreateNetworkingHandler(),

		// storage API组 (storage.k8s.io/v1)
		factory.CreateStorageHandler(),

		// rbac API组 (rbac.authorization.k8s.io/v1)
		factory.CreateRbacHandler(),

		// policy API组 (policy/v1beta1)
		factory.CreatePolicyHandler(),

		// apiextensions API组 (apiextensions.k8s.io/v1)
		factory.CreateApiExtensionsHandler(),

		// autoscaling API组 (autoscaling/v1)
		factory.CreateAutoscalingHandler(),

		// 通用工具处理程序
		factory.CreateUtilityHandler(),
	}

	return &HandlerProviderImpl{
		handlers: handlers,
	}
}
