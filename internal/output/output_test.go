package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func TestPrintJSON_EmptyFindings(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := PrintJSON(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestPrintJSON_SingleFinding(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{
			File:        "foo.go",
			Line:        10,
			EndLine:     12,
			Severity:    analyzer.SeverityHigh,
			Rule:        "null_deref",
			Description: "Possible nil pointer",
			Suggestion:  "Add nil check",
			OriginalCode: "x.Do()",
			FixCode:     "if x != nil { x.Do() }",
		},
	}
	var buf bytes.Buffer
	err := PrintJSON(findings, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result))
	}

	item := result[0]
	if item["file"] != "foo.go" {
		t.Errorf("expected file foo.go, got %v", item["file"])
	}
	if item["severity"] != "HIGH" {
		t.Errorf("expected severity HIGH, got %v", item["severity"])
	}
	if item["rule"] != "null_deref" {
		t.Errorf("expected rule null_deref, got %v", item["rule"])
	}
}

func TestPrintJSON_SeverityStrings(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{Rule: "r1", Severity: analyzer.SeverityHigh, Description: "d1"},
		{Rule: "r2", Severity: analyzer.SeverityMedium, Description: "d2"},
		{Rule: "r3", Severity: analyzer.SeverityLow, Description: "d3"},
	}
	var buf bytes.Buffer
	if err := PrintJSON(findings, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result[0]["severity"] != "HIGH" {
		t.Errorf("expected HIGH, got %v", result[0]["severity"])
	}
	if result[1]["severity"] != "MED" {
		t.Errorf("expected MED, got %v", result[1]["severity"])
	}
	if result[2]["severity"] != "LOW" {
		t.Errorf("expected LOW, got %v", result[2]["severity"])
	}
}

func TestPrintJSON_MultipleFindings(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{File: "a.go", Line: 1, Rule: "r1", Severity: analyzer.SeverityHigh, Description: "d1"},
		{File: "b.go", Line: 2, Rule: "r2", Severity: analyzer.SeverityLow, Description: "d2"},
		{File: "c.go", Line: 3, Rule: "r3", Severity: analyzer.SeverityMedium, Description: "d3"},
	}
	var buf bytes.Buffer
	if err := PrintJSON(findings, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 findings, got %d", len(result))
	}
}

func TestPrintText_Empty(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	PrintText(nil, &buf)
	out := buf.String()
	if !strings.Contains(out, "0 findings") {
		t.Errorf("expected '0 findings' in output, got:\n%s", out)
	}
}

func TestPrintText_SingleFinding(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{
			File:        "foo.go",
			Line:        42,
			Severity:    analyzer.SeverityHigh,
			Rule:        "null_deref",
			Description: "Possible nil pointer",
		},
	}
	var buf bytes.Buffer
	PrintText(findings, &buf)
	out := buf.String()

	if !strings.Contains(out, "foo.go") {
		t.Errorf("expected foo.go in output, got:\n%s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected line 42 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "HIGH") {
		t.Errorf("expected HIGH in output, got:\n%s", out)
	}
	if !strings.Contains(out, "null_deref") {
		t.Errorf("expected rule in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1 findings") {
		t.Errorf("expected '1 findings' in summary, got:\n%s", out)
	}
}

func TestPrintText_SummaryCounts(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{Rule: "r1", Severity: analyzer.SeverityHigh, Description: "h1"},
		{Rule: "r2", Severity: analyzer.SeverityHigh, Description: "h2"},
		{Rule: "r3", Severity: analyzer.SeverityMedium, Description: "m1"},
		{Rule: "r4", Severity: analyzer.SeverityLow, Description: "l1"},
		{Rule: "r5", Severity: analyzer.SeverityLow, Description: "l2"},
		{Rule: "r6", Severity: analyzer.SeverityLow, Description: "l3"},
	}
	var buf bytes.Buffer
	PrintText(findings, &buf)
	out := buf.String()

	if !strings.Contains(out, "6 findings") {
		t.Errorf("expected '6 findings', got:\n%s", out)
	}
	if !strings.Contains(out, "2 high") {
		t.Errorf("expected '2 high', got:\n%s", out)
	}
	if !strings.Contains(out, "1 med") {
		t.Errorf("expected '1 med', got:\n%s", out)
	}
	if !strings.Contains(out, "3 low") {
		t.Errorf("expected '3 low', got:\n%s", out)
	}
}

func TestPrintText_Format(t *testing.T) {
	t.Parallel()
	findings := []analyzer.Finding{
		{File: "main.go", Line: 5, Severity: analyzer.SeverityMedium, Rule: "deep_nesting", Description: "Too nested"},
	}
	var buf bytes.Buffer
	PrintText(findings, &buf)
	out := buf.String()

	if !strings.Contains(out, "[MED]") {
		t.Errorf("expected [MED] format, got:\n%s", out)
	}
	if !strings.Contains(out, "main.go:5") {
		t.Errorf("expected main.go:5 format, got:\n%s", out)
	}
}
