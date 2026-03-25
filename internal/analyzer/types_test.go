package analyzer

import (
	"testing"
)

func TestSeverityString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityLow, "LOW"},
		{SeverityMedium, "MED"},
		{SeverityHigh, "HIGH"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := tc.sev.String(); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  Severity
	}{
		{"high", SeverityHigh},
		{"HIGH", SeverityHigh},
		{"High", SeverityHigh},
		{"  high  ", SeverityHigh},
		{"med", SeverityMedium},
		{"MED", SeverityMedium},
		{"medium", SeverityMedium},
		{"MEDIUM", SeverityMedium},
		{"low", SeverityLow},
		{"LOW", SeverityLow},
		{"", SeverityLow},
		{"unknown", SeverityLow},
		{"critical", SeverityLow},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			if got := ParseSeverity(tc.input); got != tc.want {
				t.Errorf("ParseSeverity(%q): want %v, got %v", tc.input, tc.want, got)
			}
		})
	}
}

func TestFilterBySeverity(t *testing.T) {
	t.Parallel()
	findings := []Finding{
		{Rule: "r1", Severity: SeverityLow},
		{Rule: "r2", Severity: SeverityMedium},
		{Rule: "r3", Severity: SeverityHigh},
		{Rule: "r4", Severity: SeverityHigh},
	}

	tests := []struct {
		name        string
		minSeverity Severity
		wantCount   int
	}{
		{"low threshold returns all", SeverityLow, 4},
		{"medium threshold returns med+high", SeverityMedium, 3},
		{"high threshold returns only high", SeverityHigh, 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FilterBySeverity(findings, tc.minSeverity)
			if len(got) != tc.wantCount {
				t.Errorf("FilterBySeverity min=%v: want %d findings, got %d", tc.minSeverity, tc.wantCount, len(got))
			}
		})
	}
}

func TestFilterBySeverity_EmptySlice(t *testing.T) {
	t.Parallel()
	got := FilterBySeverity(nil, SeverityMedium)
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestFilterBySeverity_PreservesOrder(t *testing.T) {
	t.Parallel()
	findings := []Finding{
		{Rule: "first", Severity: SeverityHigh},
		{Rule: "second", Severity: SeverityHigh},
	}
	got := FilterBySeverity(findings, SeverityHigh)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Rule != "first" || got[1].Rule != "second" {
		t.Error("order not preserved")
	}
}
