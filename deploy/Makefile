# Deploy 主 Makefile
# 用于整合所有部署和构建命令

# 包含 Docker 构建 Makefile
include docker/Makefile

# Kubernetes 部署相关目标
k8s-deploy:
	@echo ">>> 部署 kubernetes-mcp 到 Kubernetes 集群..."
	kubectl apply -f kubernetes/all-in-one.yaml

k8s-deploy-kustomize:
	@echo ">>> 使用 kustomize 部署 kubernetes-mcp 到 Kubernetes 集群..."
	kubectl apply -k kubernetes/

k8s-delete:
	@echo ">>> 从 Kubernetes 集群删除 kubernetes-mcp..."
	kubectl delete -f kubernetes/all-in-one.yaml

k8s-delete-kustomize:
	@echo ">>> 使用 kustomize 从 Kubernetes 集群删除 kubernetes-mcp..."
	kubectl delete -k kubernetes/

# 创建 namespace
k8s-create-namespace:
	@echo ">>> 创建 mcp-system 命名空间..."
	kubectl apply -f kubernetes/01-namespace-config.yaml

# 完整部署流程：构建镜像并部署到 Kubernetes
full-deploy: docker-build docker-push k8s-deploy
	@echo ">>> 完成构建镜像并部署到 Kubernetes"

# 多架构构建和部署
multi-arch-deploy: docker-buildx-push k8s-deploy
	@echo ">>> 完成多架构构建和部署"

# 显示帮助信息
help:
	@echo "部署工具使用方法: make [target]"
	@echo ""
	@echo "可用目标:"
	@echo "  k8s-deploy             使用 all-in-one.yaml 部署到 Kubernetes"
	@echo "  k8s-deploy-kustomize   使用 kustomize 部署到 Kubernetes"
	@echo "  k8s-delete             使用 all-in-one.yaml 从 Kubernetes 删除"
	@echo "  k8s-delete-kustomize   使用 kustomize 从 Kubernetes 删除"
	@echo "  k8s-create-namespace   仅创建 mcp-system 命名空间"
	@echo "  full-deploy            构建镜像并部署到 Kubernetes"
	@echo "  multi-arch-deploy      构建多架构镜像并部署到 Kubernetes"
	@echo ""
	@echo "Docker 相关目标:"
	@echo "  docker-build           构建单架构 Docker 镜像"
	@echo "  docker-push            推送本地构建的 Docker 镜像"
	@echo "  docker-buildx-push     构建并推送多架构 Docker 镜像"
	@echo "  docker-run-sse         运行 Docker 容器 (sse 模式)"

.PHONY: k8s-deploy k8s-deploy-kustomize k8s-delete k8s-delete-kustomize k8s-create-namespace full-deploy multi-arch-deploy help