# gomux

Interactive tmux session picker. Designed as a mosh entry point for managing multiple named sessions from a phone.

## Usage

```
gomux              # interactive TUI — list sessions, pick or create
gomux <name>       # attach or create named session (bypasses TUI)
```

When no sessions exist, gomux skips the TUI and creates a "default" session automatically.

## Install

```bash
go install github.com/rlebel12/gomux/cmd/gomux@latest
```

Requires `tmux` on PATH.

## Termux setup

```bash
home() {
    mosh rob@home -- gomux "$@"
}
```

## TUI keys

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate sessions |
| `Enter` | Attach to selected session |
| `n` | Create new session (opens name input) |
| `q` | Quit |
| `Esc` | Cancel name input |

Session names cannot contain dots or colons. Entering an existing name attaches instead of creating.
