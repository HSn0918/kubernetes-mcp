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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubernetesClient 定义了与 Kubernetes API 交互的接口。
// 它封装了 controller-runtime 的 client.Client 和 client-go 的核心客户端功能。
type KubernetesClient interface {
	// 嵌入 controller-runtime 的 Client 接口，提供核心的 CRUD 等面向对象的操作。
	client.Client
	// ClientSet 提供访问标准 client-go Clientset 的方法。
	// 这对于执行 controller-runtime 不直接支持的操作（例如获取 Pod 日志）非常有用。
	ClientSet() kubernetes.Interface
	// GetCurrentNamespace 获取当前 kubeconfig 文件中配置的默认命名空间。
	// 如果使用的是集群内配置或无法确定命名空间，则可能返回错误。
	GetCurrentNamespace() (string, error)
	// GetDynamicClient 提供访问 client-go 动态客户端的方法。
	// 动态客户端允许操作任意 Kubernetes 资源，包括 CRD，而无需编译时类型信息。
	GetDynamicClient() dynamic.Interface
	// GetDiscoveryClient 提供访问 client-go discovery 客户端的方法。
	// Discovery 客户端用于发现 Kubernetes API Server 支持的 API 组、版本和资源。
	GetDiscoveryClient() discovery.DiscoveryInterface
	// GetMetricsClient 提供访问 client-go metrics 客户端的方法。
	// Metrics 客户端用于获取 Kubernetes 资源的度量信息。
	GetMetricsClient() metricsv.Interface
	// GetConfig 获取用于创建此客户端的原始 clientcmd 配置。
	// 这对于需要访问底层配置细节（如上下文、集群信息等）的场景很有用。
	GetConfig() clientcmd.ClientConfig
}

// k8sClientImpl 是 KubernetesClient 接口的具体实现。
// 它聚合了 controller-runtime client 和 client-go 的各种客户端实例。
type k8sClientImpl struct {
	// controller-runtime 客户端，用于通用的对象操作。
	client client.Client
	// 标准的 client-go Clientset，用于特定或底层操作。
	clientset kubernetes.Interface
	// 动态客户端，用于处理 CRD 或非结构化数据。
	dynamicClient dynamic.Interface
	// Discovery 客户端，用于 API 发现。
	discoveryClient discovery.DiscoveryInterface
	// Metrics 客户端，用于获取 Kubernetes 资源的度量信息。
	metricsClient metricsv.Interface
	// 加载的原始 kubeconfig 配置信息。
	rawConfig clientcmd.ClientConfig
}

// 编译时断言，确保 k8sClientImpl 实现了 KubernetesClient 接口。
var _ KubernetesClient = &k8sClientImpl{}

// defaultClient 是一个全局的 KubernetesClient 实例，通过 InitializeDefaultClient 初始化。
// 使用 GetClient() 函数来安全地访问此实例。
var defaultClient KubernetesClient

// ClientSet 返回初始化时创建并存储的 client-go Clientset。
// 这是 KubernetesClient 接口的实现方法。
func (k *k8sClientImpl) ClientSet() kubernetes.Interface {
	// 注意：此实现假设 k.clientset 在 NewClient 中已被成功初始化。
	// 如果在 NewClient 中初始化失败，则 NewClient 会返回错误，不会创建 k8sClientImpl 实例。
	if k.clientset == nil {
		// 理论上不应发生此情况，因为 NewClient 会确保 clientset 被初始化或返回错误。
		// 添加 panic 以在开发阶段捕捉意外状态。
		panic("内部错误：kubernetes clientset 在 k8sClientImpl 实例中未被初始化")
	}
	return k.clientset
}

// --- controller-runtime client.Client 接口方法的实现 (通过显式转发到嵌入的 k.client) ---
// 下面的方法显式地将调用转发给嵌入的 k.client。
// 虽然 Go 的嵌入会自动提供这些方法，但显式转发可以更清晰地表明意图，
// 并且允许在未来添加额外的逻辑（例如日志记录、度量）。

// Create 调用嵌入的 controller-runtime 客户端的 Create 方法。
func (k *k8sClientImpl) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return k.client.Create(ctx, obj, opts...)
}

// Delete 调用嵌入的 controller-runtime 客户端的 Delete 方法。
func (k *k8sClientImpl) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return k.client.Delete(ctx, obj, opts...)
}

// Update 调用嵌入的 controller-runtime 客户端的 Update 方法。
func (k *k8sClientImpl) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return k.client.Update(ctx, obj, opts...)
}

// Get 调用嵌入的 controller-runtime 客户端的 Get 方法。
func (k *k8sClientImpl) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return k.client.Get(ctx, key, obj, opts...)
}

// List 调用嵌入的 controller-runtime 客户端的 List 方法。
func (k *k8sClientImpl) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return k.client.List(ctx, list, opts...)
}

// Patch 调用嵌入的 controller-runtime 客户端的 Patch 方法。
func (k *k8sClientImpl) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return k.client.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf 调用嵌入的 controller-runtime 客户端的 DeleteAllOf 方法。
func (k *k8sClientImpl) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return k.client.DeleteAllOf(ctx, obj, opts...)
}

// Status 返回一个用于更新对象状态子资源的 StatusWriter。
func (k *k8sClientImpl) Status() client.StatusWriter {
	return k.client.Status()
}

// Scheme 返回与此客户端关联的 runtime.Scheme。
func (k *k8sClientImpl) Scheme() *runtime.Scheme {
	return k.client.Scheme()
}

// RESTMapper 返回用于 GVK (GroupVersionKind) 和资源之间映射的 RESTMapper。
func (k *k8sClientImpl) RESTMapper() meta.RESTMapper {
	return k.client.RESTMapper()
}

// SubResource 返回一个用于操作指定子资源的 SubResourceClient。
func (k *k8sClientImpl) SubResource(subResource string) client.SubResourceClient {
	return k.client.SubResource(subResource)
}

// GroupVersionKindFor 尝试为给定的 runtime.Object 确定其 GroupVersionKind。
func (k *k8sClientImpl) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return k.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced 检查给定的 runtime.Object 是否是命名空间作用域的资源。
func (k *k8sClientImpl) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return k.client.IsObjectNamespaced(obj)
}

// GetCurrentNamespace 获取 kubeconfig 中配置的当前命名空间。
// 这是 KubernetesClient 接口的实现方法。
func (k *k8sClientImpl) GetCurrentNamespace() (string, error) {
	// 如果 k.rawConfig 为 nil (例如，使用集群内配置时)，则无法从 kubeconfig 文件获取命名空间。
	if k.rawConfig == nil {
		// 对于集群内配置，通常认为命名空间是 Pod 运行所在的命名空间，
		// 但这需要通过 downward API 或其他方式获取，而不是通过 ClientConfig。
		// 在这里返回错误表明无法从配置中确定命名空间。
		return "", fmt.Errorf("kubeconfig is not available (possibly using in-cluster config)")
	}

	// 尝试从原始 clientcmd 配置中获取命名空间。
	// 第三个返回值 (bool) 表示命名空间是否在配置中被显式设置。
	namespace, _, err := k.rawConfig.Namespace()
	if err != nil {
		return "", fmt.Errorf("failed to get namespace from kubeconfig: %w", err)
	}
	// 如果 kubeconfig 中没有指定命名空间，默认通常是 "default"。
	// Namespace() 方法会处理这种情况。
	return namespace, nil
}

// NewClient 创建并返回一个新的 KubernetesClient 实例。
// 它会根据提供的配置加载 Kubernetes 配置，并初始化所有必需的客户端。
func NewClient(appCfg *config.Config) (KubernetesClient, error) {
	// 获取日志记录器实例
	log := logger.GetLogger()
	log.Info("Initializing Kubernetes client...")

	// 1. 加载 Kubernetes REST 配置
	var restConfig *rest.Config
	var err error
	// 创建默认的 kubeconfig 加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// 优先使用 appCfg.Kubeconfig 指定的路径
	if appCfg.Kubeconfig != "" {
		log.Debug("Using specific kubeconfig path", "path", appCfg.Kubeconfig)
		loadingRules.ExplicitPath = appCfg.Kubeconfig
	} else {
		// 如果未指定路径，则遵循标准加载顺序：
		// 1. KUBECONFIG 环境变量
		kubeconfigEnv := os.Getenv("KUBECONFIG")
		if kubeconfigEnv != "" {
			log.Debug("Using KUBECONFIG environment variable", "path", kubeconfigEnv)
			// loadingRules 会自动处理 KUBECONFIG
		} else {
			// 2. 默认路径 ~/.kube/config
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Warn("Could not get user home directory, defaulting kubeconfig path might fail", "error", err)
				// 即使无法获取主目录，仍然尝试加载，clientcmd 可能会处理其他情况
			} else {
				defaultPath := filepath.Join(homeDir, ".kube", "config")
				log.Debug("Using default kubeconfig path", "path", defaultPath)
				// loadingRules 会自动检查默认路径
			}
		}
	}

	// 创建 clientcmd 配置对象，它会根据加载规则和覆盖项延迟加载配置
	// 使用空的覆盖项 &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	// 从 clientcmd 配置对象获取 REST 配置
	restConfig, err = kubeConfig.ClientConfig()

	// 保存加载的原始 clientcmd 配置，即使之后使用集群内配置，这个也可能包含上下文信息
	// 如果 ClientConfig() 成功，则保存 kubeConfig
	var rawConfig clientcmd.ClientConfig
	if err == nil {
		rawConfig = kubeConfig
	} else {
		rawConfig = nil // 明确设为 nil，如果加载失败
	}

	// 如果从外部文件加载配置失败，尝试使用集群内配置 (适用于在 Kubernetes Pod 中运行的场景)
	if err != nil {
		log.Warn("Failed to load kubeconfig from file/env, attempting in-cluster config", "error", err)
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			// 如果两种方式都失败，则无法连接到集群，返回错误
			return nil, fmt.Errorf("could not configure Kubernetes client: failed to load both out-of-cluster and in-cluster config: %w", err)
		}
		log.Debug("Successfully loaded in-cluster config")
		// 使用集群内配置时，rawConfig 仍然是 nil
	} else {
		log.Debug("Successfully loaded out-of-cluster config")
	}

	// 2. 创建 runtime.Scheme 用于类型注册
	scheme := runtime.NewScheme()
	// 将 Kubernetes 内建类型（如 Pod, Service 等）添加到 Scheme
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %w", err)
	}
	// TODO: 在这里可以添加应用程序自定义资源 (CRD) 的类型到 Scheme
	// 例如: import mycrdscheme "my/api/v1"
	//       mycrdscheme.AddToScheme(scheme)

	// 调整客户端的 QPS (每秒查询数) 和 Burst (峰值并发数)，以控制请求速率
	// 增加这些值可以提高吞吐量，但需注意 API Server 的承受能力
	restConfig.QPS = 500
	restConfig.Burst = 1000
	log.Debug("Set client QPS and Burst", "qps", restConfig.QPS, "burst", restConfig.Burst)

	// 3. 创建 controller-runtime Client
	// 这个客户端提供了更高级别的、面向对象的 API
	runtimeClient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
		// MapperProvider: client.NewLazyRESTMapperLoader(restConfig), // 可以考虑使用 Lazy Mapper
	})
	if err != nil {
		return nil, fmt.Errorf("could not create controller-runtime client: %w", err)
	}
	log.Debug("Controller-runtime client created successfully")

	// 4. 创建 client-go Clientset
	// 这是标准的 Kubernetes Go 客户端，提供了访问各种 API 组的接口
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		// 如果 controller-runtime client 创建成功，但 clientset 创建失败，
		// 最好是返回错误，因为客户端功能不完整。
		return nil, fmt.Errorf("could not create kubernetes clientset: %w", err)
	}
	log.Debug("Kubernetes clientset created successfully")

	// 5. 创建 DiscoveryClient 和 DynamicClient 和 metricsClient 实例
	// DiscoveryClient 用于发现 API 资源
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create discovery client: %w", err)
	}
	log.Debug("Discovery client created successfully")
	// DynamicClient 用于操作非结构化数据（例如 CRD）
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create dynamic client: %w", err)
	}
	log.Debug("Dynamic client created successfully")
	metricsClient, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create metrics client: %w", err)
	}
	// 6. 创建并返回 k8sClientImpl 实例
	impl := &k8sClientImpl{
		client:          runtimeClient,
		clientset:       clientset,
		rawConfig:       rawConfig, // 注意这里保存的是 ClientConfig 接口，可能是 nil
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
		metricsClient:   metricsClient,
	}

	log.Info("Kubernetes client initialized successfully")
	return impl, nil
}

// InitializeDefaultClient 使用提供的配置初始化全局默认客户端实例。
// 这个函数应该在应用程序启动时调用一次。
// 返回的错误表示初始化过程中是否发生问题。
func InitializeDefaultClient(cfg *config.Config) error {
	var err error
	// 调用 NewClient 创建新的客户端实例
	defaultClient, err = NewClient(cfg)
	if err != nil {
		// 如果创建失败，返回错误
		return fmt.Errorf("failed to initialize default Kubernetes client: %w", err)
	}
	// 如果成功，全局 defaultClient 变量已被设置
	return nil
}

// GetClient 返回全局默认的 KubernetesClient 实例。
// 在调用此函数之前，必须先成功调用 InitializeDefaultClient。
// 如果 defaultClient 尚未初始化，此函数会触发 panic。
func GetClient() KubernetesClient {
	// 添加检查确保 defaultClient 已经被初始化
	if defaultClient == nil {
		// 触发 panic，强制要求开发者在使用前必须先初始化
		panic("Fatal Error: Default Kubernetes client accessed before initialization. Ensure InitializeDefaultClient() is called successfully at application startup.")
	}
	return defaultClient
}

// GetDynamicClient 返回 k8sClientImpl 实例中的动态客户端。
// 这是 KubernetesClient 接口的实现方法。
func (k *k8sClientImpl) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}

// GetDiscoveryClient 返回 k8sClientImpl 实例中的 Discovery 客户端。
// 这是 KubernetesClient 接口的实现方法。
func (k *k8sClientImpl) GetDiscoveryClient() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *k8sClientImpl) GetMetricsClient() metricsv.Interface {
	return k.metricsClient
}

// GetConfig 返回 k8sClientImpl 实例中存储的原始 clientcmd 配置。
// 这是 KubernetesClient 接口的实现方法。
func (k *k8sClientImpl) GetConfig() clientcmd.ClientConfig {
	return k.rawConfig
}
