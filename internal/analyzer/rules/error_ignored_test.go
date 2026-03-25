package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

func TestErrorIgnoredRule_Name(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	if rule.Name() != "error_ignored" {
		t.Errorf("expected error_ignored, got %q", rule.Name())
	}
}

func TestErrorIgnoredRule_DetectsBlankAssign(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("_ = getErr()", 10),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 10 {
		t.Errorf("expected line 10, got %d", findings[0].Line)
	}
	if findings[0].Severity != analyzer.SeverityHigh {
		t.Errorf("expected HIGH severity, got %v", findings[0].Severity)
	}
}

func TestErrorIgnoredRule_DetectsErrBlankAssign(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("_ = err", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for '_ = err', got %d", len(findings))
	}
}

func TestErrorIgnoredRule_NoFindingForProperHandling(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	tests := []struct {
		name string
		line string
	}{
		{"if err check", "if err != nil {"},
		{"err assignment", "err := doSomething()"},
		{"return err", "return err"},
		{"wrap err", `return fmt.Errorf("doing X: %w", err)`},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			diff := makeFileDiff("foo.go", makeAddedLine(tc.line, 1))
			findings := rule.Check(diff)
			if len(findings) != 0 {
				t.Errorf("line %q: expected no findings, got %d", tc.line, len(findings))
			}
		})
	}
}

func TestErrorIgnoredRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: "_ = err", OldNum: 3},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestErrorIgnoredRule_DetectsCapitalErr(t *testing.T) {
	t.Parallel()
	rule := &ErrorIgnoredRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("_ = someErr", 7),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding for capital Err, got %d", len(findings))
	}
}

func TestIsIgnoredError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"blank assign err", "_ = err", true},
		{"blank assign someErr", "_ = someErr", true},
		{"blank assign func returning err", "_ = doThings()", false},
		{"if err", "if err != nil {", false},
		{"err := assign", "err := call()", false},
		{"normal blank assign", "_ = result", false},
		{"tab blank assign err", "_ =\terr", true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isIgnoredError(tc.line)
			if got != tc.want {
				t.Errorf("isIgnoredError(%q) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
}
