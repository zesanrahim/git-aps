package rules

import "testing"

func TestTypeAssertRule_Name(t *testing.T) {
	t.Parallel()
	rule := &TypeAssertRule{}
	if rule.Name() != "type_assert" {
		t.Errorf("expected name type_assert, got %q", rule.Name())
	}
}

func TestTypeAssertRule_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		line      string
		wantCount int
	}{
		{
			name:      "unsafe single-value assertion",
			line:      `val := x.(string)`,
			wantCount: 1,
		},
		{
			name:      "safe comma-ok assertion",
			line:      `val, ok := x.(string)`,
			wantCount: 0,
		},
		{
			name:      "safe comma-ok with existing assign",
			line:      `val, ok = x.(string)`,
			wantCount: 0,
		},
		{
			name:      "type switch is safe",
			line:      `switch v := x.(type) {`,
			wantCount: 0,
		},
		{
			name:      "comment line skipped",
			line:      `// val := x.(string)`,
			wantCount: 0,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &TypeAssertRule{}
			diff := makeFileDiff("main.go", makeAddedLine(tc.line, 5))
			findings := rule.Check(diff)
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d for line %q", len(findings), tc.wantCount, tc.line)
			}
		})
	}
}

func TestTypeAssertRule_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &TypeAssertRule{}
	diff := makeFileDiff("main.ts",
		makeAddedLine(`val := x.(string)`, 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-Go file, got %d", len(findings))
	}
}

func TestTypeAssertRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &TypeAssertRule{}
	diff := makeFileDiff("handler.go",
		makeAddedLine(`result := iface.(MyType)`, 99),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "handler.go" {
		t.Errorf("expected file handler.go, got %q", f.File)
	}
	if f.Line != 99 {
		t.Errorf("expected line 99, got %d", f.Line)
	}
	if f.Rule != "type_assert" {
		t.Errorf("expected rule type_assert, got %q", f.Rule)
	}
}
