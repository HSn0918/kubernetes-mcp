package config

// Config 应用程序配置
type Config struct {
	// 服务器配置
	Transport  string
	Port       int
	HealthPort int
	// 日志配置
	LogLevel  string
	LogFormat string

	// Kubernetes配置
	Kubeconfig string
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Transport:  "stdio",
		Port:       8080,
		HealthPort: 8081,
		LogLevel:   "info",
		LogFormat:  "console",
		Kubeconfig: "",
	}
}
