set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

default:
  @just --list

build:
  @make build

test:
  @make test

run-stdio:
  @make run-stdio

run-sse:
  @make run-sse

skill-install:
  @bash ./scripts/install-skill.sh

skill-install-force:
  @FORCE=1 bash ./scripts/install-skill.sh

skill-show:
  @echo "Skill source: ./skills/kubernetes-mcp-usage/SKILL.md"
  @echo "Install path: ${CODEX_HOME:-$HOME/.codex}/skills/kubernetes-mcp-usage"
