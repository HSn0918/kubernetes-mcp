# Kubernetes MCP

<div align="center">
  <img src="logo.png" alt="Kubernetes MCP Logo" width="200">
</div>

[English](README.md) | 中文

✨ 一个使用 Go 语言设计的 Model Context Protocol (MCP) 服务器实现，用于与 Kubernetes 集群交互。支持 MCP 兼容的客户端通过定义的工具执行 Kubernetes 操作。

## 📌 核心功能

- 🔹 **MCP 服务器**：实现 `mcp-go` 库提供 MCP 功能
- 🔹 **Kubernetes 交互**：使用 `controller-runtime` 客户端与集群交互
- 🔹 **传输方式**：支持标准 I/O（`stdio`）或服务器发送事件（`sse`）

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

- 🔧 **传输方式**：`--transport` (stdio/sse)
- 🔧 **端口**：`--port` (默认 8080，SSE 模式)
- 🔧 **配置文件**：`--kubeconfig` (路径)
- 🔧 **日志级别**：`--log-level` (debug/info/warn/error)
- 🔧 **日志格式**：`--log-format` (console/json)

## 🧩 高级功能

### 📝 结构化工具

- 🔍 **GET_CLUSTER_INFO**：获取集群信息与版本详情
- 🔍 **GET_API_RESOURCES**：列出集群可用 API 资源
- 🔍 **SEARCH_RESOURCES**：跨命名空间和资源类型搜索
- 🔍 **EXPLAIN_RESOURCE**：获取资源结构和字段详情
- 🔍 **APPLY_MANIFEST**：应用 YAML 清单到集群
- 🔍 **VALIDATE_MANIFEST**：验证 YAML 清单格式
- 🔍 **DIFF_MANIFEST**：比较 YAML 与集群现有资源
- 🔍 **GET_EVENTS**：获取特定资源相关事件

### 💡 提示词系统

- 🔖 **KUBERNETES_YAML_PROMPT**：生成标准 Kubernetes YAML
- 🔖 **KUBERNETES_QUERY_PROMPT**：Kubernetes 操作指导
- 🔖 **TROUBLESHOOT_PODS_PROMPT**：Pod 问题排查指南
- 🔖 **TROUBLESHOOT_NODES_PROMPT**：节点问题排查指南

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

### 📊 日志分析功能

- **错误模式识别**：识别常见错误模式及其频率
- **基于时间的分布分析**：分析错误发生的时间模式
- **HTTP状态码跟踪**：监控和分类HTTP响应代码
- **性能指标**：跟踪响应时间和资源使用统计

### 📊 集群资源指标功能

- 🔍 **GET_NODE_METRICS**：获取节点资源使用情况指标，包括CPU和内存占用量及使用率
- 🔍 **GET_POD_METRICS**：获取Pod资源使用情况指标，查看容器CPU和内存消耗
- 🔍 **GET_RESOURCE_METRICS**：获取集群整体资源使用情况，包括CPU、内存、存储和Pod数量统计
- 🔍 **GET_TOP_CONSUMERS**：识别资源消耗最高的Pod，帮助定位资源瓶颈

所有指标API均支持：
- 灵活排序：按CPU、内存使用量或使用率排序
- 详细过滤：通过字段选择器和标签选择器精确定位资源
- 结果限制：控制返回结果数量
- JSON格式：所有响应均以结构化JSON格式返回，便于进一步处理

### 📝 集群指标提示词系统

- 🔖 **CLUSTER_RESOURCE_USAGE**：获取集群资源使用情况的指导
- 🔖 **NODE_RESOURCE_USAGE**：获取节点资源使用情况的指导
- 🔖 **POD_RESOURCE_USAGE**：获取Pod资源使用情况的指导

### 📋 API响应格式化

所有API响应现已标准化为JSON格式：
- 🔸 **结构化响应**：所有API响应均以一致的JSON结构返回
- 🔸 **节点列表**：包含节点名称、状态、角色、标签、污点、可分配资源等详细信息
- 🔸 **命名空间列表**：包含命名空间名称、状态、标签、注释等详细信息
- 🔸 **日志与日志分析**：日志内容及分析结果以结构化格式返回，便于处理
- 🔸 **资源指标**：CPU、内存、存储等指标以结构化格式返回，包含原始数值及百分比
- 🔸 **时间格式化**：支持中英文双语的人类可读时间格式，如"5分钟前"/"5 minutes ago"
