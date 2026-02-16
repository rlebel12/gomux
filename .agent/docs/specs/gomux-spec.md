# Specification: gomux

**Related Architecture:** [gomux plan](../plans/gomux/README.md)

## Problem Statement

**Core Problem:** Managing multiple tmux sessions on a remote machine from a phone (Termux via mosh) requires memorizing session names and manually crafting tmux commands — there's no way to discover or manage sessions interactively.

**Current State:** The user connects from Termux via mosh and lands in a single hardcoded tmux session. To use multiple sessions, they must remember exact names and type `tmux attach -t <name>` or `tmux new -s <name>`. There's no discovery mechanism — if they forget a session name, they resort to `tmux list-sessions` and manually parse the output on a small phone screen.

**Desired State:** Connecting via mosh presents an interactive, phone-friendly session picker. Existing sessions are listed with useful metadata. Creating a new session is a single keypress away. For known session names, a CLI bypass skips the UI entirely. After selection, the picker disappears and the user is in tmux with no intermediate process lingering.

**Impact:** Single user (personal tool), but used multiple times daily. Eliminates friction in the most common remote workflow: phone → mosh → tmux.

## Success Criteria

- Selecting an existing session takes at most 3 keypresses (navigate + enter)
- Creating a new session takes at most N+3 keypresses (n + type name + enter)
- The TUI is usable on a Termux screen (minimal vertical space consumption, touch-friendly line heights)
- `gomux <name>` attaches/creates without any TUI flash — zero visual overhead
- No parent process remains after session attachment (clean `exec` handoff)

## User Journeys

### Primary Journey: Select an Existing Session

**Actor:** User connecting from Termux on phone

**Context:** User has one or more tmux sessions from previous connections and wants to resume one.

**Trigger:** User runs `home` shell function, which executes `mosh rob@home -- gomux`

**Steps:**

1. **System:** gomux launches, queries tmux for existing sessions, renders the TUI.

   - **Display:** Two-line-per-session list with cursor on the first session. Persistent help footer at the bottom.
   - **Outcome:** User sees all sessions at a glance with name, attached status, window count, and age.

2. **User Action:** Navigates with j/k (or arrow keys) to the desired session.

   - **System Response:** Cursor moves, selected session is visually highlighted.

3. **User Action:** Presses Enter.

   - **System Response:** TUI exits cleanly (terminal state restored), gomux `exec`s into `tmux new-session -A -s <name>`.
   - **Outcome:** User is inside the tmux session. gomux process no longer exists.

**Success Indicators:** User is in their tmux session within seconds of connecting.

**Failure Handling:** If tmux is not installed, gomux prints a clear error and exits before showing any TUI.

### Secondary Journey: Create a New Session

**Actor:** Same user, wants a new session for a different task.

**Trigger:** User is in the session list view and presses `n`.

**Steps:**

1. **User Action:** Presses `n` in list view.

   - **System Response:** View switches to text input mode. A text field appears for the session name.

2. **User Action:** Types a session name.

   - **System Response:** Input is validated as the user types:
     - Names containing dots or colons (tmux-invalid characters) show an inline error.
     - Names matching an existing session show a warning that this will attach to the existing session rather than create a new one.

3. **User Action:** Presses Enter.

   - **System Response:** TUI exits, gomux `exec`s into `tmux new-session -A -s <typed-name>`.

4. **Alternative — User Action:** Presses Esc.

   - **System Response:** Returns to list view. Input is discarded.

### Tertiary Journey: Direct Attach (CLI Bypass)

**Actor:** User who already knows the session name.

**Trigger:** User runs `home work` (which becomes `mosh rob@home -- gomux work`)

**Steps:**

1. **System:** gomux receives `"work"` as the session name argument. No TUI is rendered.

   - **System Response:** Immediately `exec`s into `tmux new-session -A -s work`.
   - **Outcome:** User is in tmux. No visual overhead whatsoever.

### Edge Case: No Existing Sessions

**Actor:** User connecting to a freshly rebooted machine or after all sessions have been closed.

**Trigger:** `gomux` launches with no arguments and tmux reports zero sessions.

**Behavior:** gomux skips the TUI entirely and `exec`s into `tmux new-session -A -s default`. The user lands in a session named "default" with no interaction required.

**Rationale:** On a phone keyboard, there's no value in forcing the user through a "create session" flow when there's nothing to choose from. The common case is that they just want a session.

## UX Design

### List View Layout

```
  gomux

  ▸ dev
    3 windows · attached · created 2h ago

    api-server
    1 window · created 1d ago

    scratch
    2 windows · created 5m ago

  ↑↓ navigate · enter attach · n new · q quit
```

**Layout details:**

- **Title:** "gomux" at the top, styled subtly.
- **Session rows:** Two lines per session. First line is the session name (bold when selected). Second line is indented metadata in a dimmer color.
- **Cursor:** `▸` marker on the selected session. Unselected sessions have a space indent for alignment.
- **Metadata line:** `{N} window(s) · {attached} · created {relative time}`
  - Window count always shown (singular/plural: "1 window", "3 windows")
  - "attached" shown only when the session has an attached client, styled with a distinct color (e.g., green or yellow) for emphasis
  - Relative creation time (e.g., "created 2h ago", "created 1d ago")
- **Footer:** Persistent single-line help bar at the bottom with keybindings. Dimmed/muted color so it doesn't compete with the session list.

### Input View Layout

```
  gomux

  New session name:
  ▌my-new-session

  enter confirm · esc cancel
```

- **Text input** with a blinking cursor.
- **Inline validation** appears below the input:
  - Invalid characters: `Name cannot contain dots or colons` (red/error color)
  - Duplicate name: `Session "dev" already exists — will attach` (yellow/warning color)
- **Footer** changes to show input-mode keybindings.

### Styling

- **Selected session name:** Bold
- **Unselected session name:** Normal weight
- **Metadata line:** Dim/muted color
- **"attached" badge:** Distinct color (green or yellow) — the primary visual emphasis distinguishing attached from detached sessions
- **Footer help bar:** Dim, single line, separator characters between bindings (` · `)
- **Error text:** Red
- **Warning text:** Yellow

### Keybindings

| Key | List View | Input View |
|-----|-----------|------------|
| `j` / `↓` | Move cursor down | — |
| `k` / `↑` | Move cursor up | — |
| `Enter` | Attach to selected session | Confirm name and exec |
| `n` | Switch to input view | — |
| `q` | Quit (no selection) | — |
| `Esc` | — | Return to list view |
| `Ctrl+C` | Quit (no selection) | Quit (no selection) |

### Tmux Format String

Session metadata is retrieved via:
```
tmux list-sessions -F '#{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}'
```

The `session_created` field is a Unix timestamp. gomux computes relative time (e.g., "2h ago") at render time.

## Requirements

### Functional Requirements

1. **Session Listing**

   - **Description:** Display all existing tmux sessions with name, window count, attached status, and relative creation time.
   - **Acceptance Criteria:**
     - Given tmux has 3 sessions running
     - When gomux launches with no arguments
     - Then all 3 sessions appear in the list with correct metadata
   - **Priority:** Critical

2. **Session Selection**

   - **Description:** User can navigate the list and select a session to attach to.
   - **Acceptance Criteria:**
     - Given the session list is displayed
     - When the user navigates to a session and presses Enter
     - Then gomux execs into tmux attached to that session
   - **Priority:** Critical

3. **New Session Creation**

   - **Description:** User can create a new session by pressing `n` and typing a name.
   - **Acceptance Criteria:**
     - Given the user presses `n` and types "my-session" and presses Enter
     - When the name is valid and doesn't match an existing session
     - Then gomux execs into a new tmux session named "my-session"
   - **Priority:** Critical

4. **Name Validation**

   - **Description:** Validate session names as the user types, blocking tmux-invalid characters and warning on duplicates.
   - **Acceptance Criteria:**
     - Given the user types a name containing a dot or colon
     - When each character is entered
     - Then an inline error message appears and Enter does not submit
   - **Acceptance Criteria (duplicate):**
     - Given the user types a name matching an existing session
     - When the full name matches
     - Then a warning appears indicating attachment will occur instead of creation
   - **Priority:** High

5. **CLI Direct Attach**

   - **Description:** When invoked with a session name argument, skip the TUI and exec immediately.
   - **Acceptance Criteria:**
     - Given gomux is invoked as `gomux work`
     - When the process starts
     - Then no TUI is rendered and gomux execs into `tmux new-session -A -s work`
   - **Priority:** Critical

6. **Empty State Auto-Create**

   - **Description:** When no tmux sessions exist, skip the TUI and create a session named "default".
   - **Acceptance Criteria:**
     - Given tmux reports zero sessions
     - When gomux launches with no arguments
     - Then gomux execs into `tmux new-session -A -s default` without showing the TUI
   - **Priority:** High

7. **Clean Process Handoff**

   - **Description:** gomux replaces itself with tmux via `syscall.Exec` — no parent process remains.
   - **Acceptance Criteria:**
     - Given the user selects a session
     - When tmux starts
     - Then the gomux process no longer exists (PID is now tmux)
   - **Priority:** Critical

### Non-Functional Requirements

- **Startup latency:** TUI should render within 200ms of gomux launch (tmux list-sessions is fast locally).
- **Screen efficiency:** Usable on a Termux screen (roughly 50 columns × 30 rows). The two-line layout plus footer should comfortably show 10+ sessions.
- **Terminal compatibility:** Must render correctly over mosh, which supports full terminal emulation but may have latency. Avoid rapid re-renders.

## Constraints & Assumptions

**Technical Constraints:**
- Linux only (`syscall.Exec` is Unix-specific)
- Requires tmux installed on the remote machine
- Runs over mosh (Termux → remote), which adds display latency

**Assumptions:**
- The user has at most ~20 tmux sessions at any time (no pagination needed)
- Termux supports the terminal capabilities Bubbletea requires (alternate screen buffer, basic ANSI)
- mosh faithfully relays Bubbletea's terminal output (validated — mosh supports full terminal emulation)

## Out of Scope

**Explicitly Excluded:**

- Session killing, renaming, or detailed inspection (use tmux directly for management)
- Mosh/SSH connection management (handled by Termux shell function)
- Live session content preview (would require tmux capture-pane, adds complexity)
- Configuration file or persistent settings (hardcoded behavior is fine for a personal tool)
- Client-side component (gomux runs on the remote machine only)
- Scrolling/pagination for the session list (assumption: <20 sessions)
- Session grouping or tagging

---

_Validation: Problem is clearly defined (tmux session management friction from a phone). User journeys cover the three primary flows (select, create, direct-attach) plus the empty-state edge case. Success criteria are measurable. Scope is tight._
