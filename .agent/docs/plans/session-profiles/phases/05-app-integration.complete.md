# Phase 05: App Integration

**Depends on:** Phases 02, 03, 04
**Phase Type:** Refactor

---

## Existing Test Verification

**Run existing tests:** `go test ./internal/gomux/ ./internal/profile/ ./internal/tmux/ ./internal/tui/ -v`

**Expected result:** All tests pass — Phase 01 added the routing tests, Phases 02-04 implemented the components. This phase integrates them.

**Context:** The routing logic was implemented in Phase 01 (domain types). The profile loader, session creator, and picker changes were implemented in Phases 02-04. At this point, all components exist and pass their unit tests. This phase verifies that the full flow works by running the existing test suites together and ensuring no integration gaps.

### Gate: EXISTING TESTS

- [x] Existing test suite executed across all packages
- [x] All tests PASS before proceeding

---

## GREEN: Implement

**Objective:** Ensure all components integrate correctly. Fix any gaps discovered when running the full test suite.

**Files:**

- `internal/gomux/gomux.go` (minor adjustments if needed)
- `internal/gomux/gomux_test.go` (integration-level test if needed)

**Implementation Guidance:**

This phase is primarily a verification pass. If all existing tests pass, no code changes are required. The checklist below documents what to look for if tests fail. The routing logic in `App.Run` was written in Phase 01 against fakes. The real implementations now exist. Potential gaps to check:

```
// Verify these integration points:
//
// 1. App.Run calls ProfileLoader.LoadProfiles — does it handle empty result correctly?
// 2. Profile filtering (exclude profiles matching running sessions) — is the name comparison correct?
// 3. Selection.IsProfile triggers CreateSessionFromProfile → ExecIntoSession sequence
// 4. Empty state: no sessions AND no profiles → default session (skip TUI)
// 5. Empty state: no sessions BUT profiles exist → show picker with profiles only
```

If all tests already pass, this phase is a no-op verification.

### Gate: GREEN

- [x] `go test ./... -v` passes with zero failures
- [x] No integration gaps found, or gaps are fixed

---

## REFACTOR: Quality

**Focus:** Cross-package consistency review.

**Review Areas:**

- **Error chain**: Trace an error from profile loader → App.Run → caller. Is context preserved at each level?
- **Nil safety**: If ProfileLoader returns nil vs empty slice, does code handle both?
- **Empty states**: All combinations of sessions × profiles produce correct behavior

### Gate: REFACTOR

- [x] Error messages form readable chains across package boundaries
- [x] Nil slice handling is consistent (empty slice preferred over nil)
- [x] All empty state combinations verified

---

## Phase Complete

When all gates pass:

1. Rename this file to replace the current status suffix with `.complete.md` using `git mv`
2. Proceed to Phase 06

---

**Previous:** Phases 02, 03, 04 (parallel)
**Next:** Phase 06 - Main Wiring (final)
