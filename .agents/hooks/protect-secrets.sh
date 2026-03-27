#!/usr/bin/env bash
# Hook: protect-secrets (preToolUse)
# Blocks read/edit/write access to sensitive files.
#
# Compatible with GitHub Copilot (.copilot/hooks.json) and Claude Code (.claude/settings.json).
# Reads JSON from stdin. Outputs JSON deny decision or nothing (allow).

set -euo pipefail

INPUT=$(cat)

# Extract file path — handles both Copilot (toolArgs.path) and Claude Code (tool_input.file_path)
FILE_PATH=$(echo "$INPUT" | jq -r '(.toolArgs.path // .toolArgs.file_path // .tool_input.file_path // .tool_input.path // .tool_input.command // "")' 2>/dev/null || echo "")

# Sensitive file patterns
PROTECTED_PATTERNS=(
  '\.env$'
  '\.env\.'
  '\.pem$'
  '\.key$'
  '\.p12$'
  '\.pfx$'
  '\.jks$'
  'credentials'
  'secrets\.yaml'
  'secrets\.json'
  'secrets\.yml'
  '\.secret$'
  'id_rsa'
  'id_ed25519'
  '\.kube/config$'
  'kubeconfig'
  'token$'
  'password'
)

for pattern in "${PROTECTED_PATTERNS[@]}"; do
  if echo "$FILE_PATH" | grep -qiE "$pattern"; then
    jq -cn --arg reason "Blocked access to sensitive file matching: $pattern" \
      '{permissionDecision: "deny", permissionDecisionReason: $reason}'
    exit 0
  fi
done

# Allow (no output)
exit 0