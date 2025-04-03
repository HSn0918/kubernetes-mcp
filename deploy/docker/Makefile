# Docker 构建 Makefile

# 应用程序名称
APP_NAME := kubernetes-mcp

# Docker 相关变量
DOCKER_IMAGE_NAME ?= hsn0918/$(APP_NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
DOCKER_TAG := $(VERSION)
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
DOCKER_IMAGE_LATEST := $(DOCKER_IMAGE_NAME):latest

# Multi-arch 构建平台
PLATFORMS ?= linux/amd64,linux/arm64

# --- Docker 构建目标 ---

# 构建 Docker 镜像 (单架构)
docker-build:
	@echo ">>> Building single-arch Docker image ($(DOCKER_IMAGE))..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) \
		-t $(DOCKER_IMAGE_LATEST) \
		-f ./deploy/docker/Dockerfile \
		.
	@echo "Docker images built locally: $(DOCKER_IMAGE), $(DOCKER_IMAGE_LATEST)"

# 推送 Docker 镜像
docker-push:
	@echo ">>> Pushing Docker images $(DOCKER_IMAGE) and $(DOCKER_IMAGE_LATEST)..."
	@echo "Ensure you are logged in via 'docker login'"
	docker push $(DOCKER_IMAGE)
	docker push $(DOCKER_IMAGE_LATEST)

# 构建并推送多架构 Docker 镜像
docker-buildx-push:
	@echo ">>> Building and pushing multi-arch Docker image for platforms [$(PLATFORMS)]..."
	@echo "Ensure Buildx is setup and you are logged in via 'docker login'"
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) \
		-t $(DOCKER_IMAGE_LATEST) \
		-f ./deploy/docker/Dockerfile \
		.
	@echo "Multi-arch Docker images pushed: $(DOCKER_IMAGE), $(DOCKER_IMAGE_LATEST)"

# 运行 Docker 容器 (sse 模式)
docker-run-sse:
	@echo ">>> Running Docker container in sse mode on port 8080..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE_LATEST) server --transport=sse --port=8080

# 帮助信息
help:
	@echo "Docker Build Targets:"
	@echo "  docker-build         Build single-arch Docker image for the host architecture"
	@echo "  docker-push          Push the locally built Docker images"
	@echo "  docker-buildx-push   Build multi-arch Docker image (via buildx) and push"
	@echo "  docker-run-sse       Run the Docker container in sse mode (port 8080)"

.PHONY: docker-build docker-push docker-buildx-push docker-run-sse help