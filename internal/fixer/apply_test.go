package fixer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "source.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	return string(data)
}

func TestApply_SingleLineReplacement(t *testing.T) {
	t.Parallel()
	content := "line1\nline2\nline3\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         2,
		EndLine:      2,
		OriginalCode: "line2",
		FixCode:      "replaced_line2",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "replaced_line2") {
		t.Errorf("expected replaced_line2 in output, got:\n%s", got)
	}
	if strings.Contains(got, "\nline2\n") {
		t.Errorf("expected original 'line2' line to be removed, got:\n%s", got)
	}
}

func TestApply_MultiLineReplacement(t *testing.T) {
	t.Parallel()
	content := "a\nb\nc\nd\ne\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         2,
		EndLine:      4,
		OriginalCode: "b\nc\nd",
		FixCode:      "newb\nnewc",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "newb") {
		t.Errorf("expected newb in output, got:\n%s", got)
	}
	if !strings.Contains(got, "newc") {
		t.Errorf("expected newc in output, got:\n%s", got)
	}
}

func TestApply_FirstLine(t *testing.T) {
	t.Parallel()
	content := "first\nsecond\nthird\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         1,
		EndLine:      1,
		OriginalCode: "first",
		FixCode:      "new_first",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.HasPrefix(got, "new_first") {
		t.Errorf("expected new_first at start, got:\n%s", got)
	}
}

func TestApply_LastLine(t *testing.T) {
	t.Parallel()
	content := "first\nsecond\nlast"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         3,
		EndLine:      3,
		OriginalCode: "last",
		FixCode:      "new_last",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "new_last") {
		t.Errorf("expected new_last in output, got:\n%s", got)
	}
}

func TestApply_DeleteCode(t *testing.T) {
	t.Parallel()
	content := "keep\ndelete_me\nkeep2\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         2,
		EndLine:      2,
		OriginalCode: "delete_me",
		FixCode:      "",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if strings.Contains(got, "delete_me") {
		t.Errorf("expected delete_me to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "keep") {
		t.Errorf("expected keep to remain, got:\n%s", got)
	}
}

func TestApply_NoFixAvailable(t *testing.T) {
	t.Parallel()
	finding := analyzer.Finding{
		File:         "does_not_matter.go",
		Line:         1,
		FixCode:      "",
		OriginalCode: "",
	}
	err := Apply(finding)
	if err == nil {
		t.Error("expected error when no fix available")
	}
}

func TestApply_FileNotFound(t *testing.T) {
	t.Parallel()
	finding := analyzer.Finding{
		File:         "/nonexistent/path/file.go",
		Line:         1,
		OriginalCode: "something",
		FixCode:      "fix",
	}
	err := Apply(finding)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestApply_LineOutOfRange(t *testing.T) {
	t.Parallel()
	content := "only_one_line"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         100,
		EndLine:      100,
		OriginalCode: "something",
		FixCode:      "fix",
	}

	err := Apply(finding)
	if err == nil {
		t.Error("expected error for line out of range")
	}
}

func TestApply_LineZeroOutOfRange(t *testing.T) {
	t.Parallel()
	content := "line1\nline2\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         0,
		EndLine:      0,
		OriginalCode: "line1",
		FixCode:      "fix",
	}

	err := Apply(finding)
	if err == nil {
		t.Error("expected error for line 0")
	}
}

func TestApply_EndLineLessThanLine(t *testing.T) {
	t.Parallel()
	content := "a\nb\nc\n"
	path := writeTempFile(t, content)

	finding := analyzer.Finding{
		File:         path,
		Line:         2,
		EndLine:      1,
		OriginalCode: "b",
		FixCode:      "new_b",
	}

	if err := Apply(finding); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFile(t, path)
	if !strings.Contains(got, "new_b") {
		t.Errorf("expected new_b in output, got:\n%s", got)
	}
}
