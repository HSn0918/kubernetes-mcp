package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"github.com/mark3labs/mcp-go/server"
)

// Constants for tool names
const (
	LIST_RESOURCES  = "listResources"
	GET_RESOURCE    = "getResource"
	CREATE_RESOURCE = "createResource"
	UPDATE_RESOURCE = "updateResource"
	DELETE_RESOURCE = "deleteResource"
	LIST_NAMESPACES = "listNamespaces"
)

// handlerProviderImpl 实现HandlerProvider接口
type handlerProviderImpl struct {
	resourceHandler  ResourceHandler
	namespaceHandler NamespaceHandler
}

// 确保实现了接口
var _ HandlerProvider = &handlerProviderImpl{}

// GetResourceHandler 实现接口方法
func (p *handlerProviderImpl) GetResourceHandler() ResourceHandler {
	return p.resourceHandler
}

// GetNamespaceHandler 实现接口方法
func (p *handlerProviderImpl) GetNamespaceHandler() NamespaceHandler {
	return p.namespaceHandler
}

// RegisterAllHandlers 实现接口方法
func (p *handlerProviderImpl) RegisterAllHandlers(server *server.MCPServer) {
	log := logger.GetLogger()

	// 注册资源处理程序
	p.resourceHandler.Register(server)

	// 注册命名空间处理程序
	p.namespaceHandler.Register(server)

	log.Info("All handlers registered")
}

// NewHandlerProvider 创建新的处理程序提供者
func NewHandlerProvider() HandlerProvider {
	k8sClient := client.GetClient()

	return &handlerProviderImpl{
		resourceHandler:  NewResourceHandler(k8sClient),
		namespaceHandler: NewNamespaceHandler(k8sClient),
	}
}

// baseHandler 提供公共功能
type baseHandler struct {
	client client.KubernetesClient
	log    logger.Logger
}

// newBaseHandler 创建新的基础处理程序
func newBaseHandler(client client.KubernetesClient) baseHandler {
	return baseHandler{
		client: client,
		log:    logger.GetLogger(),
	}
}
