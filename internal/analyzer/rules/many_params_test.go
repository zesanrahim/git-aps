package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestManyParamsRule_Name(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 5}
	if rule.Name() != "many_params" {
		t.Errorf("expected many_params, got %q", rule.Name())
	}
}

func TestManyParamsRule_DetectsExcessiveParams(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("func process(a, b, c, d string) error {", 10),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 10 {
		t.Errorf("expected line 10, got %d", findings[0].Line)
	}
}

func TestManyParamsRule_NoFindingWithinLimit(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 5}
	diff := makeFileDiff("foo.go",
		makeAddedLine("func process(a, b, c string) error {", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
}

func TestManyParamsRule_SkipsNonFuncLines(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("result := call(a, b, c, d, e)", 3),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-func lines, got %d", len(findings))
	}
}

func TestManyParamsRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 2}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: "func old(a, b, c, d string) {", OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestManyParamsRule_NoParens(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("func noParens", 1),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings when no parens, got %d", len(findings))
	}
}

func TestManyParamsRule_EmptyParens(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("func noArgs() {", 1),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for func with no args, got %d", len(findings))
	}
}

func TestManyParamsRule_ExactlyAtLimit(t *testing.T) {
	t.Parallel()
	rule := &ManyParamsRule{MaxParams: 3}
	diff := makeFileDiff("foo.go",
		makeAddedLine("func exactly(a, b, c string) {", 1),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings at exactly max params, got %d", len(findings))
	}
}

func TestManyParamsRule_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		line       string
		maxParams  int
		wantFindings int
	}{
		{"two params under limit 3", "func f(a, b int) {", 3, 0},
		{"three params at limit 3", "func f(a, b, c int) {", 3, 0},
		{"four params over limit 3", "func f(a, b, c, d int) {", 3, 1},
		{"six params over limit 5", "func f(a, b, c, d, e, f int) {", 5, 1},
		{"not a func line", "call(a, b, c, d)", 2, 0},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &ManyParamsRule{MaxParams: tc.maxParams}
			diff := makeFileDiff("x.go", makeAddedLine(tc.line, 1))
			findings := rule.Check(diff)
			if len(findings) != tc.wantFindings {
				t.Errorf("line %q: want %d findings, got %d", tc.line, tc.wantFindings, len(findings))
			}
		})
	}
}
