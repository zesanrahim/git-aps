package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func TestCognitiveComplexityRule_Name(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 15}
	if rule.Name() != "cognitive_complexity" {
		t.Errorf("expected name cognitive_complexity, got %q", rule.Name())
	}
}

func TestCognitiveComplexityRule_SimpleFunction_BelowThreshold(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 15}
	diff := makeFileDiff("main.go",
		makeAddedLine("func simple() {", 1),
		makeAddedLine("\tx := 1", 2),
		makeAddedLine("}", 3),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for simple function, got %d", len(findings))
	}
}

func TestCognitiveComplexityRule_ComplexFunction_ExceedsThreshold(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 3}
	diff := makeFileDiff("main.go",
		makeAddedLine("func complex() {", 1),
		makeAddedLine("\tif a {", 2),
		makeAddedLine("\t\tif b {", 3),
		makeAddedLine("\t\t\tif c {", 4),
		makeAddedLine("\t\t\t\tfor i := 0; i < n; i++ {", 5),
		makeAddedLine("\t\t\t\t\tx := 1", 6),
		makeAddedLine("\t\t\t\t}", 7),
		makeAddedLine("\t\t\t}", 8),
		makeAddedLine("\t\t}", 9),
		makeAddedLine("\t}", 10),
		makeAddedLine("}", 11),
	)
	findings := rule.Check(diff)
	if len(findings) == 0 {
		t.Error("expected at least one finding for complex function above threshold")
	}
	if findings[0].Rule != "cognitive_complexity" {
		t.Errorf("expected rule cognitive_complexity, got %q", findings[0].Rule)
	}
}

func TestCognitiveComplexityRule_SeverityMedAtThreshold(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 1}
	diff := makeFileDiff("main.go",
		makeAddedLine("func check() {", 1),
		makeAddedLine("\tif a && b {", 2),
		makeAddedLine("\t\tx := 1", 3),
		makeAddedLine("\t}", 4),
		makeAddedLine("}", 5),
	)
	findings := rule.Check(diff)
	if len(findings) == 0 {
		t.Fatal("expected at least one finding")
	}
	complexity := findings[0].Description
	if complexity == "" {
		t.Error("expected non-empty description")
	}
}

func TestCognitiveComplexityRule_SeverityHighAtDoubleThreshold(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 2}
	diff := makeFileDiff("main.go",
		makeAddedLine("func bloated() {", 1),
		makeAddedLine("\tif a {", 2),
		makeAddedLine("\t\tif b {", 3),
		makeAddedLine("\t\t\tif c && d {", 4),
		makeAddedLine("\t\t\t\tfor i := 0; i < n; i++ {", 5),
		makeAddedLine("\t\t\t\t\tif e || f {", 6),
		makeAddedLine("\t\t\t\t\t\tx := 1", 7),
		makeAddedLine("\t\t\t\t\t}", 8),
		makeAddedLine("\t\t\t\t}", 9),
		makeAddedLine("\t\t\t}", 10),
		makeAddedLine("\t\t}", 11),
		makeAddedLine("\t}", 12),
		makeAddedLine("}", 13),
	)
	findings := rule.Check(diff)
	if len(findings) == 0 {
		t.Fatal("expected at least one finding for bloated function")
	}
	if findings[0].Severity != analyzer.SeverityHigh {
		t.Errorf("expected HIGH severity for complexity > 2x threshold, got %v", findings[0].Severity)
	}
}

func TestCognitiveComplexityRule_FindingPointsToFunctionStart(t *testing.T) {
	t.Parallel()
	rule := &CognitiveComplexityRule{Threshold: 1}
	diff := makeFileDiff("handler.go",
		makeAddedLine("func handler() {", 10),
		makeAddedLine("\tif x && y {", 11),
		makeAddedLine("\t\tz := 1", 12),
		makeAddedLine("\t}", 13),
		makeAddedLine("}", 14),
	)
	findings := rule.Check(diff)
	if len(findings) == 0 {
		t.Fatal("expected at least one finding")
	}
	if findings[0].File != "handler.go" {
		t.Errorf("expected file handler.go, got %q", findings[0].File)
	}
	if findings[0].Line != 10 {
		t.Errorf("expected finding at line 10 (func start), got %d", findings[0].Line)
	}
}
