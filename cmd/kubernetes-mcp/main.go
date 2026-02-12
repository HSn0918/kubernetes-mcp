package main

import (
	"os"

	"github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app"
	"github.com/hsn0918/kubernetes-mcp/pkg/client/kubernetes"
	"github.com/hsn0918/kubernetes-mcp/pkg/config"
	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

func main() {
	cfg := config.NewDefaultConfig()
	logger.InitializeDefaultLogger(cfg.LogLevel, cfg.LogFormat)
	log := logger.GetLogger()
	if err := kubernetes.InitializeDefaultClient(cfg); err != nil {
		log.Error("Failed to initialize Kubernetes client", logger.Any("error", err))
		os.Exit(1)
	}
	rootCmd := app.NewRootCommand(cfg)
	if err := rootCmd.Execute(); err != nil {
		log.Error("Failed to execute root command", logger.Any("error", err))
		os.Exit(1)
	}
}
