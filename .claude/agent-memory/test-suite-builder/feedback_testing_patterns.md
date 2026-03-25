---
name: Testing patterns for git-aps
description: Patterns, gotchas, and conventions discovered while building the test suite for git-aps
type: feedback
---

Use `makeAddedLine` and `makeFileDiff` helpers (defined in magic_numbers_test.go) across all rules tests since they're in the same package `rules`.

**Why:** All rule test files share the `rules` package so helpers in one file are available to all. Avoids duplication.

**How to apply:** When adding rule tests, the helpers `makeAddedLine(content, lineNum)` and `makeFileDiff(path, lines...)` are already available.

`isIgnoredError` only triggers when the line contains BOTH `_ = ` AND either "err" or "Err". `_ = doSomething()` does NOT trigger it — only lines that reference an error variable like `_ = getErr()` or `_ = err` do.

**Why:** Source code at `error_ignored.go:43` checks for both the blank-assign pattern AND an "err"/"Err" substring.

**How to apply:** When writing error_ignored tests, use lines like `_ = getErr()` not `_ = doSomething()`.

When checking that a string was replaced in fixer tests, avoid `strings.Contains(got, "line2")` if the replacement text also contains "line2" (e.g., "replaced_line2"). Use `strings.Contains(got, "\nline2\n")` for exact line matching.

**Why:** "replaced_line2" contains "line2" as a substring, causing false positives.

**How to apply:** Use newline boundaries when asserting line absence in fixer tests.

The hook package's `hookPath()` function calls `git rev-parse --show-toplevel` which requires a real git repo. Added `var overriddenHookPath string` to `hook.go` and checked it first in `hookPath()` to enable injection in tests.

**Why:** Tests need to redirect hook operations to temp directories without running real git commands.

**How to apply:** In hook tests, set `overriddenHookPath = someTemp` and restore via `t.Cleanup`.

The `config.Load()` function reads from `GIT_APS_CONFIG` env var. Tests use `t.Setenv("GIT_APS_CONFIG", path)` which automatically restores the env after the test. Do NOT use `t.Parallel()` with `t.Setenv` on the same test since `t.Setenv` is incompatible with parallel.

**Why:** `t.Setenv` modifies process-global state and marks the test as using that env var, which is incompatible with `t.Parallel()`.

**How to apply:** Config tests that use `t.Setenv` must not call `t.Parallel()`.

The `cli.Parse()` function reads from `os.Args` directly. Tests must temporarily replace `os.Args` and restore it via `t.Cleanup`. Do not use `t.Parallel()` in cli tests since `os.Args` is process-global.

**Why:** Multiple parallel tests modifying `os.Args` simultaneously would race.
