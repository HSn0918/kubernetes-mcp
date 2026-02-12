# Kubernetes MCP

<div align="center">
  <img src="logo.png" alt="Kubernetes MCP Logo" width="180">
</div>

English | [中文](README_ZH.md)

A Kubernetes-focused Model Context Protocol (MCP) server built in Go.
It exposes Kubernetes operations as MCP tools and supports `stdio`, `sse`, and `streamable` transports.

## Why This Project

- Unified MCP interface for Kubernetes operations
- Works with both local `kubeconfig` and in-cluster configuration
- Multiple transports for different integration modes
- Generic resource CRUD + utility tools + prompts + metrics
- `slog`-based logging with configurable level/format

## Requirements

- Go `1.24+`
- Access to a Kubernetes cluster
  - via `--kubeconfig`, or
  - via in-cluster service account

## Quick Start

### Build

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
```

### Run (stdio)

```bash
./kubernetes-mcp server transport stdio --kubeconfig ~/.kube/config
```

### Run (SSE)

```bash
./kubernetes-mcp server transport sse \
  --port 8080 \
  --health-port 8081 \
  --allow-origins "*" \
  --base-url "http://localhost:8080" \
  --kubeconfig ~/.kube/config
```

### Run (StreamableHTTP)

```bash
./kubernetes-mcp server transport streamable \
  --port 8080 \
  --health-port 8081 \
  --allow-origins "*" \
  --kubeconfig ~/.kube/config
```

## CLI Structure

```text
kubernetes-mcp
├── server
│   └── transport
│       ├── stdio
│       ├── sse
│       └── streamable
└── version
```

## Configuration Flags

Global flags:

- `--log-level` (`debug|info|warn|error`, default `info`)
- `--log-format` (`console|json`, default `console`)
- `--kubeconfig` (optional)

Network transport flags (`sse`, `streamable`):

- `--port` (default `8080`)
- `--health-port` (default `8081`)
- `--allow-origins` (default `*`)

SSE-only flags:

- `--base-url` (default `http://localhost:<port>`)

## Transport Endpoints

- StreamableHTTP MCP endpoint: `POST /mcp`
- Health endpoints (SSE/Streamable only):
  - `GET /healthz`
  - `GET /readyz`

## Tooling Overview

### 1. Generic Resource Operations

For each API group prefix, the server exposes:

- `LIST_<PREFIX>_RESOURCES`
- `GET_<PREFIX>_RESOURCE`
- `DESCRIBE_<PREFIX>_RESOURCE`
- `CREATE_<PREFIX>_RESOURCE`
- `UPDATE_<PREFIX>_RESOURCE`
- `DELETE_<PREFIX>_RESOURCE`

Registered prefixes:

- `K8S`

### 2. Core/Cluster Utilities

- `LIST_NAMESPACES`
- `LIST_NODES`
- `GET_POD_LOGS`
- `ANALYZE_POD_LOGS`

### 3. Utility Tools

- `GET_CURRENT_TIME`
- `GET_CLUSTER_INFO`
- `GET_API_RESOURCES`
- `SEARCH_RESOURCES`
- `EXPLAIN_RESOURCE`
- `APPLY_MANIFEST`
- `VALIDATE_MANIFEST`
- `DIFF_MANIFEST`
- `GET_EVENTS`

### 4. Prompt Tools

- `KUBERNETES_YAML_PROMPT`
- `KUBERNETES_QUERY_PROMPT`
- `TROUBLESHOOT_PODS_PROMPT`
- `TROUBLESHOOT_NODES_PROMPT`
- `TROUBLESHOOT_NETWORK_PROMPT`

### 5. Metrics Tools

- `GET_NODE_METRICS`
- `GET_POD_METRICS`
- `GET_RESOURCE_METRICS`
- `GET_TOP_CONSUMERS`

## Architecture (k8s-style group/version)

```text
cmd/kubernetes-mcp/
  main.go
  app/
pkg/
  client/kubernetes/
  config/
  handlers/
    apis/
      core/v1/
      apps/v1/
    base/
    tool/
    prompt/
    metrics/
  server/
  health/
  logger/
```

## Docker

Build:

```bash
docker build -t kubernetes-mcp:latest .
```

Run with explicit transport command:

```bash
docker run --rm -it \
  -v ~/.kube:/root/.kube \
  kubernetes-mcp:latest \
  server transport stdio --kubeconfig /root/.kube/config
```

## Kubernetes Deployment

Deployment manifests are in `/deploy/kubernetes`.

```bash
make k8s-deploy
# or
make k8s-deploy-kustomize
```

## Development

```bash
make build
make test
```

## Justfile + One-Click Skill Install

This repo includes a project usage skill at `skills/kubernetes-mcp-usage`.

```bash
# list just tasks
just

# common dev tasks
just build
just test

# install the project skill into Codex
just skill-install

# overwrite existing install
just skill-install-force
```

Skill install destination:
`$CODEX_HOME/skills/kubernetes-mcp-usage` (or `~/.codex/skills/kubernetes-mcp-usage` when `CODEX_HOME` is not set).

After installing, restart Codex to pick up the skill.

## Notes

- Resource operations depend on your cluster API availability and RBAC permissions.
- For network transports behind ingress/reverse proxy, set `--base-url` correctly for SSE.
