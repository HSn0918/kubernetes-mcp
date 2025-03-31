# Kubernetes-MCP

[English](README.md) | 中文

一个使用 Go 语言设计的 Model Capable Protocol (MCP) 服务器实现，用于与 Kubernetes 集群交互。该服务器允许 MCP 兼容的客户端通过定义的工具执行 Kubernetes 操作。

## ✨ 功能特点

* **MCP 服务器：** 实现 `mcp-go` 库以提供 MCP 功能。
* **Kubernetes 交互：** 使用 `controller-runtime` 客户端与 Kubernetes 集群交互。
* **多种传输方式：** 支持通过标准 I/O（`stdio`）或服务器发送事件（`sse`）进行通信。
* **资源管理工具：** 提供用于 Kubernetes 操作的 MCP 工具：
    * **核心 API 组 (v1)：**
        * 已完整实现：列出命名空间作用域资源（`listResources`），获取资源 YAML（`getResource`），资源详细描述（`describeResource`），从 YAML 创建（`createResource`），从 YAML 更新（`updateResource`），删除资源（`deleteResource`），获取 Pod 日志（`getPodLogs`）。
        * 已完整实现：列出集群作用域命名空间（`listNamespaces`），列出节点（`listNodes`）。
    * **Apps API 组 (apps/v1)：**
        * 已完整实现：列出（`listAppsResources`），获取（`getAppsResource`），详细描述（`describeAppsResource`），创建（`createAppsResource`），更新（`updateAppsResource`），删除（`deleteAppsResource`）。
    * **Batch API 组 (batch/v1)：**
        * 已完整实现：列出（`listBatchResources`），获取（`getBatchResource`），详细描述（`describeBatchResource`），创建（`createBatchResource`），更新（`updateBatchResource`），删除（`deleteBatchResource`）。
    * **Networking API 组 (networking.k8s.io/v1)：**
        * 已完整实现：列出（`listNetworkingResources`），获取（`getNetworkingResource`），详细描述（`describeNetworkingResource`），创建（`createNetworkingResource`），更新（`updateNetworkingResource`），删除（`deleteNetworkingResource`）。
    * **RBAC API 组 (rbac.authorization.k8s.io/v1)：**
        * 已完整实现：列出（`listRbacResources`），获取（`getRbacResource`），详细描述（`describeRbacResource`），创建（`createRbacResource`），更新（`updateRbacResource`），删除（`deleteRbacResource`）。
    * **Storage API 组 (storage.k8s.io/v1)：**
        * 已完整实现：列出（`listStorageResources`），获取（`getStorageResource`），详细描述（`describeStorageResource`），创建（`createStorageResource`），更新（`updateStorageResource`），删除（`deleteStorageResource`）。
    * **Policy API 组 (policy/v1beta1)：**
        * 已完整实现：列出（`listPolicyResources`），获取（`getPolicyResource`），详细描述（`describePolicyResource`），创建（`createPolicyResource`），更新（`updatePolicyResource`），删除（`deletePolicyResource`）。
    * **API Extensions API 组 (apiextensions.k8s.io/v1)：**
        * 已完整实现：列出（`listApiextensionsResources`），获取（`getApiextensionsResource`），详细描述（`describeApiextensionsResource`），创建（`createApiextensionsResource`），更新（`updateApiextensionsResource`），删除（`deleteApiextensionsResource`）。
    * **Autoscaling API 组 (autoscaling/v1)：**
        * 已完整实现：列出（`listAutoscalingResources`），获取（`getAutoscalingResource`），详细描述（`describeAutoscalingResource`），创建（`createAutoscalingResource`），更新（`updateAutoscalingResource`），删除（`deleteAutoscalingResource`）。
* **高级过滤功能：**
    * 所有列表操作都支持标签选择器（`labelSelector`）以按标签过滤资源。
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

## 🧩 主要功能和工具使用

### 使用标签选择器列出资源

所有的列表操作现在都支持使用标签选择器进行资源过滤：

```
# 基本格式
LIST_<API_GROUP>_RESOURCES kind=<资源类型> apiVersion=<API版本> [namespace=<命名空间>] [labelSelector=<选择器>]

# 示例：
# 列出 'default' 命名空间中带有 app=nginx 标签的所有 Deployment
LIST_APPS_RESOURCES kind=Deployment apiVersion=apps/v1 namespace=default labelSelector=app=nginx

# 列出 'kube-system' 命名空间中带有 tier=control-plane 标签的所有 Pod
LIST_CORE_RESOURCES kind=Pod apiVersion=v1 namespace=kube-system labelSelector=tier=control-plane

# 更复杂的选择器
LIST_CORE_RESOURCES kind=Pod apiVersion=v1 labelSelector=environment in (production,staging),tier=frontend
```

### 资源管理操作

每个 API 组都支持标准的资源操作：

- **列出资源：** 获取资源列表，可选择按命名空间和标签进行过滤
- **获取资源：** 以 YAML 格式检索特定资源
- **描述资源：** 获取资源的详细可读描述
- **创建资源：** 从 YAML 创建新资源
- **更新资源：** 使用 YAML 更新现有资源
- **删除资源：** 移除特定资源

### 核心 API 组特殊操作

- **获取 Pod 日志：** 检索特定 Pod 容器的日志
- **列出命名空间：** 查看集群中所有可用的命名空间
- **列出节点：** 查看集群中所有节点及其状态

## 🌟 支持的 API 组

以下所有 API 组均已完整实现，支持完整的 CRUD 操作：

- **核心 API 组 (v1)**：Pod、Service、ConfigMap、Secret 等
- **Apps API 组 (apps/v1)**：Deployment、ReplicaSet、StatefulSet、DaemonSet
- **Batch API 组 (batch/v1)**：Job、CronJob
- **Networking API 组 (networking.k8s.io/v1)**：Ingress、NetworkPolicy
- **RBAC API 组 (rbac.authorization.k8s.io/v1)**：Role、RoleBinding、ClusterRole、ClusterRoleBinding
- **Storage API 组 (storage.k8s.io/v1)**：StorageClass、VolumeAttachment
- **Policy API 组 (policy/v1beta1)**：PodSecurityPolicy、PodDisruptionBudget
- **API Extensions API 组 (apiextensions.k8s.io/v1)**：CustomResourceDefinition
- **Autoscaling API 组 (autoscaling/v1)**：HorizontalPodAutoscaler
