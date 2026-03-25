package analyzer

import (
	"testing"
)

func TestMergeFindings_Empty(t *testing.T) {
	t.Parallel()
	got := MergeFindings()
	if len(got) != 0 {
		t.Errorf("expected 0 findings, got %d", len(got))
	}
}

func TestMergeFindings_SingleSource(t *testing.T) {
	t.Parallel()
	source := []Finding{
		{File: "a.go", Rule: "r1", Line: 1, Severity: SeverityLow},
		{File: "a.go", Rule: "r2", Line: 2, Severity: SeverityHigh},
	}
	got := MergeFindings(source)
	if len(got) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(got))
	}
}

func TestMergeFindings_Deduplication(t *testing.T) {
	t.Parallel()
	a := []Finding{
		{File: "a.go", Rule: "r1", Line: 10, Severity: SeverityHigh},
	}
	b := []Finding{
		{File: "a.go", Rule: "r1", Line: 10, Severity: SeverityHigh},
	}
	got := MergeFindings(a, b)
	if len(got) != 1 {
		t.Errorf("expected 1 finding after dedup, got %d", len(got))
	}
}

func TestMergeFindings_NoDedupDifferentFiles(t *testing.T) {
	t.Parallel()
	a := []Finding{
		{File: "a.go", Rule: "r1", Line: 10, Severity: SeverityHigh},
	}
	b := []Finding{
		{File: "b.go", Rule: "r1", Line: 10, Severity: SeverityHigh},
	}
	got := MergeFindings(a, b)
	if len(got) != 2 {
		t.Errorf("expected 2 findings, got %d", len(got))
	}
}

func TestMergeFindings_NoDedupDifferentLines(t *testing.T) {
	t.Parallel()
	a := []Finding{
		{File: "a.go", Rule: "r1", Line: 10, Severity: SeverityHigh},
	}
	b := []Finding{
		{File: "a.go", Rule: "r1", Line: 11, Severity: SeverityHigh},
	}
	got := MergeFindings(a, b)
	if len(got) != 2 {
		t.Errorf("expected 2 findings, got %d", len(got))
	}
}

func TestMergeFindings_SortsBySeverityDesc(t *testing.T) {
	t.Parallel()
	source := []Finding{
		{File: "a.go", Rule: "r1", Line: 1, Severity: SeverityLow},
		{File: "a.go", Rule: "r2", Line: 2, Severity: SeverityHigh},
		{File: "a.go", Rule: "r3", Line: 3, Severity: SeverityMedium},
	}
	got := MergeFindings(source)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	if got[0].Severity != SeverityHigh {
		t.Errorf("first finding should be HIGH, got %v", got[0].Severity)
	}
	if got[1].Severity != SeverityMedium {
		t.Errorf("second finding should be MED, got %v", got[1].Severity)
	}
	if got[2].Severity != SeverityLow {
		t.Errorf("third finding should be LOW, got %v", got[2].Severity)
	}
}

func TestMergeFindings_SortsByFileWithinSeverity(t *testing.T) {
	t.Parallel()
	source := []Finding{
		{File: "z.go", Rule: "r1", Line: 1, Severity: SeverityHigh},
		{File: "a.go", Rule: "r2", Line: 1, Severity: SeverityHigh},
	}
	got := MergeFindings(source)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].File != "a.go" {
		t.Errorf("expected a.go first, got %q", got[0].File)
	}
}

func TestMergeFindings_SortsByLineWithinFileAndSeverity(t *testing.T) {
	t.Parallel()
	source := []Finding{
		{File: "a.go", Rule: "r1", Line: 20, Severity: SeverityHigh},
		{File: "a.go", Rule: "r2", Line: 5, Severity: SeverityHigh},
	}
	got := MergeFindings(source)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Line != 5 {
		t.Errorf("expected line 5 first, got %d", got[0].Line)
	}
}

func TestMergeFindings_MultipleSources(t *testing.T) {
	t.Parallel()
	a := []Finding{
		{File: "a.go", Rule: "static", Line: 1, Severity: SeverityLow},
	}
	b := []Finding{
		{File: "a.go", Rule: "ai_rule", Line: 5, Severity: SeverityHigh},
	}
	c := []Finding{
		{File: "b.go", Rule: "static", Line: 2, Severity: SeverityMedium},
	}
	got := MergeFindings(a, b, c)
	if len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}
