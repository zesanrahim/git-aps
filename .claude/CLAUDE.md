# git-aps — Git Anti-Pattern Search

## Project Overview
A Go TUI application that analyzes git diffs for code anti-patterns using both static rules and AI (via OpenAI-compatible API). Users can view findings, see details, and auto-apply AI-generated fixes with confirmation. Also supports CI-friendly JSON/text output and pre-commit hook integration.

## Tech Stack
- **Language:** Go
- **TUI:** Bubbletea + Lipgloss (charmbracelet)
- **AI Model:** Configurable via OpenAI-compatible API (default: Gemini 2.5 Flash)
- **AI SDK:** github.com/sashabaranov/go-openai
- **Config:** YAML (.git-aps.yaml)

## Architecture

### Directory Structure
```
cmd/git-aps/          # CLI entry point
internal/
  cli/                # CLI flag parsing
  git/                # Git operations (diff parsing, file reading)
  analyzer/           # Anti-pattern detection
    rules/            # Static rule-based checks + custom regex rules
    ai/               # AI-powered analysis (retry logic, finding validation)
  ui/                 # Bubbletea TUI (models, views, styles)
  fixer/              # Auto-apply AI-generated fixes
  config/             # Config loading (.git-aps.yaml)
  output/             # JSON and text output formatters
  hook/               # Pre-commit hook install/uninstall
```

### Key Interfaces
- `analyzer.Analyzer` — common interface for rule-based and AI analyzers
- `analyzer.Finding` — represents a detected anti-pattern (severity, location, description, suggested fix)
- `rules.Rule` — interface for both built-in and custom regex rules

### Flow
1. `git diff --staged` (or configurable: unstaged, HEAD~1, branch)
2. Parse diff into per-file hunks
3. Run static rules + custom regex rules (fast, local) in parallel with AI analysis
4. Merge and deduplicate findings
5. Apply severity filtering
6. Output: TUI (interactive), JSON, or text (CI-friendly)
7. TUI: navigate, view details, apply fixes (single or batch), filter by severity
8. Exit summary shows applied/skipped/remaining counts

### Static Rules (built-in)

**General (all languages):**
- Magic numbers (configurable threshold, skips IPs/versions/hex/indices/strings)
- Deep nesting (configurable max depth, skips comments/blanks/closing braces)
- Long functions (configurable max code lines, excludes blank/comment lines)
- TODO/FIXME/HACK/XXX comments (only in comment context, not variable names)
- Cognitive complexity (configurable threshold, nesting-weighted scoring)

**Go-specific:**
- Unused error returns (`_ = err` patterns)
- Error wrapping without %w (`fmt.Errorf` with %v/%s instead of %w)
- Defer in loops (resource leak risk)
- Unsafe type assertions (without comma-ok check)
- Goroutine without context (missing ctx propagation)
- God functions (configurable max parameters, paren-depth-aware counting)

**Security (all languages):**
- Hardcoded secrets (API keys, passwords, tokens, private keys)
- SQL injection (string concat/formatting in SQL queries)

**Performance:**
- String concatenation in loops (suggest strings.Builder)
- Regex compilation in loops (Go-specific)

### Custom Rules
Users can define regex-based rules in `.git-aps.yaml` under `custom_rules` with name, pattern, severity, description, and suggestion.

### Helper Utilities
- `rules.DetectLanguage(path)` — language detection from file extension (Go, Python, JS, TS, Java, Rust)
- `rules.IsComment()`, `rules.IsBlankOrComment()`, `rules.StripComment()` — language-aware line analysis
- Go-specific rules auto-skip non-Go files

## Commands
- `go run ./cmd/git-aps` — run the app (TUI mode)
- `go run ./cmd/git-aps --format json` — JSON output for CI
- `go run ./cmd/git-aps --format text` — text output for CI
- `go run ./cmd/git-aps --min-severity high` — only show HIGH findings
- `go run ./cmd/git-aps --mode unstaged --no-ai` — unstaged diff, no AI
- `go run ./cmd/git-aps hook install` — install pre-commit hook
- `go run ./cmd/git-aps hook uninstall` — remove pre-commit hook
- `go test ./...` — run all tests
- `go test ./... -v` — run all tests with verbose output
- `go test ./... -cover` — run tests with coverage report
- `go build -o git-aps ./cmd/git-aps` — build binary

## CLI Flags
- `--mode` — diff mode: staged, unstaged, head, branch
- `--no-ai` — disable AI analysis
- `--min-severity` — minimum severity: low, med, high
- `--format` — output format: tui, json, text
- `--debug` — enable debug output

## Environment Variables
- `AI_API_KEY` — API key for AI model
- `GIT_APS_CONFIG` — override config file path (default: .git-aps.yaml)

## TUI Key Bindings
- `↑/k` `↓/j` — navigate
- `enter` — view details
- `a` — apply single fix
- `A` — batch apply all fixable findings
- `s` — skip finding
- `f` — cycle severity filter (all → MED+ → HIGH → all)
- `q` / `Ctrl+C` — quit

## Conventions
- Use `internal/` for all non-CLI packages
- Errors should be wrapped with context: `fmt.Errorf("doing X: %w", err)`
- Tests live alongside source files (`_test.go`)
- Keep TUI model/view/update cleanly separated in `internal/ui/`
