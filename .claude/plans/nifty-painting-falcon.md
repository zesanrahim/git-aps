# git-aps Implementation Plan

## Context
Building a new Go TUI tool that scans git diffs for code anti-patterns using static rules + Kimi K2 AI, and lets users auto-apply fixes from the terminal.

## Phase 1: Project Scaffold
- [x] Create `.claude/CLAUDE.md` with project context
- [ ] Initialize Go module (`go mod init github.com/zesanrahim/git-aps`)
- [ ] Create directory structure
- [ ] Create `cmd/git-aps/main.go` entry point
- [ ] Add `.git-aps.yaml` default config
- [ ] Add `.gitignore`

## Phase 2: Core — Git Diff Parsing
- [ ] `internal/git/diff.go` — run `git diff`, parse output into structured hunks
- [ ] `internal/git/types.go` — `FileDiff`, `Hunk`, `DiffLine` types
- [ ] Tests for diff parsing

## Phase 3: Static Rule Engine
- [ ] `internal/analyzer/types.go` — `Analyzer` interface, `Finding` type
- [ ] `internal/analyzer/rules/engine.go` — rule runner
- [ ] `internal/analyzer/rules/` — individual rules (magic numbers, deep nesting, long functions, etc.)
- [ ] Tests for rules

## Phase 4: AI Analyzer (Kimi K2)
- [ ] `internal/analyzer/ai/client.go` — `ModelClient` interface + Kimi K2 implementation
- [ ] `internal/analyzer/ai/prompt.go` — prompt templates for anti-pattern analysis
- [ ] `internal/analyzer/ai/parser.go` — parse AI response into `Finding` structs
- [ ] `internal/analyzer/merge.go` — merge + deduplicate findings from rules + AI

## Phase 5: TUI
- [ ] `internal/ui/model.go` — Bubbletea main model
- [ ] `internal/ui/list.go` — findings list view
- [ ] `internal/ui/detail.go` — finding detail view with diff
- [ ] `internal/ui/styles.go` — Lipgloss styles
- [ ] `internal/ui/keymap.go` — key bindings

## Phase 6: Fix Application
- [ ] `internal/fixer/apply.go` — write AI-suggested fixes back to files
- [ ] Confirmation flow in TUI before applying

## Phase 7: Config & Polish
- [ ] `internal/config/config.go` — load `.git-aps.yaml`
- [ ] `git-aps install-hook` subcommand (optional git hook)
- [ ] README if requested

## Verification
- `go build ./cmd/git-aps` compiles
- `go test ./...` passes
- Run against a real repo with staged changes and see findings in TUI
