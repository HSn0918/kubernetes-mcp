package config

const (
	TransportStdio      = "stdio"
	TransportSSE        = "sse"
	TransportStreamable = "streamable"
)

type Config struct {
	Transport    string
	Port         int
	HealthPort   int
	BaseURL      string
	AllowOrigins string
	LogLevel     string
	LogFormat    string
	Kubeconfig   string
}

func NewDefaultConfig() *Config {
	return &Config{
		Transport:    TransportSSE,
		Port:         8080,
		HealthPort:   8081,
		BaseURL:      "",
		AllowOrigins: "*",
		LogLevel:     "info",
		LogFormat:    "console",
		Kubeconfig:   "",
	}
}
