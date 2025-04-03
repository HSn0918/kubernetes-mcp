# Kubernetes MCP 部署指南

本目录包含了部署 Kubernetes MCP 服务的所有必要配置文件和工具。

## 目录结构

```
deploy/
├── Makefile                # 主 Makefile，整合所有部署命令
├── docker/                 # Docker 相关配置
│   ├── Dockerfile          # Docker 构建文件
│   └── Makefile            # Docker 构建相关命令
└── kubernetes/             # Kubernetes 部署资源
    ├── 01-namespace-config.yaml  # 命名空间和配置映射
    ├── 02-rbac.yaml              # RBAC 权限配置
    ├── 03-deployment-service.yaml # 部署和服务配置
    ├── all-in-one.yaml           # 集成的一键部署配置
    └── kustomization.yaml        # Kustomize 配置
```

## 快速开始

### 部署到 Kubernetes

使用一键部署文件：

```bash
make k8s-deploy
```

或者使用 Kustomize：

```bash
make k8s-deploy-kustomize
```

### 构建 Docker 镜像

构建单架构 Docker 镜像：

```bash
make docker-build
```

构建并推送多架构 Docker 镜像：

```bash
make docker-buildx-push
```

### 完整部署流程

一键构建镜像并部署到 Kubernetes：

```bash
make full-deploy
```

多架构构建并部署：

```bash
make multi-arch-deploy
```

### 卸载服务

从 Kubernetes 删除：

```bash
make k8s-delete
```

## 配置说明

### Kubernetes 配置

- **01-namespace-config.yaml**: 定义 mcp-system 命名空间和配置项
- **02-rbac.yaml**: 定义服务所需的 RBAC 权限
- **03-deployment-service.yaml**: 定义部署和服务配置
- **all-in-one.yaml**: 整合以上所有配置为一个文件
- **kustomization.yaml**: 使用 Kustomize 进行配置组合

### Docker 配置

- **Dockerfile**: 多阶段构建 Docker 镜像配置
- **Makefile**: Docker 构建、推送相关命令

## 自定义配置

### 修改 ConfigMap

可以通过编辑 `kubernetes/01-namespace-config.yaml` 文件中的 ConfigMap 来修改服务配置：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubernetes-mcp-config
  namespace: mcp-system
data:
  port: "8080"              # 服务端口
  health-port: "8081"       # 健康检查端口
  log-level: "info"         # 日志级别
  log-format: "console"     # 日志格式
  allow-origins: "*"        # CORS 允许的来源
  base-url: "http://yoururl:8080"  # 服务 URL
```

### 修改镜像设置

可以在 `docker/Makefile` 中修改 Docker 镜像设置：

```makefile
DOCKER_IMAGE_NAME ?= hsn0918/$(APP_NAME)  # 镜像名称
PLATFORMS ?= linux/amd64,linux/arm64      # 支持的平台
```

## 帮助信息

查看所有可用命令：

```bash
make help
```