package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubernetesClient 定义了与Kubernetes API交互的接口 (简化后)
type KubernetesClient interface {
	client.Client // 嵌入 controller-runtime 的 Client 接口
	// ClientSet 提供访问标准 client-go Clientset 的方法，用于 controller-runtime 不直接支持的操作 (如 GetLogs)
	ClientSet() kubernetes.Interface
	// GetCurrentNamespace 获取当前kubeconfig中配置的命名空间
	GetCurrentNamespace() (string, error)
}

// k8sClientImpl 基于controller-runtime/client的Kubernetes客户端实现 (增加 clientset 字段)
type k8sClientImpl struct {
	client    client.Client          // controller-runtime 客户端
	clientset kubernetes.Interface   // 标准 client-go Clientset
	rawConfig clientcmd.ClientConfig // 保存原始kubeconfig，用于获取当前命名空间
}

// 确保k8sClientImpl实现了KubernetesClient接口
var _ KubernetesClient = &k8sClientImpl{}

// 全局客户端实例 (保持不变)
var defaultClient KubernetesClient

// ClientSet 实现接口方法 (优化后)
// 直接返回初始化时创建并存储的 clientset
func (k *k8sClientImpl) ClientSet() kubernetes.Interface {
	// 注意：这里假设 k.clientset 在 NewClient 中已经被成功初始化
	if k.clientset == nil {
		// 这种情况理论上不应该发生，如果在 NewClient 中正确初始化了
		// 但可以加一个 panic 或返回错误，取决于你希望如何处理未初始化的情况
		panic("kubernetes clientset accessed before initialization")
	}
	return k.clientset
}

// --- controller-runtime client.Client 接口的实现 (通过嵌入 k.client 自动完成) ---
// Create, Delete, Update, Get, List, Patch, DeleteAllOf, Status, Scheme, RESTMapper,
// SubResource, GroupVersionKindFor, IsObjectNamespaced 这些方法都由嵌入的 k.client 提供

// Create (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return k.client.Create(ctx, obj, opts...)
}

// Delete (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return k.client.Delete(ctx, obj, opts...)
}

// Update (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return k.client.Update(ctx, obj, opts...)
}

// Get (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return k.client.Get(ctx, key, obj, opts...)
}

// List (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return k.client.List(ctx, list, opts...)
}

// Patch (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return k.client.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return k.client.DeleteAllOf(ctx, obj, opts...)
}

// Status (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Status() client.StatusWriter {
	return k.client.Status()
}

// Scheme (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) Scheme() *runtime.Scheme {
	return k.client.Scheme()
}

// RESTMapper (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) RESTMapper() meta.RESTMapper {
	return k.client.RESTMapper()
}

// SubResource (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) SubResource(subResource string) client.SubResourceClient {
	return k.client.SubResource(subResource)
}

// GroupVersionKindFor (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return k.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced (显式转发，或者完全依赖嵌入)
func (k *k8sClientImpl) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return k.client.IsObjectNamespaced(obj)
}

// GetCurrentNamespace 获取kubeconfig中配置的当前命名空间
func (k *k8sClientImpl) GetCurrentNamespace() (string, error) {
	// 如果没有rawConfig（例如，使用的是集群内配置），则返回空字符串
	if k.rawConfig == nil {
		return "", fmt.Errorf("no kubeconfig available")
	}

	// 从rawConfig获取命名空间
	namespace, _, err := k.rawConfig.Namespace()
	return namespace, err
}

// NewClient 创建新的Kubernetes客户端 (优化后)
// 同时初始化 controller-runtime client 和 client-go clientset
func NewClient(appCfg *config.Config) (KubernetesClient, error) {
	log := logger.GetLogger()

	// 1. 加载 REST 配置 (逻辑保持不变)
	var restConfig *rest.Config
	var err error
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if appCfg.Kubeconfig != "" {
		log.Debug("Using specified kubeconfig", "path", appCfg.Kubeconfig)
		loadingRules.ExplicitPath = appCfg.Kubeconfig
	} else {
		// 检查 KUBECONFIG 环境变量
		kubeconfigEnv := os.Getenv("KUBECONFIG")
		if kubeconfigEnv != "" {
			log.Debug("Using KUBECONFIG environment variable", "path", kubeconfigEnv)
			// clientcmd 会处理 : 分隔的路径列表
		} else {
			// 默认路径 ~/.kube/config
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Warn("Could not get user home directory, defaulting might fail", "error", err)
				// 继续尝试，也许在容器内或有其他方式配置
			} else {
				log.Debug("Using default kubeconfig path", "path", filepath.Join(homeDir, ".kube", "config"))
			}
		}
		// loadingRules 将自动处理环境变量和默认路径
	}

	// 从加载规则和空覆盖创建配置
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	restConfig, err = kubeConfig.ClientConfig()

	// 保存kubeConfig，以便之后GetCurrentNamespace使用
	var rawConfig clientcmd.ClientConfig
	if err == nil {
		rawConfig = kubeConfig
	}

	// 如果加载外部配置失败，尝试集群内配置
	if err != nil {
		log.Warn("Failed to load kubeconfig, trying in-cluster config", "error", err)
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			// 如果两种方式都失败，返回错误
			return nil, fmt.Errorf("could not configure Kubernetes client: %w", err)
		}
		log.Debug("Using in-cluster config")
	} else {
		log.Debug("Using out-of-cluster config")
	}

	// 2. 创建 Scheme (逻辑保持不变)
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %w", err)
	}
	// 添加自定义资源定义 (如果需要)

	// 3. 创建 controller-runtime Client
	runtimeClient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
		// 可以添加 MapperProvider 等其他选项
	})
	if err != nil {
		return nil, fmt.Errorf("could not create controller-runtime client: %w", err)
	}
	log.Debug("Controller-runtime client created")

	// 4. 创建 client-go Clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		// 如果 controller-runtime client 创建成功，但 clientset 创建失败，
		// 可能需要决定是否回滚或只返回部分功能的客户端。
		// 这里我们选择返回错误。
		return nil, fmt.Errorf("could not create kubernetes clientset: %w", err)
	}
	log.Debug("Kubernetes clientset created")

	// 5. 返回包含两个客户端的实现
	impl := &k8sClientImpl{
		client:    runtimeClient,
		clientset: clientset,
		rawConfig: rawConfig,
	}

	log.Info("Kubernetes client initialized successfully")
	return impl, nil
}

// InitializeDefaultClient 初始化默认客户端 (逻辑保持不变)
func InitializeDefaultClient(cfg *config.Config) error {
	var err error
	defaultClient, err = NewClient(cfg)
	return err
}

// GetClient 获取默认客户端实例 (逻辑保持不变)
func GetClient() KubernetesClient {
	// 考虑增加一个检查，如果 defaultClient 是 nil，则 panic 或返回错误
	if defaultClient == nil {
		panic("Default Kubernetes client accessed before initialization. Ensure InitializeDefaultClient() is called.")
	}
	return defaultClient
}
