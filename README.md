# Kubernetes-MCP

English | [中文](README_ZH.md)

A server implementation of the Model Capable Protocol (MCP) designed for interacting with Kubernetes clusters using Go. This server allows MCP-compatible clients to perform Kubernetes operations via defined tools.

## ✨ Features

* **MCP Server:** Implements the `mcp-go` library to provide MCP capabilities.
* **Kubernetes Interaction:** Uses the `controller-runtime` client to interact with a Kubernetes cluster.
* **Multiple Transports:** Supports communication via standard I/O (`stdio`) or Server-Sent Events (`sse`).
* **Resource Management Tools:** Exposes MCP tools for Kubernetes operations:
    * **Core API Group (v1):**
        * Fully implemented: List namespace-scoped resources (`listResources`), Get resource YAML (`getResource`), Describe resource details (`describeResource`), Create from YAML (`createResource`), Update from YAML (`updateResource`), Delete resource (`deleteResource`), Get Pod Logs (`getPodLogs`).
        * Fully implemented: List cluster-scoped Namespaces (`listNamespaces`), List Nodes (`listNodes`).
    * **Apps API Group (apps/v1):**
        * Fully implemented: List (`listAppsResources`), Get (`getAppsResource`), Describe (`describeAppsResource`), Create (`createAppsResource`), Update (`updateAppsResource`), Delete (`deleteAppsResource`).
    * **Batch API Group (batch/v1):**
        * Fully implemented: List (`listBatchResources`), Get (`getBatchResource`), Describe (`describeBatchResource`), Create (`createBatchResource`), Update (`updateBatchResource`), Delete (`deleteBatchResource`).
    * **Networking API Group (networking.k8s.io/v1):**
        * Fully implemented: List (`listNetworkingResources`), Get (`getNetworkingResource`), Describe (`describeNetworkingResource`), Create (`createNetworkingResource`), Update (`updateNetworkingResource`), Delete (`deleteNetworkingResource`).
    * **RBAC API Group (rbac.authorization.k8s.io/v1):**
        * Fully implemented: List (`listRbacResources`), Get (`getRbacResource`), Describe (`describeRbacResource`), Create (`createRbacResource`), Update (`updateRbacResource`), Delete (`deleteRbacResource`).
    * **Storage API Group (storage.k8s.io/v1):**
        * Fully implemented: List (`listStorageResources`), Get (`getStorageResource`), Describe (`describeStorageResource`), Create (`createStorageResource`), Update (`updateStorageResource`), Delete (`deleteStorageResource`).
    * **Policy API Group (policy/v1beta1):**
        * Fully implemented: List (`listPolicyResources`), Get (`getPolicyResource`), Describe (`describePolicyResource`), Create (`createPolicyResource`), Update (`updatePolicyResource`), Delete (`deletePolicyResource`).
    * **API Extensions API Group (apiextensions.k8s.io/v1):**
        * Fully implemented: List (`listApiextensionsResources`), Get (`getApiextensionsResource`), Describe (`describeApiextensionsResource`), Create (`createApiextensionsResource`), Update (`updateApiextensionsResource`), Delete (`deleteApiextensionsResource`).
    * **Autoscaling API Group (autoscaling/v1):**
        * Fully implemented: List (`listAutoscalingResources`), Get (`getAutoscalingResource`), Describe (`describeAutoscalingResource`), Create (`createAutoscalingResource`), Update (`updateAutoscalingResource`), Delete (`deleteAutoscalingResource`).
* **Advanced Filtering:**
    * Label selector support (`labelSelector`) for all List operations to filter resources by labels.
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

## 🧩 Key Features and Tool Usage

### Resource Listing with Label Selectors

All list operations now support filtering resources using label selectors:

```
# Basic format
LIST_<API_GROUP>_RESOURCES kind=<kind> apiVersion=<apiVersion> [namespace=<namespace>] [labelSelector=<selector>]

# Examples:
# List all Deployments in the 'default' namespace with app=nginx label
LIST_APPS_RESOURCES kind=Deployment apiVersion=apps/v1 namespace=default labelSelector=app=nginx

# List all Pods in the 'kube-system' namespace with tier=control-plane label
LIST_CORE_RESOURCES kind=Pod apiVersion=v1 namespace=kube-system labelSelector=tier=control-plane

# More complex selectors
LIST_CORE_RESOURCES kind=Pod apiVersion=v1 labelSelector=environment in (production,staging),tier=frontend
```

### Resource Management Operations

Each API group supports standard resource operations:

- **List Resources:** Get a list of resources with optional filtering by namespace and labels
- **Get Resource:** Retrieve a specific resource in YAML format
- **Describe Resource:** Get a detailed human-readable description of a resource
- **Create Resource:** Create a new resource from YAML
- **Update Resource:** Update an existing resource using YAML
- **Delete Resource:** Remove a specific resource

### Core API Group Special Operations

- **Get Pod Logs:** Retrieve logs from a specific Pod container
- **List Namespaces:** View all available namespaces in the cluster
- **List Nodes:** View all nodes in the cluster with their status

## 🌟 Supported API Groups

All the following API groups are fully implemented with complete CRUD operations:

- **Core API Group (v1)**: Pods, Services, ConfigMaps, Secrets, etc.
- **Apps API Group (apps/v1)**: Deployments, ReplicaSets, StatefulSets, DaemonSets
- **Batch API Group (batch/v1)**: Jobs, CronJobs
- **Networking API Group (networking.k8s.io/v1)**: Ingress, NetworkPolicies
- **RBAC API Group (rbac.authorization.k8s.io/v1)**: Roles, RoleBindings, ClusterRoles, ClusterRoleBindings
- **Storage API Group (storage.k8s.io/v1)**: StorageClasses, VolumeAttachments
- **Policy API Group (policy/v1beta1)**: PodSecurityPolicies, PodDisruptionBudgets
- **API Extensions API Group (apiextensions.k8s.io/v1)**: CustomResourceDefinitions
- **Autoscaling API Group (autoscaling/v1)**: HorizontalPodAutoscalers
