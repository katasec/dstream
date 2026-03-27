#!/usr/bin/env bash
# Hook: block-dangerous-commands (preToolUse)
# Blocks destructive shell commands before they execute.
#
# Compatible with GitHub Copilot (.copilot/hooks.json) and Claude Code (.claude/settings.json).
# Reads JSON from stdin. Outputs JSON deny decision or nothing (allow).

set -euo pipefail

INPUT=$(cat)

# Extract command — handles both Copilot (toolArgs.command) and Claude Code (tool_input.command)
COMMAND=$(echo "$INPUT" | jq -r '(.toolArgs.command // .tool_input.command // "")' 2>/dev/null || echo "")

# Patterns that should never run unattended
DANGEROUS_PATTERNS=(
  'rm -rf /'
  'rm -rf ~'
  'rm -rf \.'
  'rm -rf \*'
  'mkfs\.'
  'dd if='
  ':(){ :|:& };:'        # fork bomb
  'chmod -R 777 /'
  'curl.*|.*sh'           # pipe-to-shell
  'wget.*|.*sh'
  'git reset --hard'
  'git push.*--force'
  'git push.*-f '
  'git clean -fdx'
  '> /dev/sda'
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qiE "$pattern"; then
    jq -cn --arg reason "Blocked dangerous command matching: $pattern" \
      '{permissionDecision: "deny", permissionDecisionReason: $reason}'
    exit 0
  fi
done

# Allow (no output)
exit 0