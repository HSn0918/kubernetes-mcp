# Makefile

# --- Variables ---
# 应用程序名称
APP_NAME := kubernetes-mcp
# 主命令路径
CMD_PATH := ./cmd/kubernetes-mcp/main.go
# 输出目录
OUTPUT_DIR := ./bin
# 输出二进制文件路径
BINARY := $(OUTPUT_DIR)/$(APP_NAME)

# Go 命令
GO := go

# 构建版本信息 (尝试从 Git 获取，如果失败则使用默认值)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# ldflags 用于注入版本信息 (确保包路径正确!)
# 注意：包路径是定义 Version, Commit, BuildDate 变量的包
VERSION_PKG_PATH := github.com/hsn0918/kubernetes-mcp/cmd/kubernetes-mcp/app
LDFLAGS := -w -s \
           -X $(VERSION_PKG_PATH).Version=$(VERSION) \
           -X $(VERSION_PKG_PATH).Commit=$(COMMIT) \
           -X $(VERSION_PKG_PATH).BuildDate=$(BUILD_DATE)

# 部署目录变量
DEPLOY_DIR := ./deploy

# --- Go 构建目标 ---

# 默认目标
all: build

# 构建 Go 二进制文件
build:
	@echo ">>> Building $(APP_NAME) binary..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD_PATH)
	@echo "Binary available at $(BINARY)"

# 运行 Go 测试
test:
	@echo ">>> Running tests..."
	$(GO) test ./... -v

# 清理构建产物
clean:
	@echo ">>> Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)

# 运行本地构建的二进制文件 (stdio 模式)
run-stdio: build
	@echo ">>> Running $(APP_NAME) in stdio mode..."
	$(BINARY) server --transport=stdio

# 运行本地构建的二进制文件 (sse 模式)
run-sse: build
	@echo ">>> Running $(APP_NAME) in sse mode on port 8080..."
	$(BINARY) server --transport=sse --port=8080

# --- 部署相关目标（重定向到deploy目录） ---

# Docker 构建 (重定向到deploy目录)
docker-build: export VERSION=$(VERSION)
docker-build: export COMMIT=$(COMMIT)
docker-build: export BUILD_DATE=$(BUILD_DATE)
docker-build: build
	@echo ">>> 重定向到deploy目录的Docker构建..."
	$(MAKE) -C $(DEPLOY_DIR) docker-build

# Docker 推送
docker-push:
	@echo ">>> 重定向到deploy目录的Docker推送..."
	$(MAKE) -C $(DEPLOY_DIR) docker-push

# 多架构Docker构建并推送
docker-buildx-push: export VERSION=$(VERSION)
docker-buildx-push: export COMMIT=$(COMMIT)
docker-buildx-push: export BUILD_DATE=$(BUILD_DATE)
docker-buildx-push: build
	@echo ">>> 重定向到deploy目录的多架构Docker构建与推送..."
	$(MAKE) -C $(DEPLOY_DIR) docker-buildx-push

# 运行Docker容器 (sse模式)
docker-run-sse:
	@echo ">>> 重定向到deploy目录运行Docker容器..."
	$(MAKE) -C $(DEPLOY_DIR) docker-run-sse

# 运行Docker容器 (stdio模式)
docker-run-stdio:
	@echo ">>> 重定向到deploy目录运行Docker容器(stdio模式)..."
	$(MAKE) -C $(DEPLOY_DIR) docker-run-sse

# --- Kubernetes 部署目标 ---

# 部署到Kubernetes
k8s-deploy:
	@echo ">>> 重定向到deploy目录部署到Kubernetes..."
	$(MAKE) -C $(DEPLOY_DIR) k8s-deploy

# 使用kustomize部署到Kubernetes
k8s-deploy-kustomize:
	@echo ">>> 重定向到deploy目录使用kustomize部署到Kubernetes..."
	$(MAKE) -C $(DEPLOY_DIR) k8s-deploy-kustomize

# 从Kubernetes删除
k8s-delete:
	@echo ">>> 重定向到deploy目录从Kubernetes删除..."
	$(MAKE) -C $(DEPLOY_DIR) k8s-delete

# 创建命名空间
k8s-create-namespace:
	@echo ">>> 重定向到deploy目录创建命名空间..."
	$(MAKE) -C $(DEPLOY_DIR) k8s-create-namespace

# 完整部署流程：构建镜像并部署到Kubernetes
full-deploy: build
	@echo ">>> 重定向到deploy目录进行完整部署流程..."
	$(MAKE) -C $(DEPLOY_DIR) full-deploy

# 多架构构建和部署
multi-arch-deploy: build
	@echo ">>> 重定向到deploy目录进行多架构构建和部署..."
	$(MAKE) -C $(DEPLOY_DIR) multi-arch-deploy

# 显示帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Go构建目标:"
	@echo "  all                 构建二进制文件 (默认)"
	@echo "  build               构建Go二进制文件"
	@echo "  test                运行Go测试"
	@echo "  clean               清理构建产物"
	@echo "  run-stdio           以stdio模式运行本地二进制文件"
	@echo "  run-sse             以sse模式运行本地二进制文件 (端口8080)"
	@echo ""
	@echo "Docker相关目标:"
	@echo "  docker-build        构建Docker镜像"
	@echo "  docker-push         推送Docker镜像"
	@echo "  docker-buildx-push  构建并推送多架构Docker镜像"
	@echo "  docker-run-stdio    运行Docker容器 (stdio模式)"
	@echo "  docker-run-sse      运行Docker容器 (sse模式)"
	@echo ""
	@echo "Kubernetes部署目标:"
	@echo "  k8s-deploy          部署到Kubernetes集群"
	@echo "  k8s-deploy-kustomize 使用kustomize部署到Kubernetes集群"
	@echo "  k8s-delete          从Kubernetes集群删除"
	@echo "  k8s-create-namespace 创建命名空间"
	@echo "  full-deploy         构建镜像并部署到Kubernetes"
	@echo "  multi-arch-deploy   构建多架构镜像并部署到Kubernetes"
	@echo ""
	@echo "详细的部署命令请查看 $(DEPLOY_DIR)/README.md"

# 声明伪目标 (这些目标不代表文件)
.PHONY: all build test clean run-stdio run-sse \
        docker-build docker-push docker-buildx-push docker-run-stdio docker-run-sse \
        k8s-deploy k8s-deploy-kustomize k8s-delete k8s-create-namespace \
        full-deploy multi-arch-deploy help
