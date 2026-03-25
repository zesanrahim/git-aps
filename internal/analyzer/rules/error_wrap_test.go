package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestErrorWrapRule_Name(t *testing.T) {
	t.Parallel()
	rule := &ErrorWrapRule{}
	if rule.Name() != "error_wrap" {
		t.Errorf("expected name error_wrap, got %q", rule.Name())
	}
}

func TestErrorWrapRule_DetectsVerbFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		line      string
		wantCount int
	}{
		{
			name:      "percent v with err",
			line:      `return fmt.Errorf("reading config: %v", err)`,
			wantCount: 1,
		},
		{
			name:      "percent s with err",
			line:      `return fmt.Errorf("failed: %s", err)`,
			wantCount: 1,
		},
		{
			name:      "percent v with Err-prefixed variable",
			line:      `return fmt.Errorf("loading: %v", parseErr)`,
			wantCount: 1,
		},
		{
			name:      "correct wrap with percent w",
			line:      `return fmt.Errorf("reading config: %w", err)`,
			wantCount: 0,
		},
		{
			name:      "non-error variable with percent v",
			line:      `return fmt.Errorf("value is %v", count)`,
			wantCount: 0,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &ErrorWrapRule{}
			diff := makeFileDiff("service.go", makeAddedLine(tc.line, 5))
			findings := rule.Check(diff)
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d for line %q", len(findings), tc.wantCount, tc.line)
			}
		})
	}
}

func TestErrorWrapRule_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &ErrorWrapRule{}
	diff := makeFileDiff("service.py",
		makeAddedLine(`return fmt.Errorf("reading config: %v", err)`, 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-Go file, got %d", len(findings))
	}
}

func TestErrorWrapRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &ErrorWrapRule{}
	diff := makeFileDiff("service.go",
		git.DiffLine{Type: git.LineRemoved, Content: `return fmt.Errorf("reading: %v", err)`, OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestErrorWrapRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &ErrorWrapRule{}
	diff := makeFileDiff("handler.go",
		makeAddedLine(`return fmt.Errorf("reading: %v", err)`, 42),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "handler.go" {
		t.Errorf("expected file handler.go, got %q", f.File)
	}
	if f.Line != 42 {
		t.Errorf("expected line 42, got %d", f.Line)
	}
	if f.Rule != "error_wrap" {
		t.Errorf("expected rule error_wrap, got %q", f.Rule)
	}
}
