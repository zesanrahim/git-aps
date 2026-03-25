package ai

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func TestParseResponse_NoIssues(t *testing.T) {
	t.Parallel()
	findings, err := parseResponse("foo.go", "NO_ISSUES")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseResponse_NoIssuesWithWhitespace(t *testing.T) {
	t.Parallel()
	findings, err := parseResponse("foo.go", "  NO_ISSUES  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for whitespace-padded NO_ISSUES, got %d", len(findings))
	}
}

func TestParseResponse_Empty(t *testing.T) {
	t.Parallel()
	findings, err := parseResponse("foo.go", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty response, got %d", len(findings))
	}
}

func TestParseResponse_SingleFinding(t *testing.T) {
	t.Parallel()
	response := `FINDING
LINE: 42
END_LINE: 45
SEVERITY: HIGH
RULE: null_deref
DESCRIPTION: Possible nil pointer dereference
SUGGESTION: Add nil check before use
FIX:
if x != nil {
    x.DoThing()
}
END_FIX`

	findings, err := parseResponse("foo.go", response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "foo.go" {
		t.Errorf("expected file foo.go, got %q", f.File)
	}
	if f.Line != 42 {
		t.Errorf("expected line 42, got %d", f.Line)
	}
	if f.EndLine != 45 {
		t.Errorf("expected end line 45, got %d", f.EndLine)
	}
	if f.Severity != analyzer.SeverityHigh {
		t.Errorf("expected HIGH severity, got %v", f.Severity)
	}
	if f.Rule != "null_deref" {
		t.Errorf("expected rule null_deref, got %q", f.Rule)
	}
	if f.Description != "Possible nil pointer dereference" {
		t.Errorf("unexpected description: %q", f.Description)
	}
	if f.Suggestion != "Add nil check before use" {
		t.Errorf("unexpected suggestion: %q", f.Suggestion)
	}
	if f.FixCode == "" {
		t.Error("expected non-empty FixCode")
	}
}

func TestParseResponse_MultipleFindings(t *testing.T) {
	t.Parallel()
	response := `FINDING
LINE: 10
SEVERITY: HIGH
RULE: race_condition
DESCRIPTION: Concurrent map write without lock
SUGGESTION: Use sync.Mutex
FIX:
mu.Lock()
defer mu.Unlock()
END_FIX
FINDING
LINE: 25
SEVERITY: LOW
RULE: naming
DESCRIPTION: Variable name too short
SUGGESTION: Use descriptive name
FIX:
END_FIX`

	findings, err := parseResponse("bar.go", response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Line != 10 {
		t.Errorf("first finding: expected line 10, got %d", findings[0].Line)
	}
	if findings[1].Line != 25 {
		t.Errorf("second finding: expected line 25, got %d", findings[1].Line)
	}
}

func TestParseResponse_SkipsBlocksWithNoDescription(t *testing.T) {
	t.Parallel()
	response := `FINDING
LINE: 10
SEVERITY: HIGH
RULE: something`

	findings, err := parseResponse("foo.go", response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings without description, got %d", len(findings))
	}
}

func TestParseResponse_SeverityParsing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		severityStr string
		want        analyzer.Severity
	}{
		{"HIGH", analyzer.SeverityHigh},
		{"MED", analyzer.SeverityMedium},
		{"MEDIUM", analyzer.SeverityMedium},
		{"LOW", analyzer.SeverityLow},
		{"unknown", analyzer.SeverityLow},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.severityStr, func(t *testing.T) {
			t.Parallel()
			response := "FINDING\nLINE: 1\nSEVERITY: " + tc.severityStr + "\nRULE: r\nDESCRIPTION: desc\nFIX:\nEND_FIX"
			findings, err := parseResponse("f.go", response)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(findings))
			}
			if findings[0].Severity != tc.want {
				t.Errorf("severity %q: want %v, got %v", tc.severityStr, tc.want, findings[0].Severity)
			}
		})
	}
}

func TestParseResponse_FixCodeWithCodeFences(t *testing.T) {
	t.Parallel()
	response := `FINDING
LINE: 5
SEVERITY: MED
RULE: style
DESCRIPTION: Use standard library
SUGGESTION: replace with stdlib
FIX:
` + "```go" + `
result, err := strconv.Atoi(s)
if err != nil {
    return err
}
` + "```" + `
END_FIX`

	findings, err := parseResponse("foo.go", response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].FixCode == "" {
		t.Error("expected non-empty FixCode")
	}
}

func TestStripCodeFences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: "code here",
			want:  "code here",
		},
		{
			name:  "with go fence",
			input: "```go\nfunc foo() {}\n```",
			want:  "func foo() {}",
		},
		{
			name:  "with plain fence",
			input: "```\nsome code\n```",
			want:  "some code",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "only backticks",
			input: "```\n```",
			want:  "",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := stripCodeFences(tc.input)
			if got != tc.want {
				t.Errorf("stripCodeFences(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  analyzer.Severity
	}{
		{"HIGH", analyzer.SeverityHigh},
		{"high", analyzer.SeverityHigh},
		{"MED", analyzer.SeverityMedium},
		{"MEDIUM", analyzer.SeverityMedium},
		{"LOW", analyzer.SeverityLow},
		{"", analyzer.SeverityLow},
		{"other", analyzer.SeverityLow},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := parseSeverity(tc.input)
			if got != tc.want {
				t.Errorf("parseSeverity(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
