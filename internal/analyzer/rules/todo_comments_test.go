package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestTodoCommentRule_Name(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	if rule.Name() != "todo_comments" {
		t.Errorf("expected todo_comments, got %q", rule.Name())
	}
}

func TestTodoCommentRule_DetectsTODO(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("// TODO: fix this later", 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for TODO, got %d", len(findings))
	}
	if findings[0].Line != 5 {
		t.Errorf("expected line 5, got %d", findings[0].Line)
	}
}

func TestTodoCommentRule_AllMarkers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
	}{
		{"TODO", "// TODO: do something"},
		{"FIXME", "// FIXME: broken"},
		{"HACK", "// HACK: workaround"},
		{"XXX", "// XXX: danger"},
		{"lowercase todo", "// todo: fix this"},
		{"lowercase fixme", "// fixme: fix"},
		{"mixed case", "// Todo: later"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &TodoCommentRule{}
			diff := makeFileDiff("foo.go", makeAddedLine(tc.content, 1))
			findings := rule.Check(diff)
			if len(findings) != 1 {
				t.Errorf("line %q: expected 1 finding, got %d", tc.content, len(findings))
			}
		})
	}
}

func TestTodoCommentRule_NoFindingForNormalCode(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("result := doWork()", 2),
		makeAddedLine("return result", 3),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for normal code, got %d", len(findings))
	}
}

func TestTodoCommentRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: "// TODO: old todo", OldNum: 3},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed TODO lines, got %d", len(findings))
	}
}

func TestTodoCommentRule_SkipsContextLines(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineContext, Content: "// TODO: old todo", OldNum: 3, NewNum: 3},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for context TODO lines, got %d", len(findings))
	}
}

func TestTodoCommentRule_OnlyOneMatchPerLine(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("// TODO: FIXME: this is really bad", 10),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding even with multiple markers, got %d", len(findings))
	}
}

func TestTodoCommentRule_SkipsCodeStrings(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine(`fmt.Println("TODO: remove this debug")`, 8),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for TODO in string literal, got %d", len(findings))
	}
}

func TestTodoCommentRule_SkipsVariableNames(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("todoCount := 5", 8),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for variable name containing TODO, got %d", len(findings))
	}
}

func TestTodoCommentRule_InlineComment(t *testing.T) {
	t.Parallel()
	rule := &TodoCommentRule{}
	diff := makeFileDiff("foo.go",
		makeAddedLine("x := 1 // TODO: cleanup", 8),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding for inline TODO comment, got %d", len(findings))
	}
}
