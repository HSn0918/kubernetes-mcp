package kubernetes

import (
	"context"
	"fmt"

	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	client.Client
	ClientSet() kubernetes.Interface
	GetCurrentNamespace() (string, error)
	GetDynamicClient() dynamic.Interface
	GetDiscoveryClient() discovery.DiscoveryInterface
	GetMetricsClient() metricsv.Interface
	GetConfig() clientcmd.ClientConfig
}

type k8sClientImpl struct {
	client.Client
	clientset       kubernetes.Interface
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	metricsClient   metricsv.Interface
	rawConfig       clientcmd.ClientConfig
}

var _ Client = &k8sClientImpl{}

var defaultClient Client

func (k *k8sClientImpl) ClientSet() kubernetes.Interface {
	if k.clientset == nil {
		panic("internal error: kubernetes clientset was not initialized")
	}
	return k.clientset
}

func (k *k8sClientImpl) GetCurrentNamespace() (string, error) {
	if k.rawConfig == nil {
		return "", fmt.Errorf("kubeconfig is not available (possibly using in-cluster config)")
	}
	namespace, _, err := k.rawConfig.Namespace()
	if err != nil {
		return "", fmt.Errorf("failed to get namespace from kubeconfig: %w", err)
	}
	return namespace, nil
}

func (k *k8sClientImpl) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}

func (k *k8sClientImpl) GetDiscoveryClient() discovery.DiscoveryInterface {
	return k.discoveryClient
}

func (k *k8sClientImpl) GetMetricsClient() metricsv.Interface {
	return k.metricsClient
}

func (k *k8sClientImpl) GetConfig() clientcmd.ClientConfig {
	return k.rawConfig
}

func NewClient(appCfg *config.Config) (Client, error) {
	log := logger.GetLogger()
	log.Info("Initializing Kubernetes client")

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if appCfg.Kubeconfig != "" {
		loadingRules.ExplicitPath = appCfg.Kubeconfig
		log.Debug("Using kubeconfig from flag", logger.String("path", appCfg.Kubeconfig))
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	restConfig, err := kubeConfig.ClientConfig()
	var rawConfig clientcmd.ClientConfig
	if err == nil {
		rawConfig = kubeConfig
	} else {
		log.Warn("Failed to load kubeconfig, trying in-cluster config", logger.Any("error", err))
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("could not configure Kubernetes client: %w", err)
		}
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %w", err)
	}

	restConfig.QPS = 500
	restConfig.Burst = 1000

	runtimeClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("could not create controller-runtime client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes clientset: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create discovery client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create dynamic client: %w", err)
	}

	metricsClient, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create metrics client: %w", err)
	}

	impl := &k8sClientImpl{
		Client:          runtimeClient,
		clientset:       clientset,
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
		metricsClient:   metricsClient,
		rawConfig:       rawConfig,
	}

	log.Info("Kubernetes client initialized")
	return impl, nil
}

func InitializeDefaultClient(cfg *config.Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize default Kubernetes client: %w", err)
	}
	defaultClient = client
	return nil
}

func GetClient() Client {
	if defaultClient == nil {
		panic("default Kubernetes client accessed before initialization")
	}
	return defaultClient
}

func (k *k8sClientImpl) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return k.Client.Apply(ctx, obj, opts...)
}
