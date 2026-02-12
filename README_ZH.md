# Kubernetes MCP

<div align="center">
  <img src="logo.png" alt="Kubernetes MCP Logo" width="180">
</div>

[English](README.md) | 中文

一个面向 Kubernetes 的 Model Context Protocol (MCP) 服务端，使用 Go 实现。
它将 Kubernetes 操作能力暴露为 MCP 工具，支持 `stdio`、`sse`、`streamable` 三种传输方式。

## 项目价值

- 用统一 MCP 接口访问 Kubernetes 能力
- 同时支持 `kubeconfig` 与集群内配置
- 提供多种传输模式，方便对接不同客户端
- 覆盖通用资源 CRUD、工具类能力、提示词与指标查询
- 使用官方 `slog` 日志体系，支持级别与格式配置

## 环境要求

- Go `1.24+`
- 可访问的 Kubernetes 集群
  - 通过 `--kubeconfig` 指定，或
  - 运行在集群内并使用 ServiceAccount

## 快速开始

### 构建

```bash
git clone https://github.com/HSn0918/kubernetes-mcp.git
cd kubernetes-mcp
go build -o kubernetes-mcp ./cmd/kubernetes-mcp
```

### 启动（stdio）

```bash
./kubernetes-mcp server transport stdio --kubeconfig ~/.kube/config
```

### 启动（SSE）

```bash
./kubernetes-mcp server transport sse \
  --port 8080 \
  --health-port 8081 \
  --allow-origins "*" \
  --base-url "http://localhost:8080" \
  --kubeconfig ~/.kube/config
```

### 启动（StreamableHTTP）

```bash
./kubernetes-mcp server transport streamable \
  --port 8080 \
  --health-port 8081 \
  --allow-origins "*" \
  --kubeconfig ~/.kube/config
```

## 命令结构

```text
kubernetes-mcp
├── server
│   └── transport
│       ├── stdio
│       ├── sse
│       └── streamable
└── version
```

## 配置参数

全局参数：

- `--log-level`：`debug|info|warn|error`，默认 `info`
- `--log-format`：`console|json`，默认 `console`
- `--kubeconfig`：可选

网络传输参数（`sse`、`streamable`）：

- `--port`：默认 `8080`
- `--health-port`：默认 `8081`
- `--allow-origins`：默认 `*`

SSE 专属参数：

- `--base-url`：默认 `http://localhost:<port>`

## 传输端点

- StreamableHTTP MCP 端点：`POST /mcp`
- 健康检查端点（仅 SSE/Streamable）：
  - `GET /healthz`
  - `GET /readyz`

## 工具能力总览

### 1. 通用资源操作

每个 API 组前缀都会注册以下工具：

- `LIST_<PREFIX>_RESOURCES`
- `GET_<PREFIX>_RESOURCE`
- `DESCRIBE_<PREFIX>_RESOURCE`
- `CREATE_<PREFIX>_RESOURCE`
- `UPDATE_<PREFIX>_RESOURCE`
- `DELETE_<PREFIX>_RESOURCE`

当前前缀：

- `K8S`

### 2. Core/集群基础工具

- `LIST_NAMESPACES`
- `LIST_NODES`
- `GET_POD_LOGS`
- `ANALYZE_POD_LOGS`

### 3. Utility 工具

- `GET_CURRENT_TIME`
- `GET_CLUSTER_INFO`
- `GET_API_RESOURCES`
- `SEARCH_RESOURCES`
- `EXPLAIN_RESOURCE`
- `APPLY_MANIFEST`
- `VALIDATE_MANIFEST`
- `DIFF_MANIFEST`
- `GET_EVENTS`

### 4. Prompt 工具

- `KUBERNETES_YAML_PROMPT`
- `KUBERNETES_QUERY_PROMPT`
- `TROUBLESHOOT_PODS_PROMPT`
- `TROUBLESHOOT_NODES_PROMPT`
- `TROUBLESHOOT_NETWORK_PROMPT`

### 5. Metrics 工具

- `GET_NODE_METRICS`
- `GET_POD_METRICS`
- `GET_RESOURCE_METRICS`
- `GET_TOP_CONSUMERS`

## 架构目录（k8s 风格 group/version）

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

## Docker 使用

构建镜像：

```bash
docker build -t kubernetes-mcp:latest .
```

建议显式指定传输子命令启动：

```bash
docker run --rm -it \
  -v ~/.kube:/root/.kube \
  kubernetes-mcp:latest \
  server transport stdio --kubeconfig /root/.kube/config
```

## Kubernetes 部署

部署文件位于 `/deploy/kubernetes`。

```bash
make k8s-deploy
# 或
make k8s-deploy-kustomize
```

## 开发命令

```bash
make build
make test
```

## Justfile 与一键安装 Skill

仓库内已内置项目使用指南 skill：`skills/kubernetes-mcp-usage`。

```bash
# 查看 just 任务
just

# 常用开发任务
just build
just test

# 一键安装项目 skill 到 Codex
just skill-install

# 强制覆盖已安装 skill
just skill-install-force
```

skill 安装目录：
`$CODEX_HOME/skills/kubernetes-mcp-usage`（若未设置 `CODEX_HOME`，则为 `~/.codex/skills/kubernetes-mcp-usage`）。

安装后请重启 Codex 以加载新 skill。

## 说明

- 资源操作能力受集群 API 可用性与 RBAC 权限影响。
- 在反向代理/Ingress 场景下，请正确配置 SSE 的 `--base-url`。
