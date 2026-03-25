package rules

import (
	"testing"

	"github.com/zesanrahim/git-aps/internal/git"
)

func TestSQLInjectionRule_Name(t *testing.T) {
	t.Parallel()
	rule := &SQLInjectionRule{}
	if rule.Name() != "sql_injection" {
		t.Errorf("expected name sql_injection, got %q", rule.Name())
	}
}

func TestSQLInjectionRule_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		line      string
		wantCount int
	}{
		{
			name:      "string concat in query",
			line:      `query := "SELECT * FROM users WHERE id = " + userID`,
			wantCount: 1,
		},
		{
			name:      "sprintf with query",
			line:      `db.Query(fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))`,
			wantCount: 1,
		},
		{
			name:      "parameterized query with question mark",
			line:      `db.Query("SELECT * FROM users WHERE id = ?", userID)`,
			wantCount: 0,
		},
		{
			name:      "parameterized query with dollar placeholder",
			line:      `db.Query("SELECT * FROM users WHERE id = $1", userID)`,
			wantCount: 0,
		},
		{
			name:      "comment line skipped",
			line:      `// query := "SELECT * FROM users WHERE id = " + userID`,
			wantCount: 0,
		},
		{
			name:      "insert with concat",
			line:      `q := "INSERT INTO logs VALUES ('" + val + "')"`,
			wantCount: 1,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rule := &SQLInjectionRule{}
			diff := makeFileDiff("db.go", makeAddedLine(tc.line, 5))
			findings := rule.Check(diff)
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d for line %q", len(findings), tc.wantCount, tc.line)
			}
		})
	}
}

func TestSQLInjectionRule_SkipsRemovedLines(t *testing.T) {
	t.Parallel()
	rule := &SQLInjectionRule{}
	diff := makeFileDiff("db.go",
		git.DiffLine{Type: git.LineRemoved, Content: `query := "SELECT * FROM users WHERE id = " + id`, OldNum: 5},
	)
	findings := rule.Check(diff)
	if len(findings) != 0 {
		t.Errorf("expected no findings for removed lines, got %d", len(findings))
	}
}

func TestSQLInjectionRule_WorksOnNonGoFiles(t *testing.T) {
	t.Parallel()
	rule := &SQLInjectionRule{}
	diff := makeFileDiff("query.py",
		makeAddedLine(`query = "SELECT * FROM users WHERE id = " + user_id`, 5),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Errorf("expected 1 finding for Python file with SQL injection, got %d", len(findings))
	}
}

func TestSQLInjectionRule_FindingMetadata(t *testing.T) {
	t.Parallel()
	rule := &SQLInjectionRule{}
	diff := makeFileDiff("repo.go",
		makeAddedLine(`q := "SELECT * FROM users WHERE id = " + id`, 30),
	)
	findings := rule.Check(diff)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.File != "repo.go" {
		t.Errorf("expected file repo.go, got %q", f.File)
	}
	if f.Line != 30 {
		t.Errorf("expected line 30, got %d", f.Line)
	}
	if f.Rule != "sql_injection" {
		t.Errorf("expected rule sql_injection, got %q", f.Rule)
	}
}
