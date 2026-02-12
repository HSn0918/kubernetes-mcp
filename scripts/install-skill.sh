#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
SKILL_NAME="${SKILL_NAME:-kubernetes-mcp-usage}"
SRC_DIR="${ROOT_DIR}/skills/${SKILL_NAME}"

CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"
DEST_BASE="${CODEX_HOME_DIR}/skills"
DEST_DIR="${DEST_BASE}/${SKILL_NAME}"

FORCE_INSTALL="${FORCE:-0}"

if [[ ! -f "${SRC_DIR}/SKILL.md" ]]; then
  echo "Error: skill source not found: ${SRC_DIR}/SKILL.md" >&2
  exit 1
fi

if [[ -e "${DEST_DIR}" && "${FORCE_INSTALL}" != "1" ]]; then
  echo "Error: ${DEST_DIR} already exists." >&2
  echo "Use FORCE=1 to overwrite, for example:" >&2
  echo "  FORCE=1 bash ./scripts/install-skill.sh" >&2
  exit 1
fi

mkdir -p "${DEST_BASE}"

if [[ -e "${DEST_DIR}" ]]; then
  rm -rf "${DEST_DIR}"
fi

cp -R "${SRC_DIR}" "${DEST_DIR}"

echo "Installed skill: ${SKILL_NAME}"
echo "Destination: ${DEST_DIR}"
echo "Restart Codex to pick up new skills."
