package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestDeferLoopRule_Name(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	if rule.Name() != "defer_loop" {
		t.Errorf("expected name defer_loop, got %q", rule.Name())
	}
}

func TestDeferLoopRule_DetectsDeferInsideLoop(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, f := range files {", 10),
		makeAddedLine("\tdefer f.Close()", 11),
		makeAddedLine("}", 12),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 11 {
		t.Errorf("expected line 11, got %d", findings[0].Line)
	}
	if findings[0].Rule != "defer_loop" {
		t.Errorf("expected rule defer_loop, got %q", findings[0].Rule)
	}
}

func TestDeferLoopRule_SkipsDeferOutsideLoop(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("f, _ := os.Open(name)", 5),
		makeAddedLine("defer f.Close()", 6),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for defer outside loop, got %d", len(findings))
	}
}

func TestDeferLoopRule_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	diff := makeFileDiff("main.py",
		makeAddedLine("for f in files:", 10),
		makeAddedLine("\tdefer f.close()", 11),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-Go file, got %d", len(findings))
	}
}

func TestDeferLoopRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	diff := makeFileDiff("main.go",
		makeAddedLine("for _, f := range files {", 10),
		git.DiffLine{Type: git.LineRemoved, Content: "\tdefer f.Close()", OldNum: 11},
		makeAddedLine("}", 11),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed defer, got %d", len(findings))
	}
}

func TestDeferLoopRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &DeferLoopRule{}
	diff := makeFileDiff("handler.go",
		makeAddedLine("for i := 0; i < n; i++ {", 20),
		makeAddedLine("\tdefer cleanup(i)", 21),
		makeAddedLine("}", 22),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "handler.go" {
		t.Errorf("expected file handler.go, got %q", f.File)
	}
	if f.Line != 21 {
		t.Errorf("expected line 21, got %d", f.Line)
	}
}
