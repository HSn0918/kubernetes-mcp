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
# Docker 命令
DOCKER := docker
# Docker Buildx 命令
BUILDX := docker buildx

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

# Docker 镜像名称和标签
DOCKER_IMAGE_NAME ?= your-dockerhub-username/$(APP_NAME) # 或者其他镜像仓库/名称
DOCKER_TAG := $(VERSION)
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
DOCKER_IMAGE_LATEST := $(DOCKER_IMAGE_NAME):latest

# Multi-arch build platforms (可以覆盖, e.g., make docker-buildx-push PLATFORMS=linux/amd64)
PLATFORMS ?= linux/amd64,linux/arm64

# --- Targets ---

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

# 构建 Docker 镜像 (单架构，用于本地主机)
docker-build:
	@echo ">>> Building single-arch Docker image for host ($(DOCKER_IMAGE))..."
	$(DOCKER) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) \
		-t $(DOCKER_IMAGE_LATEST) \
		.
		# --load # 可选，通常是 build 的默认行为，确保镜像加载到本地
	@echo "Docker images built locally: $(DOCKER_IMAGE), $(DOCKER_IMAGE_LATEST)"

# 推送本地构建的单架构 Docker 镜像 (需要先运行 docker-build)
docker-push:
	@echo ">>> Pushing single-arch Docker images $(DOCKER_IMAGE) and $(DOCKER_IMAGE_LATEST)..."
	@echo "Ensure you are logged in via 'docker login' to the target registry for $(DOCKER_IMAGE_NAME)"
	$(DOCKER) push $(DOCKER_IMAGE)
	$(DOCKER) push $(DOCKER_IMAGE_LATEST)

# 构建并推送多架构 Docker 镜像 (使用 buildx)
# 需要 Buildx 环境设置 (e.g., docker buildx create --use mybuilder)
# 需要登录到目标 registry (docker login)
docker-buildx-push:
	@echo ">>> Building and pushing multi-arch Docker image for platforms [$(PLATFORMS)] ($(DOCKER_IMAGE))..."
	@echo "Ensure Buildx is setup and you are logged in via 'docker login' to the target registry for $(DOCKER_IMAGE_NAME)"
	$(BUILDX) build \
		--platform $(PLATFORMS) \
		--push \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) \
		-t $(DOCKER_IMAGE_LATEST) \
		.
	@echo "Multi-arch Docker images pushed: $(DOCKER_IMAGE), $(DOCKER_IMAGE_LATEST)"

# 运行 Docker 容器 (stdio 模式, 交互式) - 使用本地单架构镜像
docker-run-stdio: docker-build
	@echo ">>> Running Docker container in stdio mode (interactive)..."
	$(DOCKER) run --rm -it $(DOCKER_IMAGE) server --transport=stdio

# 运行 Docker 容器 (sse 模式, 端口映射) - 使用本地单架构镜像
# 注意：可能需要挂载 kubeconfig 才能连接到集群
# 例如: -v ~/.kube:/root/.kube:ro (假设容器用户是 root)
docker-run-sse: docker-build
	@echo ">>> Running Docker container in sse mode on port 8080..."
	$(DOCKER) run --rm -p 8080:8080 $(DOCKER_IMAGE) server --transport=sse --port=8080
	# Example with kubeconfig mount:
	# $(DOCKER) run --rm -p 8080:8080 -v ~/.kube:/root/.kube:ro $(DOCKER_IMAGE) server --transport=sse --port=8080 --kubeconfig=/root/.kube/config

# 显示帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all                 Build the binary (default)"
	@echo "  build               Build the Go binary"
	@echo "  test                Run Go tests"
	@echo "  clean               Remove build artifacts"
	@echo "  run-stdio           Run the locally built binary in stdio mode"
	@echo "  run-sse             Run the locally built binary in sse mode (port 8080)"
	@echo "  docker-build        Build single-arch Docker image for the host architecture (loads locally)"
	@echo "  docker-push         Push the locally built single-arch image (run docker-build first, requires login)"
	@echo "  docker-buildx-push  Build multi-arch Docker image (via buildx) and push (requires buildx setup & login)"
	@echo "                      Override platforms: make docker-buildx-push PLATFORMS=linux/amd64,linux/arm/v7"
	@echo "  docker-run-stdio    Run the Docker container (local image) in stdio mode"
	@echo "  docker-run-sse      Run the Docker container (local image) in sse mode (port 8080)"
	@echo "  help                Show this help message"

# 声明伪目标 (这些目标不代表文件)
.PHONY: all build test clean run-stdio run-sse docker-build docker-push docker-buildx-push docker-run-stdio docker-run-sse help
