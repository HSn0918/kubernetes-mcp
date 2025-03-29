# Kubernetes-MCP

English | [中文](README_ZH.md)

A server implementation of the Model Capable Protocol (MCP) designed for interacting with Kubernetes clusters using Go. This server allows MCP-compatible clients to perform Kubernetes operations via defined tools.

## ✨ Features

* **MCP Server:** Implements the `mcp-go` library to provide MCP capabilities.
* **Kubernetes Interaction:** Uses the `controller-runtime` client to interact with a Kubernetes cluster.
* **Multiple Transports:** Supports communication via standard I/O (`stdio`) or Server-Sent Events (`sse`).
* **Resource Management Tools:** Exposes MCP tools for Kubernetes operations:
    * **Core API Group (v1):**
        * Fully implemented: List namespace-scoped resources (`listResources`), Get resource YAML (`getResource`), Create from YAML (`createResource`), Update from YAML (`updateResource`), Delete resource (`deleteResource`).
        * Fully implemented: List cluster-scoped Namespaces (`listNamespaces`).
    * **Apps API Group (apps/v1):**
        * Implemented: List namespace-scoped resources (`listAppsResources`).
        * Placeholders (Not Implemented): Get (`getAppsResource`), Create (`createAppsResource`), Update (`updateAppsResource`), Delete (`deleteAppsResource`).
    * **Batch API Group (batch/v1):**
        * Placeholders (Not Implemented): List (`listBatchResources`), Get (`getBatchResource`), Create (`createBatchResource`), Update (`updateBatchResource`), Delete (`deleteBatchResource`).
    * **Networking API Group (networking.k8s.io/v1):**
        * Placeholders (Not Implemented): List (`listNetworkingResources`), Get (`getNetworkingResource`), Create (`createNetworkingResource`), Update (`updateNetworkingResource`), Delete (`deleteNetworkingResource`).
* **Configuration:** Configurable via command-line flags (transport, port, kubeconfig, logging level/format).
* **Logging:** Uses `zap` for structured logging.
* **CLI:** Built with `cobra` framework.

## 📋 Prerequisites

* **Go 1.24**
* Access to a Kubernetes cluster (via a `kubeconfig` file or in-cluster service account).

## 📦 Major Dependencies

This project relies on several key Go modules:

* `github.com/mark3labs/mcp-go` (for MCP server/protocol)
* `sigs.k8s.io/controller-runtime` (for Kubernetes client interaction)
* `k8s.io/client-go` (core Kubernetes libraries)
* `github.com/spf13/cobra` (for CLI structure)
* `go.uber.org/zap` (for logging)
* `sigs.k8s.io/yaml` (for YAML handling)

*(Note: Specific versions are managed in `go.mod` which was not provided)*

## 🔨 Building

### From Source

To build the executable:

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
./kubernetes-mcp server --transport=sse --port 8080
```

### Using Docker

Build the Docker image:

```bash
docker build -t kubernetes-mcp:latest \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") .
```

Run using Docker:

```bash
# Using stdio transport (default)
docker run -v ~/.kube:/root/.kube kubernetes-mcp:latest

# Using SSE transport
docker run -p 8080:8080 -v ~/.kube:/root/.kube kubernetes-mcp:latest server --transport=sse

# Check version
docker run kubernetes-mcp:latest version

# Specify a custom kubeconfig
docker run -v /path/to/config:/config kubernetes-mcp:latest server --kubeconfig=/config
```

## 🚀 Usage

The server is started using the `server` subcommand.

### Using Standard I/O (stdio - default):
```shell
./kubernetes-mcp server
```

### Using Server-Sent Events (SSE):
```shell
./kubernetes-mcp server --transport sse --port 8080
```
(Listens on port 8080 by default when using SSE)

### Specifying Kubeconfig:

Use the `--kubeconfig` flag if your configuration file is not in a standard location:
```shell
./kubernetes-mcp server --kubeconfig /path/to/your/kubeconfig
```

### Checking Version:

Displays version information set at build time.
```shell
./kubernetes-mcp version
# Output example: Kubernetes-mcp version dev (commit: none, build date: unknown)
```

## ⚙️ Configuration Flags
- `--transport`: Communication mode. Options: stdio (default), sse.
- `--port`: Port number for SSE transport. Default: 8080.
- `--kubeconfig`: Path to the Kubernetes configuration file. (Defaults to standard discovery: KUBECONFIG env var or ~/.kube/config).
- `--log-level`: Logging verbosity. Options: debug, info (default), warn, error.
- `--log-format`: Log output format. Options: console (default), json.

## 🧩 Supported MCP Tools (Kubernetes Operations)

The following MCP tools are registered based on Kubernetes API groups and actions:

### Core API Group (v1)
✅ `listResources`: List Core v1 namespace-scoped resources (Pods, Services, etc.)  
✅ `getResource`: Get Core v1 namespace-scoped resource YAML  
✅ `createResource`: Create Core v1 namespace-scoped resource from YAML  
✅ `updateResource`: Update Core v1 namespace-scoped resource from YAML  
✅ `deleteResource`: Delete Core v1 namespace-scoped resource  
✅ `listNamespaces`: List Core v1 Namespaces (cluster-scoped)

### Apps API Group (apps/v1)
✅ `listAppsResources`: List Apps v1 namespace-scoped resources (Deployments, etc.)  
❌ `getAppsResource`: Get Apps v1 resource  
❌ `createAppsResource`: Create Apps v1 resource  
❌ `updateAppsResource`: Update Apps v1 resource  
❌ `deleteAppsResource`: Delete Apps v1 resource

### Batch API Group (batch/v1)
❌ `listBatchResources`: List Batch v1 resources  
❌ `getBatchResource`: Get Batch v1 resource  
❌ `createBatchResource`: Create Batch v1 resource  
❌ `updateBatchResource`: Update Batch v1 resource  
❌ `deleteBatchResource`: Delete Batch v1 resource

### Networking API Group (networking.k8s.io/v1)
❌ `listNetworkingResources`: List Networking v1 resources  
❌ `getNetworkingResource`: Get Networking v1 resource  
❌ `createNetworkingResource`: Create Networking v1 resource  
❌ `updateNetworkingResource`: Update Networking v1 resource  
❌ `deleteNetworkingResource`: Delete Networking v1 resource
