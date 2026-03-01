# Design Summary: Session Profiles

**Status:** Pending Review

## Problem Statement

When a tmux session is destroyed (exiting last shell, server restart, resurrect failure), it's gone. The user must manually recreate the session with the right name, windows, pane splits, working directories, and startup commands. For a phone-based workflow, this is particularly painful — rebuilding layout from a small screen keyboard is slow and error-prone.

Sessions should be definable as persistent profiles that always appear in the picker and can be instantiated on demand with their full layout.

## Proposed Solution

Add a profile system to gomux. Profiles are defined in a single YAML file (`~/.config/gomux/profiles.yml`). Each profile specifies a session name and a layout of windows, optional pane splits, working directories, and startup commands.

The picker gains a second section below running sessions showing available (not-yet-running) profiles. Selecting a profile creates the full tmux session layout, then attaches. Profiles whose sessions are already running appear in the existing sessions section as normal — they don't appear twice.

No templating engine, no plugin hooks, no environment variable substitution. Profiles are static layout definitions.

## Referenced Patterns

| Pattern | Source | Application |
|---------|--------|-------------|
| Dependency Injection | `~/.agent/docs/go/references/dependency-injection.md` | `ProfileLoader` interface injected into App, testable with in-memory fakes |
| Error Handling | `~/.agent/docs/go/references/error-handling.md` | `fmt.Errorf` wrapping for file/YAML parse errors (no BaseError — too heavy for this tool) |
| Context Lifecycle | `~/.agent/docs/go/references/context-lifecycle.md` | Profile loading methods accept `context.Context` |

**Patterns considered but excluded:**

- **Type-Safe IDs**: Overkill for a personal CLI with two string identifiers (session name, profile name). The type safety doesn't pay for itself here.
- **BaseError/slog integration**: gomux has no logging. Errors go to stderr via `fmt.Fprintf`. Adding structured error types for a tool with one consumer adds complexity with no benefit.

## Design Decisions

### Decision 1: Single File vs. Directory of Files

**Context:** Where and how to store profile definitions.

**Options Considered:**
- **Single file (`profiles.yml`)**: All profiles in one file. Simple to edit, back up, and version control.
- **One file per profile**: More modular, but adds directory scanning, naming conventions, and more filesystem operations.

**Decision:** Single file at `~/.config/gomux/profiles.yml`.

**Rationale:** The user expects <20 sessions. A single file is easier to edit on a phone and simpler to implement. The file path follows XDG conventions.

**Tradeoffs:** Can't drop profile files into a directory for ad-hoc additions. Acceptable for a personal tool.

### Decision 2: Profile Layout Depth

**Context:** How much tmux layout detail can a profile specify?

**Decision:** Full layout — windows with optional pane splits, each with directory and startup command.

**Rationale:** The whole point is avoiding manual reconstruction. If profiles can't recreate pane splits, users still have manual work after session creation.

**Tradeoffs:** More complex tmux command orchestration. Pane splits require sequential `split-window` + `send-keys` calls. Worth it for the UX.

### Decision 3: Session Creation Strategy

**Context:** How to build a multi-window, multi-pane session via tmux CLI.

**Decision:** Use tmux commands sequentially:
1. `tmux new-session -d -s <name> -c <dir>` (creates session with first window)
2. `tmux send-keys -t <name>:0 '<command>' Enter` (run command in first window)
3. `tmux split-window -t <name>:0 -c <dir>` (split pane)
4. `tmux new-window -t <name> -c <dir>` (additional windows)
5. `tmux select-window -t <name>:0` (focus first window)
6. Finally, `syscall.Exec` into `tmux attach -t <name>`

**Rationale:** This is how tmuxinator and tmuxp work under the hood. No tmux API — just CLI commands. Reliable and well-understood.

**Tradeoffs:** Sequential command execution adds latency proportional to layout complexity. For typical layouts (2-4 windows, 1-2 panes each), this is <100ms total.

### Decision 4: ProfileLoader as New Interface

**Context:** How to integrate profile loading into the existing App architecture.

**Decision:** Define a `ProfileLoader` interface in the `gomux` package, injected into `App` via constructor. The `tmux` package does NOT implement this — a new `profile` package handles file loading and YAML parsing.

**Rationale:** Profile loading is file I/O + YAML parsing, not tmux interaction. Keeping it in a separate package maintains the clean separation of concerns. The `tmux.Client` gains a `SessionCreator` interface method for building sessions from profile data.

### Decision 5: Picker Integration

**Context:** How to display profiles alongside running sessions.

**Decision:** Separate section in the picker — running sessions first, then a divider, then available profiles. Profiles show their window count from the definition. The cursor can move between both sections.

**Rationale:** Visual separation makes it clear what's running vs. what will be created. A unified list would require badges/icons that are hard to distinguish on a phone screen.

## Component Scope

**In Scope:**
- YAML profile file parsing and validation
- Profile-aware picker UI with separate section
- Full session layout creation (windows, panes, dirs, commands)
- Filtering: profiles whose sessions already exist don't appear in the profile section
- `SessionCreator` interface on `tmux.Client` for building sessions from profiles

**Out of Scope:**
- Profile creation/editing via TUI (edit the YAML file directly)
- Environment variable substitution in profiles
- Conditional logic or templating in profiles
- Profile inheritance or composition
- Session killing/management
- Profile syncing or remote profile loading

## Interface Definitions

### Domain Types (`internal/gomux/gomux.go`)

```go
// Profile represents a tmux session profile defined in configuration.
type Profile struct {
	Name    string
	Windows []Window
}

// Window represents a tmux window within a profile.
type Window struct {
	Name    string
	Dir     string
	Command string
	Panes   []Pane
}

// Pane represents a pane split within a window.
type Pane struct {
	Dir       string
	Command   string
	Split     string // "horizontal" or "vertical"
}
```

### ProfileLoader Interface (`internal/gomux/gomux.go`)

```go
// ProfileLoader loads session profiles from configuration.
type ProfileLoader interface {
	LoadProfiles(ctx context.Context) ([]Profile, error)
}
```

### SessionCreator Interface (`internal/gomux/gomux.go`)

```go
// SessionCreator creates a tmux session from a profile layout.
type SessionCreator interface {
	CreateSessionFromProfile(ctx context.Context, profile Profile) error
}
```

### Updated App Constructor

```go
func New(
	lister SessionLister,
	picker SessionPicker,
	executor SessionExecutor,
	profiles ProfileLoader,
	creator SessionCreator,
) *App
```

### Updated SessionPicker

```go
// SessionPicker presents sessions and profiles to the user.
type SessionPicker interface {
	Pick(ctx context.Context, sessions []Session, profiles []Profile) (selection Selection, err error)
}

// Selection represents the user's choice from the picker.
type Selection struct {
	Name      string
	IsProfile bool // true if selected from profiles section
}
```

### YAML Schema (`~/.config/gomux/profiles.yml`)

```yaml
profiles:
  - name: dev
    windows:
      - name: editor
        dir: ~/code/project
        command: nvim
      - name: shell
        dir: ~/code/project
      - name: logs
        dir: ~/code/project
        panes:
          - command: "journalctl -f -u myservice"
          - split: horizontal
            command: "tail -f /var/log/app.log"

  - name: admin
    windows:
      - name: main
        dir: ~/Documents/admin
```

## Dependencies

**Internal Dependencies:**
- `internal/gomux` — domain types (Profile, Window, Pane, Selection)
- `internal/tmux` — gains `CreateSessionFromProfile` method
- `internal/tui` — picker model updated for two-section display

**External Dependencies:**
- `gopkg.in/yaml.v3` (new) — YAML parsing. Standard, well-maintained, no alternatives needed.

## Testing Strategy

**Approach:** Same pattern as existing tests — in-memory fakes, parameterized test tables, no mocking frameworks.

**Key Test Areas:**
- **YAML parsing**: Valid profiles, malformed YAML, missing required fields, empty profiles list
- **Profile filtering**: Profiles whose sessions exist are excluded from picker
- **Picker model**: Navigation across session and profile sections, selection returns correct `Selection` type
- **Session creation**: Correct sequence of tmux commands generated from profile definition (tested via fake executor that records commands)
- **App.Run integration**: Profile selected → session created → exec into session

## Risks and Tradeoffs

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| YAML parse errors on phone are hard to debug | Medium | Low | Clear error messages with line numbers; validate on load, fail fast |
| Sequential tmux commands fail mid-creation (partial session) | Low | Medium | If creation fails partway, kill the partial session before returning error |
| `~` expansion in `dir` fields doesn't work | Medium | Medium | Explicitly expand `~` to `$HOME` before passing to tmux |
| Profile file doesn't exist on first run | High | Low | Return empty profiles list (not an error) when file is absent |
| Pane split ordering affects layout | Low | Low | Document that panes split from the previous pane; match tmuxinator convention |

## File Layout

```
internal/
├── gomux/
│   └── gomux.go                     # [modify] Add Profile, Window, Pane, Selection types;
│                                     #          ProfileLoader, SessionCreator interfaces;
│                                     #          update App struct, constructor, Run method
│   └── gomux_test.go                # [modify] Add tests for profile-aware Run routing
├── profile/
│   ├── loader.go                    # [new] YAML file loader implementing ProfileLoader
│   └── loader_test.go              # [new] YAML parsing tests, missing file handling
├── tmux/
│   ├── tmux.go                      # [modify] Add CreateSessionFromProfile method
│   └── tmux_test.go                # [modify] Add session creation command tests
└── tui/
    ├── model.go                     # [modify] Add profiles section, Selection return,
    │                                #          two-section navigation
    ├── model_test.go               # [modify] Add profile section navigation/selection tests
    └── styles.go                    # [modify] Add profile section styles (divider, badge)
cmd/
└── gomux/
    └── main.go                      # [modify] Wire ProfileLoader, pass to App constructor
go.mod                               # [modify] Add gopkg.in/yaml.v3 dependency
```

## Implementation Strategy

### Architectural Concerns

- **Separation of concerns**: Profile loading (file I/O + YAML) stays in its own package. The `tmux` package handles session creation commands. The `gomux` domain package defines the interfaces and orchestrates the flow.
- **Backward compatibility**: When no profiles file exists, behavior is identical to current gomux. Profiles are purely additive.
- **Testability**: `ProfileLoader` interface allows tests to inject fake profiles without touching the filesystem. Session creation can be tested by recording tmux commands via a fake.

### Key Implementation Notes

- The `Pick` signature changes from `(ctx, sessions)` to `(ctx, sessions, profiles)` — this is a breaking interface change but the consumer (App) and implementation (Picker) are in the same project.
- `CreateSessionFromProfile` should be a method on `tmux.Client`, not a standalone function, since it needs the tmux binary path.
- `~` expansion must happen in the profile loader, not in the tmux client — the client receives absolute paths.
- The picker's cursor navigation needs to handle crossing the section boundary (sessions → divider → profiles) cleanly.
- Empty states: no sessions + no profiles = current behavior (auto-create "default"). No sessions + profiles = show only profiles section. Sessions + no profiles = current behavior (no profile section shown).

---

_This summary is ephemeral - it will be deleted with the plan after implementation completes._
