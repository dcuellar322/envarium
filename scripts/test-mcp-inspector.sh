#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BINARY=${ENVARIUM_BINARY:-"$ROOT/bin/envarium"}
INSPECTOR_VERSION=${MCP_INSPECTOR_VERSION:-0.21.2}
DATA_DIR=$(mktemp -d "${TMPDIR:-/tmp}/envarium-mcp-inspector.XXXXXX")
DAEMON_LOG="$DATA_DIR/daemon.log"
IPC_ADDRESS="$DATA_DIR/envarium.sock"

cleanup() {
  if [ "${DAEMON_PID:-}" != "" ]; then
    kill "$DAEMON_PID" 2>/dev/null || true
    wait "$DAEMON_PID" 2>/dev/null || true
  fi
  rm -rf "$DATA_DIR"
}
trap cleanup EXIT INT TERM

"$BINARY" daemon --data-dir "$DATA_DIR" --ipc-address "$IPC_ADDRESS" --address 127.0.0.1:0 >"$DAEMON_LOG" 2>&1 &
DAEMON_PID=$!

attempt=0
while [ ! -S "$IPC_ADDRESS" ]; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 100 ]; then
    printf 'Envarium daemon did not create its IPC socket\n' >&2
    sed -n '1,120p' "$DAEMON_LOG" >&2
    exit 1
  fi
  sleep 0.05
done

inspect() {
  npx --yes "@modelcontextprotocol/inspector@$INSPECTOR_VERSION" --cli \
    "$BINARY" --data-dir "$DATA_DIR" --ipc-address "$IPC_ADDRESS" mcp serve --transport stdio \
    --provider inspector --agent-id smoke --profile observe --method "$1"
}

TOOLS_FILE="$DATA_DIR/tools.json"
inspect tools/list >"$TOOLS_FILE"
grep -Fq 'envarium_system_info' "$TOOLS_FILE"
if grep -Fq 'envarium_project_teardown' "$TOOLS_FILE"; then
  printf 'Observe profile unexpectedly exposed destructive teardown\n' >&2
  exit 1
fi

RESOURCES_FILE="$DATA_DIR/resources.json"
inspect resources/list >"$RESOURCES_FILE"
grep -Fq 'envarium://system' "$RESOURCES_FILE"

PROMPTS_FILE="$DATA_DIR/prompts.json"
inspect prompts/list >"$PROMPTS_FILE"
grep -Fq 'envarium_start_and_verify' "$PROMPTS_FILE"

printf 'MCP Inspector smoke passed (tools, resources, prompts; observe profile)\n'
