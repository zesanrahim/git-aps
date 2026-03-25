package rules

import "testing"

func TestGoroutineCtxRule_Name(t *testing.T) {
	t.Parallel()
	rule := &GoroutineCtxRule{}
	if rule.Name() != "goroutine_ctx" {
		t.Errorf("expected name goroutine_ctx, got %q", rule.Name())
	}
}

func TestGoroutineCtxRule_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		line      string
		wantCount int
	}{
		{
			name:      "anon goroutine without context",
			line:      `go func() { doWork() }()`,
			wantCount: 1,
		},
		{
			name:      "named goroutine without ctx",
			line:      `go processItem(item)`,
			wantCount: 1,
		},
		{
			name:      "anon goroutine with context.Context param",
			line:      `go func(ctx context.Context) { doWork(ctx) }(ctx)`,
			wantCount: 0,
		},
		{
			name:      "anon goroutine passing ctx",
			line:      `go func(ctx context.Context) {}(ctx)`,
			wantCount: 0,
		},
		{
			name:      "named goroutine with ctx",
			line:      `go process(ctx, item)`,
			wantCount: 0,
		},
		{
			name:      "comment line skipped",
			line:      `// go func() { doWork() }()`,
			wantCount: 0,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &GoroutineCtxRule{}
			diff := makeFileDiff("main.go", makeAddedLine(tc.line, 5))
			findings := rule.Check(diff)
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d for line %q", len(findings), tc.wantCount, tc.line)
			}
		})
	}
}

func TestGoroutineCtxRule_SkipsNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &GoroutineCtxRule{}
	diff := makeFileDiff("worker.py",
		makeAddedLine(`go func() { doWork() }()`, 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for non-Go file, got %d", len(findings))
	}
}

func TestGoroutineCtxRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &GoroutineCtxRule{}
	diff := makeFileDiff("worker.go",
		makeAddedLine(`go func() { doWork() }()`, 77),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "worker.go" {
		t.Errorf("expected file worker.go, got %q", f.File)
	}
	if f.Line != 77 {
		t.Errorf("expected line 77, got %d", f.Line)
	}
	if f.Rule != "goroutine_ctx" {
		t.Errorf("expected rule goroutine_ctx, got %q", f.Rule)
	}
}
