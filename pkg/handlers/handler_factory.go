package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/apps"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/batch"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/core"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/networking"
)

// HandlerFactoryImpl 实现HandlerFactory接口
type HandlerFactoryImpl struct {
	client client.KubernetesClient
}

// 确保实现了接口
var _ interfaces.HandlerFactory = &HandlerFactoryImpl{}

// NewHandlerFactory 创建新的处理程序工厂
func NewHandlerFactory(client client.KubernetesClient) interfaces.HandlerFactory {
	return &HandlerFactoryImpl{
		client: client,
	}
}

// CreateCoreHandler 创建核心资源处理程序
func (f *HandlerFactoryImpl) CreateCoreHandler() interfaces.ResourceHandler {
	return core.NewResourceHandler(f.client)
}

// CreateAppsHandler 创建应用资源处理程序
func (f *HandlerFactoryImpl) CreateAppsHandler() interfaces.ResourceHandler {
	return apps.NewResourceHandler(f.client)
}

// CreateBatchHandler 创建批处理资源处理程序
func (f *HandlerFactoryImpl) CreateBatchHandler() interfaces.ResourceHandler {
	return batch.NewResourceHandler(f.client)
}

// CreateNetworkingHandler 创建网络资源处理程序
func (f *HandlerFactoryImpl) CreateNetworkingHandler() interfaces.ResourceHandler {
	return networking.NewResourceHandler(f.client)
}

// CreateNamespaceHandler 创建命名空间处理程序
func (f *HandlerFactoryImpl) CreateNamespaceHandler() interfaces.NamespaceHandler {
	return core.NewNamespaceHandler(f.client)
}
