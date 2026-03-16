# usg

Lightweight CLI to report Claude Code and Codex usage/limits.

## Install

```bash
go install github.com/sujshe/usg@latest
```

Or build from source:

```bash
make build
```

## Usage

```bash
usg claude          # Claude Code usage
usg codex           # Codex usage
usg all             # Both (default)
usg claude --json   # JSON output
usg all --watch 15  # Poll every 15s
```

## Auth

**Claude Code**: Reads OAuth token from macOS Keychain (`Claude Code-credentials`), falls back to `~/.claude/.credentials.json`.

**Codex**: Reads OAuth token from `~/.codex/auth.json`. Auto-refreshes if token is older than 8 days.

## Cross-compile

```bash
make all   # darwin-arm64 + linux-arm64 + linux-amd64
```

Binaries output to `dist/`.
