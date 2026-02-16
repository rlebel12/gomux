# Phase 03: Tmux Session Parsing

**Depends on:** Phase 01
**Phase Type:** Standard
**Status:** Complete

<!-- Standard phase: Full RED-GREEN-REFACTOR -->
<!-- Status progression: Pending -> RED -> GREEN -> Complete -->

---

## RED: Write Tests

**Objective:** Test the `parseSessions` function that converts raw `tmux list-sessions -F` output into `[]gomux.Session` structs.

**Files:**

- `internal/tmux/tmux_test.go` (use `package tmux`, not `package tmux_test`, since `parseSessions` is unexported)

**Test Cases:**

**Parameterized Tests** (table format -- MUST use `t.Run`):

| Case | Input (raw output) | Expected Sessions | Notes |
|------|-------------------|-------------------|-------|
| `single_detached` | `"dev\t3\t0\t1700000000\n"` | `[{Name:"dev", Windows:3, Attached:false, Created:time.Unix(1700000000,0)}]` | Basic single session |
| `single_attached` | `"work\t1\t1\t1700001000\n"` | `[{Name:"work", Windows:1, Attached:true, Created:time.Unix(1700001000,0)}]` | Attached flag parsing |
| `multiple_sessions` | `"dev\t3\t0\t1700000000\napi\t1\t1\t1700001000\nscratch\t2\t0\t1700002000\n"` | 3 sessions with correct fields | Multiple sessions, mixed attached states |
| `empty_output` | `""` | `[]` (empty slice, not nil) | No tmux server or no sessions |
| `whitespace_only` | `"\n"` | `[]` (empty slice) | Trailing newline only |

**Error Case Tests** (table format -- MUST use `t.Run`):

| Case | Input (malformed) | Expected Error Contains | Notes |
|------|-------------------|------------------------|-------|
| `missing_fields` | `"dev\t3\n"` | `"expected 4 fields"` | Only 2 of 4 tab-separated fields |
| `non_numeric_windows` | `"dev\tabc\t0\t1700000000\n"` | `"parse window count"` | Non-integer window count |
| `non_numeric_timestamp` | `"dev\t3\t0\tnot-a-number\n"` | `"parse created timestamp"` | Non-integer creation timestamp |
| `invalid_attached` | `"dev\t3\t2\t1700000000\n"` | `"parse attached flag"` | Attached field is not 0 or 1 |

**Assertions:**

- Successful parse returns correct Session field values
- Empty/whitespace input returns empty slice (not nil), no error
- Malformed lines return descriptive errors with line context
- Session order is preserved from input

### Gate: RED

- [x] Test file created with all enumerated test cases
- [x] All tests FAIL (`parseSessions` does not exist yet)
- [x] Test coverage includes happy path and all edge cases

---

## GREEN: Implement

**Objective:** Implement `parseSessions` as an unexported function in `internal/tmux/tmux.go`.

**Files:**

- `internal/tmux/tmux.go`

**Implementation Guidance:**

```go
// parseSessions parses tmux list-sessions -F output into Session structs.
//
// Expected format per line (tab-delimited):
//   #{session_name}\t#{session_windows}\t#{?session_attached,1,0}\t#{session_created}
//
// Implementation approach:
// 1. Split output by newlines
// 2. Skip empty lines (handles trailing newline)
// 3. For each non-empty line:
//    a. Split by tab → expect exactly 4 fields
//    b. Field 0: session name (string, use as-is)
//    c. Field 1: window count → strconv.Atoi
//    d. Field 2: attached flag → "1" means true, "0" means false, else error
//    e. Field 3: created timestamp → strconv.ParseInt → time.Unix(ts, 0)
// 4. Return slice of Session structs
//
// Error handling:
// - Wrong field count → fmt.Errorf("line %d: expected 4 fields, got %d", lineNum, count)
// - Invalid window count → fmt.Errorf("line %d: parse window count: %w", lineNum, err)
// - Invalid attached → fmt.Errorf("line %d: parse attached flag: expected 0 or 1, got %q", ...)
// - Invalid timestamp → fmt.Errorf("line %d: parse created timestamp: %w", lineNum, err)
```

**Pattern Application:**

- **Error Handling**: Use `fmt.Errorf` with `%w` wrapping for parse errors. Include line number context for debugging malformed tmux output.

### Gate: GREEN

- [x] All tests from RED phase now PASS
- [x] Test command: `go test ./internal/tmux/ -v` (9 tests)
- [x] Implementation follows pseudocode logic flow

---

## REFACTOR: Quality

**Focus:** Parsing code clarity and robustness.

**Review Areas:**

- **Error messages**: Include line number and the problematic value for easy debugging
- **Naming**: Clear variable names for parsed fields
- **Simplification**: Ensure the parsing loop is straightforward
- **Edge cases**: Confirm empty input returns empty slice (not nil)

### Gate: REFACTOR

- [x] Reviewed for code duplication and extracted common patterns
- [x] Variable and function names are clear and descriptive
- [x] Complex logic simplified where possible
- [x] Error messages are helpful and actionable

---

## Phase Complete

When all gates pass:

1. Update this file's status to **Complete**
2. Update index.md status table
3. Proceed to Phase 04

---

**Previous:** [Phase 01](01-domain.md)
**Next:** [Phase 04](04-tmux-client.md)
