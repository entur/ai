#!/usr/bin/env bash
# Convenience wrapper for the MCP knowledge-base quality suite.
#
# Reuses the Go runner in ../ but injects an MCP-only system prompt and
# restricts the agent to MCP tools so failures attribute cleanly to the
# entur-kompass MCP + the markdown docs it indexes (not to local Read/Grep).
#
# Usage:
#   tests/mcp/run.sh [runner flags...]
#
# Examples:
#   tests/mcp/run.sh --dry-run
#   tests/mcp/run.sh --verbose
#   tests/mcp/run.sh --scenario "04-*" --verbose --model sonnet
#
# Pass any flag accepted by the parent runner (see tests/README.md).

set -euo pipefail

# Resolve script dir so the wrapper works from any CWD.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="$(dirname "$SCRIPT_DIR")"

SYSTEM_PROMPT='You are evaluating the entur/ai knowledge base via the entur-kompass MCP server.
You may ONLY use the mcp__entur-kompass__search_entur_kb tool. Do not call any other tool.
Do NOT use Read, Grep, Glob, WebFetch, WebSearch, or any prior knowledge of Entur conventions.
You MUST end your turn with a final assistant text message in the exact format the user requests.
Never end your turn on a tool call. If the MCP results do not contain the answer, write `answer: insufficient` -- do not fabricate.'

ALLOWED_TOOLS='mcp__entur-kompass__search_entur_kb'

cd "$TESTS_DIR"

exec go run . \
  --dir mcp/scenarios \
  --allowed-tools "$ALLOWED_TOOLS" \
  --system-prompt "$SYSTEM_PROMPT" \
  "$@"
