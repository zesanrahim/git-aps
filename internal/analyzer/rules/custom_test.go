package rules

import (
	"regexp"
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

func TestCustomRule_Name(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName: "my_custom_rule",
		Pattern:  regexp.MustCompile(`panic\(`),
		Sev:      analyzer.SeverityHigh,
	}
	if rule.Name() != "my_custom_rule" {
		t.Errorf("expected my_custom_rule, got %q", rule.Name())
	}
}

func TestCustomRule_DetectsPattern(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName:    "no_panic",
		Pattern:     regexp.MustCompile(`panic\(`),
		Sev:         analyzer.SeverityHigh,
		Desc:        "avoid panic",
		Suggestion_: "return an error instead",
	}
	diff := makeFileDiff("foo.go",
		makeAddedLine(`panic("something went wrong")`, 8),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Line != 8 {
		t.Errorf("expected line 8, got %d", f.Line)
	}
	if f.EndLine != 8 {
		t.Errorf("expected EndLine 8, got %d", f.EndLine)
	}
	if f.Rule != "no_panic" {
		t.Errorf("expected rule no_panic, got %q", f.Rule)
	}
	if f.Severity != analyzer.SeverityHigh {
		t.Errorf("expected HIGH severity, got %v", f.Severity)
	}
	if f.Description != "avoid panic" {
		t.Errorf("expected description 'avoid panic', got %q", f.Description)
	}
	if f.Suggestion != "return an error instead" {
		t.Errorf("expected suggestion, got %q", f.Suggestion)
	}
}

func TestCustomRule_NoMatchOnNonPattern(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName: "no_panic",
		Pattern:  regexp.MustCompile(`panic\(`),
		Sev:      analyzer.SeverityHigh,
	}
	diff := makeFileDiff("foo.go",
		makeAddedLine("return errors.New(\"something went wrong\")", 3),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
}

func TestCustomRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName: "no_panic",
		Pattern:  regexp.MustCompile(`panic\(`),
		Sev:      analyzer.SeverityHigh,
	}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineRemoved, Content: `panic("old code")`, OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestCustomRule_SkipsContextLines(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName: "no_panic",
		Pattern:  regexp.MustCompile(`panic\(`),
		Sev:      analyzer.SeverityHigh,
	}
	diff := makeFileDiff("foo.go",
		git.DiffLine{Type: git.LineContext, Content: `panic("context")`, OldNum: 5, NewNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for context lines, got %d", len(findings))
	}
}

func TestCustomRule_MultipleMatches(t *testing.T) {
	t.Parallel()
	rule := &CustomRule{
		RuleName: "no_print",
		Pattern:  regexp.MustCompile(`fmt\.Print`),
		Sev:      analyzer.SeverityLow,
	}
	diff := makeFileDiff("foo.go",
		makeAddedLine(`fmt.Println("debug1")`, 1),
		makeAddedLine(`fmt.Printf("debug2 %v", x)`, 2),
		makeAddedLine("doRealWork()", 3),
		makeAddedLine(`fmt.Print("debug3")`, 4),
	)
	findings := rule.Check(diff)
	if len(findings) != 3 {
		t.Errorf("expected 3 findings, got %d", len(findings))
	}
}
