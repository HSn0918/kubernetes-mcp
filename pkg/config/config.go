package config

// Config 应用程序配置
type Config struct {
	// 服务器配置
	Transport  string
	Port       int
	HealthPort int
	BaseURL    string
	// CORS配置
	AllowOrigins string
	// 日志配置
	LogLevel  string
	LogFormat string
	// Kubernetes配置
	Kubeconfig string
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Transport:    "sse",
		Port:         8080,
		HealthPort:   8081,
		BaseURL:      "",
		AllowOrigins: "*",
		LogLevel:     "info",
		LogFormat:    "console",
		Kubeconfig:   "",
	}
}
