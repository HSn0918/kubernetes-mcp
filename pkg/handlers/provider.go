package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/mark3labs/mcp-go/server"
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

	// 按照API组和作用域组织处理程序
	handlers := []interfaces.ToolHandler{
		// 集群级别资源
		factory.CreateNamespaceHandler(), // 集群作用域, 核心API组

		// 命名空间级别资源
		factory.CreateCoreHandler(),       // 核心API组
		factory.CreateAppsHandler(),       // 应用API组
		factory.CreateBatchHandler(),      // 批处理API组
		factory.CreateNetworkingHandler(), // 网络API组

		// 可以根据需要添加更多API组
	}

	return &HandlerProviderImpl{
		handlers: handlers,
	}
}
