# Phase 06: Main Wiring

**Depends on:** Phase 05
**Phase Type:** Infrastructure

---

## Existing Test Verification

**Run existing tests:** `go test ./... -v`

**Expected result:** All tests pass before wiring changes.

### Gate: EXISTING TESTS

- [ ] Existing test suite executed
- [ ] All tests PASS before proceeding

---

## GREEN: Implement

**Objective:** Wire the ProfileLoader and SessionCreator into main.go and add the yaml.v3 dependency.

**Files:**

- `cmd/gomux/main.go`
- `go.mod` (via `go get`)

**Implementation Guidance:**

```go
// cmd/gomux/main.go — updated run() function:
//
// Implementation approach:
// 1. Add import for internal/profile package
// 2. Determine config path: os.UserHomeDir() + "/.config/gomux/profiles.yml"
//    - If UserHomeDir fails, use empty path (loader will return empty profiles)
// 3. Create profile.NewLoader(configPath)
// 4. Update gomux.New() call:
//    gomux.New(client, tui.NewPicker(), client, loader, client)
//    - client (*tmux.Client) satisfies SessionLister, SessionExecutor, and SessionCreator
//    - tui.NewPicker() satisfies the updated SessionPicker interface (Phase 04 changed Pick signature)
//    - loader (*profile.Loader) satisfies ProfileLoader
// 5. Error wrapping: wrap UserHomeDir error if it fails
```

```bash
# Add yaml.v3 dependency:
go get gopkg.in/yaml.v3
```

### Gate: GREEN

- [ ] `go build ./cmd/gomux/` compiles successfully
- [ ] `go test ./... -v` passes
- [ ] `go mod tidy` leaves clean go.mod/go.sum
- [ ] Binary runs: `./gomux` shows picker (or default session if no tmux)

---

## REFACTOR: Quality

**Focus:** Clean wiring, no unnecessary complexity.

**Review Areas:**

- **Config path**: Is the path construction clean? Consider a const for the relative path
- **Error handling**: UserHomeDir failure should degrade gracefully (empty profiles), not crash
- **Import organization**: New imports grouped correctly

### Gate: REFACTOR

- [ ] Config path construction is clean and readable
- [ ] Graceful degradation when home dir unavailable
- [ ] Imports properly organized
- [ ] No unnecessary error wrapping

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Implementation is complete — delete entire plan directory

---

**Previous:** Phase 05 - App Integration
**Next:** Final phase — implementation complete
