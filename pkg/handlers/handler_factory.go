package handlers

import (
	"github.com/hsn0918/kubernetes-mcp/pkg/client"
	apiextensionsv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/apiextensions/v1"
	appsv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/apps/v1"
	autoscalingv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/autoscaling/v1"
	batchv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/batch/v1"
	networkingv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/networking/v1"
	policyv1beta1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/policy/v1beta1"
	rbacv1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/rbac/v1"
	storagev1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/storage/v1"
	corev1 "github.com/hsn0918/kubernetes-mcp/pkg/handlers/apis/v1"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/base"
	"github.com/hsn0918/kubernetes-mcp/pkg/handlers/interfaces"
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
	return corev1.NewResourceHandler(f.client)
}

// CreateAppsHandler 创建应用资源处理程序
func (f *HandlerFactoryImpl) CreateAppsHandler() interfaces.ResourceHandler {
	return appsv1.NewResourceHandler(f.client)
}

// CreateBatchHandler 创建批处理资源处理程序
func (f *HandlerFactoryImpl) CreateBatchHandler() interfaces.ResourceHandler {
	return batchv1.NewResourceHandler(f.client)
}

// CreateNetworkingHandler 创建网络资源处理程序
func (f *HandlerFactoryImpl) CreateNetworkingHandler() interfaces.ResourceHandler {
	return networkingv1.NewResourceHandler(f.client)
}

// CreateStorageHandler 创建存储资源处理程序
func (f *HandlerFactoryImpl) CreateStorageHandler() interfaces.ResourceHandler {
	return storagev1.NewResourceHandler(f.client)
}

// CreateRbacHandler 创建RBAC资源处理程序
func (f *HandlerFactoryImpl) CreateRbacHandler() interfaces.ResourceHandler {
	return rbacv1.NewResourceHandler(f.client)
}

// CreatePolicyHandler 创建策略资源处理程序
func (f *HandlerFactoryImpl) CreatePolicyHandler() interfaces.ResourceHandler {
	return policyv1beta1.NewResourceHandler(f.client)
}

// CreateApiExtensionsHandler 创建API扩展资源处理程序
func (f *HandlerFactoryImpl) CreateApiExtensionsHandler() interfaces.ResourceHandler {
	return apiextensionsv1.NewResourceHandler(f.client)
}

// CreateAutoscalingHandler 创建自动伸缩资源处理程序
func (f *HandlerFactoryImpl) CreateAutoscalingHandler() interfaces.ResourceHandler {
	return autoscalingv1.NewResourceHandler(f.client)
}

// CreateNamespaceHandler 创建命名空间处理程序
func (f *HandlerFactoryImpl) CreateNamespaceHandler() interfaces.NamespaceHandler {
	return corev1.NewNamespaceHandler(f.client)
}

// CreateNodeHandler 创建节点处理程序
func (f *HandlerFactoryImpl) CreateNodeHandler() interfaces.ToolHandler {
	return corev1.NewNodeHandler(f.client)
}

// CreateUtilityHandler 创建通用工具处理程序
func (f *HandlerFactoryImpl) CreateUtilityHandler() interfaces.ToolHandler {
	return base.NewUtilityHandler(f.client)
}

// CreatePromptHandler 创建提示词处理程序
func (f *HandlerFactoryImpl) CreatePromptHandler() interfaces.ToolHandler {
	return base.NewPromptHandler(f.client)
}
