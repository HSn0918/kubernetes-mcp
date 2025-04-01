# Kubernetes MCP

<div align="center">
  <img src="logo.png" alt="Kubernetes MCP Logo" width="200">
</div>

English | [中文](README_ZH.md)

✨ A Model Context Protocol (MCP) server implementation designed with Go for interacting with Kubernetes clusters. This server allows MCP-compatible clients to perform Kubernetes operations through defined tools.

## 📌 Core Features

- 🔹 **MCP Server**: Implements the `mcp-go` library to provide MCP functionality
- 🔹 **Kubernetes Interaction**: Uses `controller-runtime` client to interact with clusters
- 🔹 **Transport Methods**: Supports standard I/O (`stdio`) or Server-Sent Events (`sse`)

## 🛠️ Resource Management Tools

### 📊 Implemented API Groups

🔸 **Core API Group (v1)**
- List, get, describe, create, update, delete operations
- Cluster-scoped: list namespaces, list nodes
- Get Pod logs functionality

🔸 **Apps API Group (apps/v1)**
- Full support for Deployment, ReplicaSet, StatefulSet, DaemonSet

🔸 **Batch API Group (batch/v1)**
- Full support for Job, CronJob

🔸 **Networking API Group (networking.k8s.io/v1)**
- Full support for Ingress, NetworkPolicy

🔸 **RBAC API Group (rbac.authorization.k8s.io/v1)**
- Full support for Role, RoleBinding, ClusterRole, ClusterRoleBinding

🔸 **Storage API Group (storage.k8s.io/v1)**
- Full support for StorageClass, VolumeAttachment

🔸 **Policy API Group (policy/v1beta1)**
- Full support for PodSecurityPolicy, PodDisruptionBudget

🔸 **API Extensions API Group (apiextensions.k8s.io/v1)**
- Full support for CustomResourceDefinition

🔸 **Autoscaling API Group (autoscaling/v1)**
- Full support for HorizontalPodAutoscaler

## 📋 Requirements

📌 **Go 1.24**
📌 **Kubernetes cluster access** (via `kubeconfig` or in-cluster service account)

## 📦 Key Dependencies

🧩 **Core Libraries**:
- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- `sigs.k8s.io/controller-runtime` - Kubernetes client
- `k8s.io/client-go` - Core Kubernetes libraries
- `github.com/spf13/cobra` - CLI structure
- `go.uber.org/zap` - Logging
- `sigs.k8s.io/yaml` - YAML processing

## 🔨 Build Methods

### 📥 Build from Source

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
./kubernetes-mcp server --transport=sse --port 8080
```

### 🐳 Docker Build

```bash
# Build image
docker build -t kubernetes-mcp:latest \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") .

# Run with stdio transport (default)
docker run -v ~/.kube:/root/.kube kubernetes-mcp:latest

# Run with SSE transport
docker run -p 8080:8080 -v ~/.kube:/root/.kube kubernetes-mcp:latest server --transport=sse

# View version info
docker run kubernetes-mcp:latest version

# Specify custom kubeconfig
docker run -v /path/to/config:/config kubernetes-mcp:latest server --kubeconfig=/config
```

## 🚀 Usage

### 🔄 Starting the Server

```shell
# Using standard I/O (default)
./kubernetes-mcp server

# Using SSE (Server-Sent Events)
./kubernetes-mcp server --transport sse --port 8080

# Specifying Kubeconfig
./kubernetes-mcp server --kubeconfig /path/to/your/kubeconfig

# View version
./kubernetes-mcp version
```

### ⚙️ Configuration Options

- 🔧 **Transport**: `--transport` (stdio/sse)
- 🔧 **Port**: `--port` (default 8080, SSE mode)
- 🔧 **Config file**: `--kubeconfig` (path)
- 🔧 **Log level**: `--log-level` (debug/info/warn/error)
- 🔧 **Log format**: `--log-format` (console/json)

## 🧩 Advanced Features

### 🔍 Structured Tools

- 🔍 **GET_CLUSTER_INFO**: Get cluster information and version details
- 🔍 **GET_API_RESOURCES**: List available API resources in the cluster
- 🔍 **SEARCH_RESOURCES**: Search across namespaces and resource types
- 🔍 **EXPLAIN_RESOURCE**: Get resource structure and field details
- 🔍 **APPLY_MANIFEST**: Apply YAML manifests to the cluster
- 🔍 **VALIDATE_MANIFEST**: Validate YAML manifest format
- 🔍 **DIFF_MANIFEST**: Compare YAML with existing cluster resources
- 🔍 **GET_EVENTS**: Get events related to specific resources

### 💡 Prompt System

- 🔖 **KUBERNETES_YAML_PROMPT**: Generate standard Kubernetes YAML
- 🔖 **KUBERNETES_QUERY_PROMPT**: Kubernetes operation guidance
- 🔖 **TROUBLESHOOT_PODS_PROMPT**: Pod troubleshooting guide
- 🔖 **TROUBLESHOOT_NODES_PROMPT**: Node troubleshooting guide

### 🔄 Standard Resource Operations

Each API group supports the following operations:
- **List resources**: Get resource lists, filterable by namespace and labels
- **Get resource**: Retrieve specific resources in YAML format
- **Describe resource**: Get detailed readable descriptions of resources
- **Create resource**: Create new resources from YAML
- **Update resource**: Update existing resources using YAML
- **Delete resource**: Remove specific resources

### 🌟 Core API Group Special Operations

- **Get Pod logs**: Retrieve logs from specific Pod containers
- **List namespaces**: View all available namespaces in the cluster
- **List nodes**: View all nodes and their status in the cluster

### 📊 Log Analysis Features

- **Error Pattern Recognition**: Identifies common error patterns and frequencies
- **Time-based Distribution Analysis**: Analyzes error occurrence patterns over time
- **HTTP Status Code Tracking**: Monitors and categorizes HTTP response codes
- **Performance Metrics**: Tracks response times and resource usage statistics
