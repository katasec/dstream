# Agent Hooks

Safety hooks that prevent destructive actions. These scripts read JSON from stdin and output a JSON deny decision to block an action, or produce no output to allow.

**Requires**: `jq` (pre-installed on macOS; `apt install jq` / `brew install jq` otherwise).

## Available Hooks

| Hook | Event | What it prevents |
|------|-------|-----------------|
| `block-dangerous-commands.sh` | preToolUse | `rm -rf`, fork bombs, pipe-to-shell, force push, `git reset --hard` |
| `protect-secrets.sh` | preToolUse | Access to `.env`, `*.pem`, `*.key`, credentials, secrets files |

## Pre-Wired: GitHub Copilot

Hooks are pre-configured in [`.copilot/hooks.json`](../../.copilot/hooks.json) — no setup needed. Copilot's coding agent, CLI, and VS Code agent mode will run them automatically on every tool use.

## Wiring for Other Tools

The scripts are compatible with any tool that sends JSON on stdin and reads JSON deny decisions. They auto-detect both Copilot (`toolArgs`) and Claude Code (`tool_input`) field formats.

### Claude Code (`.claude/settings.json`)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{ "type": "command", "command": "bash \"$CLAUDE_PROJECT_DIR\"/.agents/hooks/block-dangerous-commands.sh" }]
      },
      {
        "matcher": "Read|Edit|Write",
        "hooks": [{ "type": "command", "command": "bash \"$CLAUDE_PROJECT_DIR\"/.agents/hooks/protect-secrets.sh" }]
      }
    ]
  }
}
```

### Contract

- **Input**: JSON on stdin with tool name and arguments
- **Output**: `{"permissionDecision": "deny", "permissionDecisionReason": "..."}` to block, or no output to allow