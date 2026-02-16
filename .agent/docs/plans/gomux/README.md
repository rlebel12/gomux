# Design Summary: gomux

**Status:** Pending Review
**Specification:** [gomux-spec.md](../../specs/gomux-spec.md)

## Problem Statement

Connecting to a home machine from a phone (Termux) via mosh+tmux currently supports only a single persistent session. Managing multiple named sessions requires remembering session names and manually crafting tmux commands. There's no way to discover existing sessions interactively.

## Proposed Solution

A Go terminal UI application (`gomux`) that runs on the remote machine as the mosh entry point. It presents an interactive session picker using Bubbletea, allowing the user to list existing tmux sessions, create new named sessions, and attach — all from a polished TUI. It also supports a direct-attach CLI bypass for known session names.

**Invocation from Termux:**
```bash
home() {
    mosh rob@home -- gomux "$@"
}
```

**Modes:**
- `gomux` — interactive TUI showing existing sessions
- `gomux <name>` — directly attach/create the named session (bypasses TUI)

After selection, gomux calls `syscall.Exec` to replace itself with tmux.

## Referenced Patterns

| Pattern | Source | Application |
|---------|--------|-------------|
| Dependency Injection | `~/.agent/docs/go/references/dependency-injection.md` | Interfaces defined in consuming package (`internal/gomux`), injected via constructor |
| Error Handling | `~/.agent/docs/go/references/error-handling.md` | Domain errors for tmux failures (not found, binary missing), `fmt.Errorf` wrapping for command errors |
| Context Lifecycle | `~/.agent/docs/go/references/context-lifecycle.md` | Root context with signal cancellation for clean TUI shutdown |

## Design Decisions

### Decision 1: TUI Framework — Bubbletea

**Context:** Need an interactive terminal UI for session selection.

**Options Considered:**
- **Bubbletea**: Elm-architecture TUI framework from Charm. Rich ecosystem (lipgloss for styling, bubbles for reusable components). De facto standard for Go TUIs.
- **tview**: Widget-based approach, more batteries-included but less flexible for custom layouts.
- **Raw terminal**: No framework, use ANSI escape codes directly. Maximum control but high effort.

**Decision:** Bubbletea

**Rationale:** Best balance of flexibility and ecosystem. Lipgloss styling makes the UI polished with minimal code. The Elm architecture (Model → Update → View) is clean and testable.

**Tradeoffs:** Heavier dependency than raw terminal, but the ergonomics are worth it for a TUI tool.

### Decision 2: Process Replacement via syscall.Exec

**Context:** After the user selects a session, gomux needs to hand off to tmux.

**Options Considered:**
- **syscall.Exec**: Replace the gomux process entirely with tmux. Clean, no parent process lingering.
- **os/exec.Command + Run**: Spawn tmux as a child process. Simpler API but leaves gomux as a parent, complicating signal handling.

**Decision:** `syscall.Exec`

**Rationale:** The gomux process has no further purpose after selection. Replacing it with tmux is the cleanest handoff — identical to how a shell `exec` works. No zombie processes, no signal forwarding needed.

**Tradeoffs:** Platform-specific (Unix only), but this tool only targets Linux.

### Decision 3: Tmux CLI Wrapping via Structured Format Strings

**Context:** Need to parse tmux session information reliably.

**Options Considered:**
- **Default output parsing**: Parse human-readable `tmux list-sessions` output with regex.
- **Format strings**: Use `tmux list-sessions -F` with structured format variables (`#{session_name}`, `#{session_windows}`, etc.) and a delimiter.

**Decision:** Format strings with tab delimiter

**Rationale:** Format strings produce machine-parseable output that won't break when tmux changes its human-readable format. Delimiter-based splitting is simpler and more robust than regex.

### Decision 4: CLI Argument Handling — os.Args

**Context:** The CLI has exactly two modes: no args (TUI) and one arg (session name).

**Options Considered:**
- **os.Args**: Direct argument access. Zero dependencies.
- **flag stdlib**: Standard flag parsing. Overkill for positional args.
- **cobra**: Full CLI framework. Massive overkill.

**Decision:** `os.Args`

**Rationale:** The interface is `gomux [session-name]`. No flags, no subcommands. Raw `os.Args` is the simplest correct solution.

### Decision 5: Unified Tmux Command for Attach/Create

**Context:** When the user selects a session (existing or new), we need to either attach or create.

**Decision:** Always use `tmux new-session -A -s <name>`. The `-A` flag makes tmux attach if the session exists, or create it if it doesn't. This eliminates the need to differentiate between "attach" and "create" in the exec path.

## Component Scope

**In Scope:**
- Interactive TUI listing existing tmux sessions (name, window count with singular/plural, attached status, relative creation time)
- Create new session with user-provided name
- Attach to existing session
- Direct-attach CLI bypass (`gomux <name>`)
- Clean process handoff to tmux via `syscall.Exec`
- Empty state auto-create: skip TUI and exec into "default" when zero sessions exist
- Inline name validation: block tmux-invalid characters (dots, colons), warn on duplicate names

**Out of Scope:**
- Mosh/SSH connection management (handled by Termux shell function)
- Session killing, renaming, or detailed inspection
- Live session content preview
- Configuration file or persistent settings
- Client-side component (this is remote-only)

## Interface Definitions

### `internal/gomux` — domain types, interfaces, and orchestration

The core domain package. Owns the `Session` domain type, the interfaces (consumed here, implemented by `tmux` and `tui`), the `App` orchestrator, and shared sentinel errors. Has no local imports — all dependencies point inward toward this package.

**Dependency graph** (no cycles, domain has zero local deps):
```
cmd/gomux/main.go → internal/gomux, internal/tmux, internal/tui
internal/tmux     → internal/gomux (for Session type)
internal/tui      → internal/gomux (for Session type, ErrNoSelection)
internal/gomux    → (no local imports)
```

```go
package gomux

// Session represents a tmux session's metadata.
type Session struct {
    Name     string
    Windows  int
    Attached bool
    Created  time.Time
}

// DefaultSessionName is used when no sessions exist and the TUI is skipped.
const DefaultSessionName = "default"

// ErrNoSelection indicates the user quit without selecting a session.
var ErrNoSelection = errors.New("no session selected")

// SessionLister retrieves existing tmux sessions.
type SessionLister interface {
    ListSessions(ctx context.Context) ([]Session, error)
}

// SessionPicker presents sessions to the user and returns the chosen session name.
// Returns ErrNoSelection if the user quits without selecting.
type SessionPicker interface {
    Pick(ctx context.Context, sessions []Session) (string, error)
}

// SessionExecutor replaces the current process with a tmux session.
// Does not return on success (process is replaced).
type SessionExecutor interface {
    ExecIntoSession(ctx context.Context, sessionName string) error
}

// App orchestrates session listing, picking, and execution.
type App struct {
    lister   SessionLister
    picker   SessionPicker
    executor SessionExecutor
}

// New creates an App with the given dependencies.
func New(lister SessionLister, picker SessionPicker, executor SessionExecutor) *App

// Run executes the main logic.
//
// Mode routing:
//   - sessionName non-empty → bypass picker, exec directly into named session
//   - sessionName empty, zero sessions → exec into DefaultSessionName (skip TUI)
//   - sessionName empty, 1+ sessions → list sessions → present picker → exec into selection
//
// Returns nil if the user quits without selecting (ErrNoSelection is handled internally).
func (a *App) Run(ctx context.Context, sessionName string) error
```

### `internal/tmux` — concrete client

```go
package tmux

// Client interacts with the tmux CLI.
// Satisfies gomux.SessionLister and gomux.SessionExecutor.
type Client struct {
    tmuxPath string // absolute path resolved at construction via exec.LookPath
}

// NewClient creates a Client. Returns an error if tmux is not found on PATH.
// Resolves the tmux binary path eagerly so ExecIntoSession doesn't need to.
func NewClient() (*Client, error)

// ListSessions runs `tmux list-sessions -F` and parses the output.
// Returns an empty slice (not an error) if the tmux server isn't running.
func (c *Client) ListSessions(ctx context.Context) ([]gomux.Session, error)

// ExecIntoSession calls syscall.Exec to replace the current process with
// `tmux new-session -A -s <name>`. Does not return on success.
func (c *Client) ExecIntoSession(ctx context.Context, sessionName string) error

// parseSessions parses `tmux list-sessions -F` output into Session structs.
// Unexported — tested directly via unit tests on sample output.
func parseSessions(output string) ([]gomux.Session, error)
```

**Format string used by ListSessions:**
```
tmux list-sessions -F '#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}'
```

`session_created` is a Unix timestamp. Parsed into `time.Time` by `parseSessions`.

### `internal/tui` — picker

```go
package tui

// Picker displays an interactive session list using Bubbletea.
// Satisfies gomux.SessionPicker.
type Picker struct{}

func NewPicker() *Picker

// Pick displays the interactive session picker and returns the chosen session name.
// Returns gomux.ErrNoSelection if the user quits (q/ctrl+c) without selecting.
func (p *Picker) Pick(ctx context.Context, sessions []gomux.Session) (string, error)
```

**Internal TUI state machine** (not part of public API, described for implementation guidance):
- **List view** (default): Navigate sessions with j/k or arrows, Enter to attach, `n` to switch to input view, `q` to quit. Two-line-per-session layout: name on first line, metadata on second (window count with singular/plural, "attached" badge when applicable, relative creation time).
- **Input view**: Text input for new session name. Enter to confirm (returns name), Esc to return to list view. Inline validation: names with dots/colons show error and block submission; names matching an existing session show a warning that attachment will occur instead of creation. The picker has the full `[]Session` list to validate against.

### `cmd/gomux/main.go` — wiring only

Thin entrypoint. Creates concrete dependencies, builds the App, and calls Run.

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    client, err := tmux.NewClient() // fail fast if tmux not on PATH
    // handle err → print message, os.Exit(1)

    var sessionName string
    if len(os.Args) > 1 {
        sessionName = os.Args[1]
    }

    a := gomux.New(client, tui.NewPicker(), client)

    if err := a.Run(ctx, sessionName); err != nil {
        fmt.Fprintf(os.Stderr, "gomux: %v\n", err)
        os.Exit(1)
    }
}
```

### Test fakes (in `internal/gomux/gomux_test.go`)

```go
package gomux_test

type fakeLister struct {
    sessions []gomux.Session
    err      error
}

func (f *fakeLister) ListSessions(ctx context.Context) ([]gomux.Session, error) {
    return f.sessions, f.err
}

type fakePicker struct {
    name     string
    err      error
    received []gomux.Session // captures what sessions were passed
}

func (f *fakePicker) Pick(ctx context.Context, sessions []gomux.Session) (string, error) {
    f.received = sessions
    return f.name, f.err
}

type fakeExecutor struct {
    called      bool
    sessionName string
    err         error
}

func (f *fakeExecutor) ExecIntoSession(ctx context.Context, name string) error {
    f.called = true
    f.sessionName = name
    return f.err
}
```

## Dependencies

**External Dependencies:**
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/charmbracelet/bubbles` — Reusable TUI components (textinput for session name)

**System Dependencies:**
- `tmux` — must be installed and on PATH
- Linux — `syscall.Exec` is Unix-specific

## Testing Strategy

**Approach:** Interface-driven testing with fakes at the orchestration layer (`internal/gomux`). Concrete packages tested via their own unit tests.

**Key Test Areas:**

- **Orchestration (`internal/gomux/gomux_test.go`)**: Test `App.Run()` with fake lister, picker, and executor. Verify:
  - Direct mode: non-empty sessionName → executor called with that name, picker never called
  - Empty state: empty sessionName + lister returns empty slice → executor called with DefaultSessionName, picker never called
  - TUI mode: empty sessionName + lister returns sessions → sessions passed to picker, executor called with picker result
  - User quit: picker returns `ErrNoSelection` → `Run` returns nil (clean exit)
  - Lister error: propagated as error from `Run`
  - Executor error: propagated as error from `Run`
- **Session parsing (`internal/tmux/tmux_test.go`)**: Unit test `parseSessions` with sample `tmux list-sessions -F` output:
  - Multiple sessions with varying window counts, attached states, and creation timestamps
  - Empty output (no sessions / server not running)
  - Malformed lines (partial fields, non-numeric window count, invalid timestamp)
- **TUI model (`internal/tui/model_test.go`)**: Unit test the bubbletea `Update` function:
  - j/k navigation moves cursor
  - Enter on a session returns that session's name
  - `n` switches to input view
  - Typing in input view + Enter returns the typed name
  - Esc in input view returns to list view
  - `q` in list view returns `ErrNoSelection`
- **Name validation (`internal/tui/model_test.go`)**: Unit test input view validation:
  - Name with dot or colon → error message shown, Enter blocked
  - Name matching existing session → warning message shown, Enter still allowed (attaches)
  - Valid unique name → no error/warning, Enter submits

**Not tested (too thin to test):**
- `syscall.Exec` call in `tmux.Client.ExecIntoSession` (platform-specific, no meaningful fake)
- `exec.CommandContext` call in `tmux.Client.ListSessions` (real tmux invocation — integration concern)
- `tea.NewProgram` wiring in `tui.Picker.Pick` (thin delegation to bubbletea)
- `cmd/gomux/main.go` (pure wiring, no logic)

## Risks and Tradeoffs

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Bubbletea rendering issues over mosh | Low | Medium | Mosh supports full terminal emulation; bubbletea's alternate screen buffer works over mosh. Test early. |
| tmux format string changes between versions | Very Low | Low | Using stable format variables (`session_name`, `session_windows`, `session_attached`) that haven't changed in years. |
| syscall.Exec leaves terminal in bad state | Low | Medium | Bubbletea's `tea.Program` restores terminal state on exit before we exec. Sequence: TUI shutdown → exec. |

## File Layout

```
~/code/gomux/
├── cmd/
│   └── gomux/
│       └── main.go              # [new] Wiring only: create deps, build App, call Run
├── go.mod                       # [new] Module definition
├── lefthook.yml                 # [new] Pre-commit hooks (gofumpt, go mod tidy, go build, golangci-lint)
├── .golangci.yml                # [new] Linter configuration (default: all, with exclusions)
├── internal/
│   ├── gomux/
│   │   ├── gomux.go             # [new] Session type, interfaces, ErrNoSelection, App struct, New, Run
│   │   └── gomux_test.go        # [new] App.Run orchestration tests with fakes
│   ├── tmux/
│   │   ├── tmux.go              # [new] Client (ListSessions, ExecIntoSession, parseSessions)
│   │   └── tmux_test.go         # [new] parseSessions tests with sample tmux output
│   └── tui/
│       ├── model.go             # [new] Bubbletea Model (Init, Update, View), internal state machine
│       ├── model_test.go        # [new] Update function tests (key handling, state transitions)
│       ├── picker.go            # [new] Picker type, NewPicker, Pick (public API)
│       └── styles.go            # [new] Lipgloss style definitions
└── README.md                    # [new] Usage instructions and Termux setup
```

## Implementation Strategy

### Architectural Concerns

- **Testability**: All business logic lives in `internal/gomux` behind interfaces. Tests inject fakes for all three dependencies — no real tmux, no real TUI, no real exec.
- **Clean handoff**: The TUI must fully restore terminal state before `syscall.Exec`. Bubbletea handles this when `tea.Program.Run()` returns.
- **Error UX**: If tmux isn't installed, fail fast with a clear message rather than showing a broken TUI. `tmux.NewClient()` checks at construction time.

### Key Implementation Notes

- Use `tmux list-sessions -F '#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}'` for reliable parsing
- `#{session_created}` is a Unix timestamp — parse with `time.Unix(ts, 0)`, display as relative time ("2h ago") at render time
- `tmux new-session -A -s <name>` is the universal attach/create command for both TUI selection and CLI bypass
- The TUI has two states: **list view** (default, navigate + select) and **input view** (typing a new session name after pressing `n`)
- `exec.LookPath("tmux")` at startup to get the absolute path needed for `syscall.Exec`

---

_This summary is ephemeral - it will be deleted with the plan after implementation completes._
