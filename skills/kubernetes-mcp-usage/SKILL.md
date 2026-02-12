---
name: kubernetes-mcp-usage
description: "Usage guide skill for kubernetes-mcp MCP server. Use when connecting the MCP server to clients, troubleshooting tool calls, and selecting the right Kubernetes tools for diagnosis and operations."
---

# Kubernetes MCP Usage Guide

## When to Use
Use this skill when the user wants to operate Kubernetes through `kubernetes-mcp`:
- connect MCP client to this server (`stdio`, `sse`, `streamable`);
- discover resources, inspect cluster state, check logs/events/metrics;
- troubleshoot failed tool calls and parameter mismatch.

## Connection Checklist
1. Confirm server is reachable:
   - `stdio`: binary command works with local kubeconfig.
   - `streamable`: endpoint is `POST /mcp`.
2. Confirm Kubernetes auth is valid:
   - explicit `--kubeconfig` or in-cluster service account.
3. Prefer smallest-scope query first:
   - namespace + kind filters before wide cluster scans.

## Recommended Tool Flow
1. Health and inventory:
   - `GET_CLUSTER_INFO`
   - `LIST_NAMESPACES`
2. Resource location:
   - `SEARCH_RESOURCES` (supports `name=...`, `label=...`, `annotation=...`, wildcard `*`)
   - `LIST_K8S_RESOURCES`
3. Deep inspection:
   - `GET_K8S_RESOURCE` / `DESCRIBE_K8S_RESOURCE`
   - `GET_EVENTS`
4. Runtime diagnosis:
   - `GET_POD_LOGS`
   - `GET_POD_METRICS`, `GET_NODE_METRICS`, `GET_TOP_CONSUMERS`

## Practical Query Patterns
- Find all pods in namespace:
  - `SEARCH_RESOURCES {"kinds":"Pod","namespaces":"mcp-system","query":"name=*"}`
- Find by label wildcard:
  - `SEARCH_RESOURCES {"query":"label=app:nginx-*","namespaces":"default"}`
- Find by annotation:
  - `SEARCH_RESOURCES {"query":"annotation=team:platform","namespaces":"default"}`
- List known kind explicitly:
  - `LIST_K8S_RESOURCES {"kind":"Pod","apiVersion":"v1","namespace":"default"}`
- List namespace resources by discovery:
  - `LIST_K8S_RESOURCES {"namespace":"default"}`

## Error Handling Guidance
- If listing fails with missing `apiVersion`:
  - provide both `kind` and `apiVersion`, or omit `kind` for discovery listing.
- If search returns empty unexpectedly:
  - check `kinds` value (`Pod`/`pods` both accepted),
  - verify namespace scope,
  - test with `query:"name=*"` first.
- If metrics tools fail:
  - confirm metrics-server is installed and accessible.

## Safety
- Avoid destructive actions (`DELETE`, large apply) unless user explicitly requests.
- For changes, prefer `VALIDATE_MANIFEST` and `DIFF_MANIFEST` before `APPLY_MANIFEST`.
