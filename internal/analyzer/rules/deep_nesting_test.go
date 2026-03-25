package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestDeepNestingRule_Name(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	if rule.Name() != "deep_nesting" {
		t.Errorf("expected deep_nesting, got %q", rule.Name())
	}
}

func TestDeepNestingRule_DetectsTabs(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("\t\t\t\tdeepCode()", 10),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 10 {
		t.Errorf("expected line 10, got %d", findings[0].Line)
	}
}

func TestDeepNestingRule_NoFindingWithinDepth(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("\t\t\tthreeDeep()", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings at exactly max depth, got %d", len(findings))
	}
}

func TestDeepNestingRule_DetectsSpaces(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("                    deepSpaces()", 7),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (space indented), got %d", len(findings))
	}
}

func TestDeepNestingRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: "\t\t\t\tdeepCode()", OldNum: 3},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestDeepNestingRule_EmptyLine(t *testing.T) {
	t.Parallel()
	rule := &DeepNestingRule{MaxDepth: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("", 4),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for empty line, got %d", len(findings))
	}
}

func TestNestingDepth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		line  string
		want  int
	}{
		{"no indent", "code()", 0},
		{"one tab", "\tcode()", 1},
		{"two tabs", "\t\tcode()", 2},
		{"four tabs", "\t\t\t\tcode()", 4},
		{"four spaces", "    code()", 1},
		{"eight spaces", "        code()", 2},
		{"sixteen spaces", "                code()", 4},
		{"empty", "", 0},
		{"only spaces", "    ", 0},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := nestingDepth(tc.line)
			if got != tc.want {
				t.Errorf("nestingDepth(%q) = %d, want %d", tc.line, got, tc.want)
			}
		})
	}
}
