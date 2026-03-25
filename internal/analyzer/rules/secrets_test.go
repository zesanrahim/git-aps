package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestSecretsRule_Name(t *testing.T) {
	t.Parallel()
	rule := &SecretsRule{}
	if rule.Name() != "secrets" {
		t.Errorf("expected name secrets, got %q", rule.Name())
	}
}

func TestSecretsRule_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		path      string
		line      string
		wantCount int
	}{
		{
			name:      "api_key with equals assignment",
			path:      "config.go",
			line:      `api_key = "sk-abc123def456ghi789jklmnop"`,
			wantCount: 1,
		},
		{
			name:      "password with equals assignment",
			path:      "auth.go",
			line:      `password = "super_secret_password_123"`,
			wantCount: 1,
		},
		{
			name:      "secret colon assignment",
			path:      "app.go",
			line:      `secret: "my-very-long-secret-value-here"`,
			wantCount: 1,
		},
		{
			name:      "aws access key id",
			path:      "cloud.go",
			line:      `AKIAIOSFODNN7EXAMPLE`,
			wantCount: 1,
		},
		{
			name:      "github pat token",
			path:      "ci.go",
			line:      `github_pat_11ABC12345678901234567890123456789`,
			wantCount: 1,
		},
		{
			name:      "env var lookup is safe",
			path:      "config.go",
			line:      `apiKey := os.Getenv("API_KEY")`,
			wantCount: 0,
		},
		{
			name:      "config.Get is safe",
			path:      "config.go",
			line:      `apiKey := config.Get("api_key")`,
			wantCount: 0,
		},
		{
			name:      "short value not flagged",
			path:      "config.go",
			line:      `password = "short"`,
			wantCount: 0,
		},
		{
			name:      "test file skipped",
			path:      "config_test.go",
			line:      `api_key = "sk-abc123def456ghi789jklmnop"`,
			wantCount: 0,
		},
		{
			name:      "comment line skipped",
			path:      "config.go",
			line:      `// password = "super_secret_password_123"`,
			wantCount: 0,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &SecretsRule{}
			diff := makeFileDiff(tc.path, makeAddedLine(tc.line, 5))
			findings := rule.Check(diff)
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d for line %q", len(findings), tc.wantCount, tc.line)
			}
		})
	}
}

func TestSecretsRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &SecretsRule{}
	diff := makeFileDiff("config.go",
		git.DiffLine{Type: git.LineRemoved, Content: `api_key = "sk-abc123def456ghi789jklmnop"`, OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestSecretsRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &SecretsRule{}
	diff := makeFileDiff("app.go",
		makeAddedLine(`api_key = "sk-abc123def456ghi789jklmnop"`, 15),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "app.go" {
		t.Errorf("expected file app.go, got %q", f.File)
	}
	if f.Line != 15 {
		t.Errorf("expected line 15, got %d", f.Line)
	}
	if f.Rule != "secrets" {
		t.Errorf("expected rule secrets, got %q", f.Rule)
	}
}
