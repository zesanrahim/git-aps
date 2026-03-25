package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func makeAddedLine(content string, lineNum int) git.DiffLine {
	return git.DiffLine{
		Type:    git.LineAdded,
		Content: content,
		NewNum:  lineNum,
	}
}

func makeFileDiff(path string, lines ...git.DiffLine) git.FileDiff {
	return git.FileDiff{
		Path: path,
		Hunks: []git.Hunk{
			{Lines: lines},
		},
	}
}

func TestMagicNumberRule_DetectsAboveThreshold(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("timeout := 300", 10),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 10 {
		t.Errorf("expected line 10, got %d", findings[0].Line)
	}
	if findings[0].File != "foo.go" {
		t.Errorf("expected file foo.go, got %q", findings[0].File)
	}
}

func TestMagicNumberRule_BelowThreshold(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 10}
	diff := makeFileDiff("foo.go",
		makeAddedLine("x := 2", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
}

func TestMagicNumberRule_SkipsConstDeclarations(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("const MaxRetries = 100", 1),
		makeAddedLine("var DefaultTimeout = 500", 2),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for const/var declarations, got %d", len(findings))
	}
}

func TestMagicNumberRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: "x := 999", OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestMagicNumberRule_SkipsContextLines(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineContext, Content: "x := 999", OldNum: 5, NewNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for context lines, got %d", len(findings))
	}
}

func TestMagicNumberRule_NegativeNumbers(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	diff := makeFileDiff("foo.go",
		makeAddedLine("offset := -500", 3),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding for negative magic number, got %d", len(findings))
	}
}

func TestMagicNumberRule_Name(t *testing.T) {
	t.Parallel()
	rule := &MagicNumberRule{Threshold: 2}
	if rule.Name() != "magic_numbers" {
		t.Errorf("expected name magic_numbers, got %q", rule.Name())
	}
}

func TestContainsMagicNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		line      string
		threshold int
		want      bool
	}{
		{"large number", "sleep(3600)", 2, true},
		{"within threshold", "x := 2", 2, false},
		{"exactly threshold", "x := 2", 2, false},
		{"just above threshold", "x := 3", 2, true},
		{"float", "ratio := 3.14", 2, true},
		{"const decl", "const N = 1000", 2, false},
		{"var decl", "var max = 1000", 2, false},
		{"negative large", "delta := -100", 2, true},
		{"zero", "x := 0", 2, false},
		{"one", "x := 1", 2, false},
		{"ip address", `addr := "192.168.1.100"`, 2, false},
		{"hex literal", "mask := 0xFF", 2, false},
		{"comment with number", "// timeout is 300", 2, false},
		{"array index", "items[3]", 2, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := containsMagicNumber(tc.line, tc.threshold, LangGo)
			if got != tc.want {
				t.Errorf("containsMagicNumber(%q, %d) = %v, want %v", tc.line, tc.threshold, got, tc.want)
			}
		})
	}
}
