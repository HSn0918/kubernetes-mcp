# Kubernetes-MCP

[English](README.md) | 中文

一个使用 Go 语言设计的 Model Capable Protocol (MCP) 服务器实现，用于与 Kubernetes 集群交互。该服务器允许 MCP 兼容的客户端通过定义的工具执行 Kubernetes 操作。

## ✨ 功能特点

* **MCP 服务器：** 实现 `mcp-go` 库以提供 MCP 功能。
* **Kubernetes 交互：** 使用 `controller-runtime` 客户端与 Kubernetes 集群交互。
* **多种传输方式：** 支持通过标准 I/O（`stdio`）或服务器发送事件（`sse`）进行通信。
* **资源管理工具：** 公开用于 Kubernetes 操作的 MCP 工具：
    * **核心 API 组 (v1)：**
        * 已完整实现：列出命名空间作用域资源（`listResources`），获取资源 YAML（`getResource`），从 YAML 创建（`createResource`），从 YAML 更新（`updateResource`），删除资源（`deleteResource`）。
        * 已完整实现：列出集群作用域命名空间（`listNamespaces`）。
    * **Apps API 组 (apps/v1)：**
        * 已实现：列出命名空间作用域资源（`listAppsResources`）。
        * 占位符（未实现）：获取（`getAppsResource`），创建（`createAppsResource`），更新（`updateAppsResource`），删除（`deleteAppsResource`）。
    * **Batch API 组 (batch/v1)：**
        * 占位符（未实现）：列出（`listBatchResources`），获取（`getBatchResource`），创建（`createBatchResource`），更新（`updateBatchResource`），删除（`deleteBatchResource`）。
    * **Networking API 组 (networking.k8s.io/v1)：**
        * 占位符（未实现）：列出（`listNetworkingResources`），获取（`getNetworkingResource`），创建（`createNetworkingResource`），更新（`updateNetworkingResource`），删除（`deleteNetworkingResource`）。
* **配置：** 可通过命令行标志配置（传输方式、端口、kubeconfig、日志级别/格式）。
* **日志记录：** 使用 `zap` 进行结构化日志记录。
* **命令行界面：** 使用 `cobra` 框架构建。

## 📋 前提条件

* **Go 1.24**
* 访问 Kubernetes 集群（通过 `kubeconfig` 文件或集群内服务账户）。

## 📦 主要依赖

本项目依赖于几个关键的 Go 模块：

* `github.com/mark3labs/mcp-go`（用于 MCP 服务器/协议）
* `sigs.k8s.io/controller-runtime`（用于 Kubernetes 客户端交互）
* `k8s.io/client-go`（核心 Kubernetes 库）
* `github.com/spf13/cobra`（用于 CLI 结构）
* `go.uber.org/zap`（用于日志记录）
* `sigs.k8s.io/yaml`（用于 YAML 处理）

*（注：具体版本在未提供的 `go.mod` 中管理）*

## 🔨 构建

### 从源代码构建

构建可执行文件：

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
./kubernetes-mcp server --transport=sse --port 8080
```

### 使用 Docker

构建 Docker 镜像：

```bash
docker build -t kubernetes-mcp:latest \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") .
```

使用 Docker 运行：

```bash
# 使用 stdio 传输方式（默认）
docker run -v ~/.kube:/root/.kube kubernetes-mcp:latest

# 使用 SSE 传输方式
docker run -p 8080:8080 -v ~/.kube:/root/.kube kubernetes-mcp:latest server --transport=sse

# 查看版本信息
docker run kubernetes-mcp:latest version

# 指定自定义 kubeconfig
docker run -v /path/to/config:/config kubernetes-mcp:latest server --kubeconfig=/config
```

## 🚀 使用方法

使用 `server` 子命令启动服务器。

### 使用标准 I/O（stdio - 默认）：
```shell
./kubernetes-mcp server
```

### 使用服务器发送事件（SSE）：
```shell
./kubernetes-mcp server --transport sse --port 8080
```
（使用 SSE 时默认监听 8080 端口）

### 指定 Kubeconfig：

如果您的配置文件不在标准位置，请使用 `--kubeconfig` 标志：
```shell
./kubernetes-mcp server --kubeconfig /path/to/your/kubeconfig
```

### 查看版本：

显示构建时设置的版本信息。
```shell
./kubernetes-mcp version
# 输出示例：Kubernetes-mcp version dev (commit: none, build date: unknown)
```

## ⚙️ 配置标志
- `--transport`：通信模式。选项：stdio（默认），sse。
- `--port`：SSE 传输的端口号。默认：8080。
- `--kubeconfig`：Kubernetes 配置文件的路径。（默认为标准发现：KUBECONFIG 环境变量或 ~/.kube/config）。
- `--log-level`：日志详细程度。选项：debug，info（默认），warn，error。
- `--log-format`：日志输出格式。选项：console（默认），json。

## 🧩 支持的 MCP 工具（Kubernetes 操作）

以下基于 Kubernetes API 组和操作注册了 MCP 工具：

### 核心 API 组 (v1)
✅ `listResources`：列出核心 v1 命名空间作用域资源（Pod、Service 等）  
✅ `getResource`：获取核心 v1 命名空间作用域资源 YAML  
✅ `createResource`：从 YAML 创建核心 v1 命名空间作用域资源  
✅ `updateResource`：从 YAML 更新核心 v1 命名空间作用域资源  
✅ `deleteResource`：删除核心 v1 命名空间作用域资源  
✅ `listNamespaces`：列出核心 v1 命名空间（集群作用域）

### Apps API 组 (apps/v1)
✅ `listAppsResources`：列出 Apps v1 命名空间作用域资源（Deployment 等）  
✅ `getAppsResource`：获取 Apps v1 资源  
✅ `createAppsResource`：创建 Apps v1 资源  
✅ `updateAppsResource`：更新 Apps v1 资源  
✅ `deleteAppsResource`：删除 Apps v1 资源

### Batch API 组 (batch/v1)
❌ `listBatchResources`：列出 Batch v1 资源  
❌ `getBatchResource`：获取 Batch v1 资源  
❌ `createBatchResource`：创建 Batch v1 资源  
❌ `updateBatchResource`：更新 Batch v1 资源  
❌ `deleteBatchResource`：删除 Batch v1 资源

### Networking API 组 (networking.k8s.io/v1)
❌ `listNetworkingResources`：列出 Networking v1 资源  
❌ `getNetworkingResource`：获取 Networking v1 资源  
❌ `createNetworkingResource`：创建 Networking v1 资源  
❌ `updateNetworkingResource`：更新 Networking v1 资源  
❌ `deleteNetworkingResource`：删除 Networking v1 资源
