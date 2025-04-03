package main

import (
	"os"

	"github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app"
	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

func main() {
	// 初始化配置
	cfg := config.NewDefaultConfig()

	// 初始化日志
	logger.InitializeDefaultLogger(cfg.LogLevel, cfg.LogFormat)
	log := logger.GetLogger()

	// 初始化客户端
	if err := kubernetes.InitializeDefaultClient(cfg); err != nil {
		log.Error("Failed to initialize Kubernetes client", "error", err)
		os.Exit(1)
	}

	// 创建命令行应用
	rootCmd := app.NewRootCommand(cfg)

	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		log.Error("Failed to execute root command", "error", err)
		os.Exit(1)
	}
}
