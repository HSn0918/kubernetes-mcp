package client

import (
	"context"
	"fmt"
	"os"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubernetesClient 定义了与Kubernetes API交互的接口
type KubernetesClient interface {
	client.Client
}

// k8sClientImpl 基于controller-runtime/client的Kubernetes客户端实现
type k8sClientImpl struct {
	client client.Client
}

// 确保k8sClientImpl实现了KubernetesClient接口
var _ KubernetesClient = &k8sClientImpl{}

// 全局客户端实例
var defaultClient KubernetesClient

// Create 实现接口方法
func (k *k8sClientImpl) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return k.client.Create(ctx, obj, opts...)
}

// Delete 实现接口方法
func (k *k8sClientImpl) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return k.client.Delete(ctx, obj, opts...)
}

// Update 实现接口方法
func (k *k8sClientImpl) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return k.client.Update(ctx, obj, opts...)
}

// Get 实现接口方法 - with options
func (k *k8sClientImpl) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return k.client.Get(ctx, key, obj, opts...)
}

// List 实现接口方法
func (k *k8sClientImpl) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return k.client.List(ctx, list, opts...)
}

// Patch 实现接口方法
func (k *k8sClientImpl) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return k.client.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf 实现接口方法
func (k *k8sClientImpl) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return k.client.DeleteAllOf(ctx, obj, opts...)
}

// Status 实现接口方法
func (k *k8sClientImpl) Status() client.StatusWriter {
	return k.client.Status()
}

// Scheme 实现接口方法
func (k *k8sClientImpl) Scheme() *runtime.Scheme {
	return k.client.Scheme()
}

// RESTMapper 实现接口方法
func (k *k8sClientImpl) RESTMapper() meta.RESTMapper {
	return k.client.RESTMapper()
}

// SubResource 实现接口方法
func (k *k8sClientImpl) SubResource(subResource string) client.SubResourceClient {
	return k.client.SubResource(subResource)
}

// GroupVersionKindFor 实现接口方法
func (k *k8sClientImpl) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return k.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced 实现接口方法
func (k *k8sClientImpl) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return k.client.IsObjectNamespaced(obj)
}

// NewClient 创建新的Kubernetes客户端
func NewClient(cfg *config.Config) (KubernetesClient, error) {
	log := logger.GetLogger()

	// 创建新的scheme
	scheme := runtime.NewScheme()

	// 注册标准Kubernetes类型到scheme
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %v", err)
	}

	// 在这里添加自定义资源定义（如果需要）
	// 例如:
	// if err := myapiv1.AddToScheme(scheme); err != nil {
	//     return nil, fmt.Errorf("failed to add custom API scheme: %v", err)
	// }

	var restConfig *rest.Config
	var err error

	// 根据配置选择加载方式
	if cfg.Kubeconfig != "" {
		log.Debug("Using specified kubeconfig", "path", cfg.Kubeconfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		// 尝试从环境变量加载
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("could not get user home directory: %v", err)
			}
			kubeconfig = homeDir + "/.kube/config"
		}

		log.Debug("Using kubeconfig", "path", kubeconfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

		// 如果失败，尝试集群内配置
		if err != nil {
			log.Debug("Failed to load kubeconfig, trying in-cluster config")
			restConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, fmt.Errorf("could not configure Kubernetes client: %v", err)
			}
			log.Debug("Using in-cluster config")
		}
	}

	// 使用自定义scheme创建客户端
	c, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes client: %v", err)
	}

	log.Info("Kubernetes client initialized successfully")
	return &k8sClientImpl{client: c}, nil
}

// InitializeDefaultClient 初始化默认客户端
func InitializeDefaultClient(cfg *config.Config) error {
	var err error
	defaultClient, err = NewClient(cfg)
	return err
}

// GetClient 获取默认客户端实例
func GetClient() KubernetesClient {
	return defaultClient
}
