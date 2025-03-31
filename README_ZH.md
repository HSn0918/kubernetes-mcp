# Kubernetes MCP

<div align="center">
  <img src="logo.png" alt="Kubernetes MCP Logo" width="200">
</div>

[English](README.md) | 中文

✨ 一个使用 Go 语言设计的 Model Capable Protocol (MCP) 服务器实现，用于与 Kubernetes 集群交互。支持 MCP 兼容的客户端通过定义的工具执行 Kubernetes 操作。

## 📌 核心功能

🔹 **MCP 服务器**：实现 `mcp-go` 库提供 MCP 功能
🔹 **Kubernetes 交互**：使用 `controller-runtime` 客户端与集群交互
🔹 **传输方式**：支持标准 I/O（`stdio`）或服务器发送事件（`sse`）

## 🛠️ 资源管理工具

### 📊 已实现的 API 组

🔸 **核心 API 组 (v1)**
- 列出资源、获取资源、详细描述、创建、更新、删除等操作
- 集群作用域：列出命名空间、列出节点
- 获取 Pod 日志功能

🔸 **应用 API 组 (apps/v1)**
- Deployment、ReplicaSet、StatefulSet、DaemonSet 完整支持

🔸 **批处理 API 组 (batch/v1)**
- Job、CronJob 完整支持

🔸 **网络 API 组 (networking.k8s.io/v1)**
- Ingress、NetworkPolicy 完整支持

🔸 **RBAC API 组 (rbac.authorization.k8s.io/v1)**
- Role、RoleBinding、ClusterRole、ClusterRoleBinding 完整支持

🔸 **存储 API 组 (storage.k8s.io/v1)**
- StorageClass、VolumeAttachment 完整支持

🔸 **策略 API 组 (policy/v1beta1)**
- PodSecurityPolicy、PodDisruptionBudget 完整支持

🔸 **API 扩展 API 组 (apiextensions.k8s.io/v1)**
- CustomResourceDefinition 完整支持

🔸 **自动扩缩容 API 组 (autoscaling/v1)**
- HorizontalPodAutoscaler 完整支持

## 📋 使用要求

📌 **Go 1.24**
📌 **Kubernetes 集群访问**（通过 `kubeconfig` 或集群内服务账户）

## 📦 主要依赖

🧩 **核心库**：
- `github.com/mark3labs/mcp-go` - MCP 协议实现
- `sigs.k8s.io/controller-runtime` - Kubernetes 客户端
- `k8s.io/client-go` - 核心 Kubernetes 库
- `github.com/spf13/cobra` - CLI 结构
- `go.uber.org/zap` - 日志记录
- `sigs.k8s.io/yaml` - YAML 处理

## 🔨 构建方法

### 📥 源代码构建

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
./kubernetes-mcp server --transport=sse --port 8080
```

### 🐳 Docker 构建

```bash
# 构建镜像
docker build -t kubernetes-mcp:latest \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") .

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

### 🔄 启动服务器

```shell
# 使用标准 I/O（默认）
./kubernetes-mcp server

# 使用 SSE（服务器发送事件）
./kubernetes-mcp server --transport sse --port 8080

# 指定 Kubeconfig
./kubernetes-mcp server --kubeconfig /path/to/your/kubeconfig

# 查看版本
./kubernetes-mcp version
```

### ⚙️ 配置选项

🔧 **传输方式**：`--transport` (stdio/sse)
🔧 **端口**：`--port` (默认 8080，SSE 模式)
🔧 **配置文件**：`--kubeconfig` (路径)
🔧 **日志级别**：`--log-level` (debug/info/warn/error)
🔧 **日志格式**：`--log-format` (console/json)

## 🧩 高级功能

### 📝 结构化工具

🔍 **GET_CLUSTER_INFO**：获取集群信息与版本详情
🔍 **GET_API_RESOURCES**：列出集群可用 API 资源
🔍 **SEARCH_RESOURCES**：跨命名空间和资源类型搜索
🔍 **EXPLAIN_RESOURCE**：获取资源结构和字段详情
🔍 **APPLY_MANIFEST**：应用 YAML 清单到集群
🔍 **VALIDATE_MANIFEST**：验证 YAML 清单格式
🔍 **DIFF_MANIFEST**：比较 YAML 与集群现有资源
🔍 **GET_EVENTS**：获取特定资源相关事件

### 💡 提示词系统

🔖 **KUBERNETES_YAML_PROMPT**：生成标准 Kubernetes YAML
🔖 **KUBERNETES_QUERY_PROMPT**：Kubernetes 操作指导
🔖 **TROUBLESHOOT_PODS_PROMPT**：Pod 问题排查指南
🔖 **TROUBLESHOOT_NODES_PROMPT**：节点问题排查指南

### 🔄 标准资源操作

每个 API 组支持以下操作：
- **列出资源**：获取资源列表，支持按命名空间和标签过滤
- **获取资源**：以 YAML 格式检索特定资源
- **描述资源**：获取资源详细可读描述
- **创建资源**：从 YAML 创建新资源
- **更新资源**：使用 YAML 更新现有资源
- **删除资源**：移除特定资源

### 🌟 核心 API 组特殊操作

- **获取 Pod 日志**：检索特定 Pod 容器的日志
- **列出命名空间**：查看集群中所有可用命名空间
- **列出节点**：查看集群中所有节点及其状态
