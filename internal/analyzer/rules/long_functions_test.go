package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestLongFunctionRule_Name(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 50}
	if rule.Name() != "long_functions" {
		t.Errorf("expected long_functions, got %q", rule.Name())
	}
}

func makeFuncLines(startNum int, count int) []git.DiffLine {
	lines := make([]git.DiffLine, 0, count+1)
	lines = append(lines, git.DiffLine{
		Type:    git.LineAdded,
		Content: "func bigFunc() {",
		NewNum:  startNum,
	})
	for i := 1; i < count; i++ {
		lines = append(lines, git.DiffLine{
			Type:    git.LineAdded,
			Content: "\tcode()",
			NewNum:  startNum + i,
		})
	}
	return lines
}

func TestLongFunctionRule_DetectsLongFunction(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 5}
	lines := makeFuncLines(1, 10)
	diff := makeFileDiff("foo.go", lines...)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 1 {
		t.Errorf("expected finding at line 1, got %d", findings[0].Line)
	}
}

func TestLongFunctionRule_NoFindingForShortFunction(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 50}
	lines := makeFuncLines(1, 10)
	diff := makeFileDiff("foo.go", lines...)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for short function, got %d", len(findings))
	}
}

func TestLongFunctionRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 5}
	var lines []git.DiffLine
	lines = append(lines, git.DiffLine{Type: git.LineAdded, Content: "func myFunc() {", NewNum: 1})
	for i := 2; i < 10; i++ {
		lines = append(lines, git.DiffLine{Type: git.LineRemoved, Content: "\tremoved()", OldNum: i})
	}
	diff := makeFileDiff("foo.go", lines...)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings (removed lines skipped), got %d", len(findings))
	}
}

func TestLongFunctionRule_MultipleFunctions(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 3}

	var lines []git.DiffLine
	lines = append(lines, git.DiffLine{Type: git.LineAdded, Content: "func shortFunc() {", NewNum: 1})
	lines = append(lines, git.DiffLine{Type: git.LineAdded, Content: "\tx()", NewNum: 2})

	lines = append(lines, git.DiffLine{Type: git.LineAdded, Content: "func longFunc() {", NewNum: 3})
	for i := 4; i <= 10; i++ {
		lines = append(lines, git.DiffLine{Type: git.LineAdded, Content: "\tx()", NewNum: i})
	}

	diff := makeFileDiff("foo.go", lines...)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (only longFunc), got %d", len(findings))
	}
	if findings[0].Line != 3 {
		t.Errorf("expected finding at line 3 (longFunc), got %d", findings[0].Line)
	}
}

func TestLongFunctionRule_EmptyHunk(t *testing.T) {
	t.Parallel()
	rule := &LongFunctionRule{MaxLines: 5}
	diff := git.FileDiff{
		Path:  "foo.go",
		Hunks: []git.Hunk{},
	}
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for empty hunk, got %d", len(findings))
	}
}
