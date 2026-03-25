package git

import (
	"testing"
)

func TestParseDiff_Empty(t *testing.T) {
	t.Parallel()
	files, err := ParseDiff("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestParseDiff_SingleAddedFile(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/foo.go b/foo.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/foo.go
@@ -0,0 +1,3 @@
+package main
+
+func main() {}
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Path != "foo.go" {
		t.Errorf("expected path foo.go, got %q", f.Path)
	}
	if !f.IsNew {
		t.Error("expected IsNew=true")
	}
	if f.Deleted {
		t.Error("expected Deleted=false")
	}
}

func TestParseDiff_DeletedFile(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package main
-
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].Deleted {
		t.Error("expected Deleted=true")
	}
}

func TestParseDiff_RenamedFile(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/old.go b/new.go
similarity index 90%
rename from old.go
rename to new.go
index 1234567..abcdefg 100644
--- a/old.go
+++ b/new.go
@@ -1,3 +1,3 @@
 package main
-// old comment
+// new comment
 func main() {}
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.Path != "new.go" {
		t.Errorf("expected path new.go, got %q", f.Path)
	}
	if f.OldPath != "old.go" {
		t.Errorf("expected OldPath old.go, got %q", f.OldPath)
	}
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/a.go b/a.go
index 1111111..2222222 100644
--- a/a.go
+++ b/a.go
@@ -1,2 +1,3 @@
 package main
+
+var x = 1
diff --git a/b.go b/b.go
index 3333333..4444444 100644
--- a/b.go
+++ b/b.go
@@ -5,3 +5,4 @@
 func foo() {
+	bar()
 }
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "a.go" {
		t.Errorf("expected a.go, got %q", files[0].Path)
	}
	if files[1].Path != "b.go" {
		t.Errorf("expected b.go, got %q", files[1].Path)
	}
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/multi.go b/multi.go
index 1111111..2222222 100644
--- a/multi.go
+++ b/multi.go
@@ -1,3 +1,4 @@
 package main
+
+import "fmt"

@@ -20,3 +21,4 @@
 func bar() {
+	fmt.Println("bar")
 }
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if len(files[0].Hunks) != 2 {
		t.Errorf("expected 2 hunks, got %d", len(files[0].Hunks))
	}
}

func TestParseDiff_LineNumberTracking(t *testing.T) {
	t.Parallel()
	raw := `diff --git a/nums.go b/nums.go
index 1111111..2222222 100644
--- a/nums.go
+++ b/nums.go
@@ -5,6 +5,8 @@
 line5
 line6
+added7
+added8
 line7
 line8
`
	files, err := ParseDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no files parsed")
	}
	hunk := files[0].Hunks[0]

	var added []DiffLine
	for _, l := range hunk.Lines {
		if l.Type == LineAdded {
			added = append(added, l)
		}
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 added lines, got %d", len(added))
	}
	if added[0].NewNum != 7 {
		t.Errorf("expected first added line at 7, got %d", added[0].NewNum)
	}
	if added[1].NewNum != 8 {
		t.Errorf("expected second added line at 8, got %d", added[1].NewNum)
	}
}

func TestParseHunk_Header(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		lines      []string
		wantOld    int
		wantNew    int
		wantErrStr string
	}{
		{
			name:    "standard header",
			lines:   []string{"@@ -10,5 +20,7 @@", " context"},
			wantOld: 10,
			wantNew: 20,
		},
		{
			name:    "no counts",
			lines:   []string{"@@ -1 +1 @@", "+added"},
			wantOld: 1,
			wantNew: 1,
		},
		{
			name:       "empty lines",
			lines:      []string{},
			wantErrStr: "empty hunk",
		},
		{
			name:       "invalid header",
			lines:      []string{"not a hunk header"},
			wantErrStr: "invalid hunk header",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h, err := parseHunk(tc.lines)
			if tc.wantErrStr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if h.OldStart != tc.wantOld {
				t.Errorf("OldStart: want %d, got %d", tc.wantOld, h.OldStart)
			}
			if h.NewStart != tc.wantNew {
				t.Errorf("NewStart: want %d, got %d", tc.wantNew, h.NewStart)
			}
		})
	}
}

func TestParseHunk_LineTypes(t *testing.T) {
	t.Parallel()
	lines := []string{
		"@@ -1,3 +1,3 @@",
		" context line",
		"-removed line",
		"+added line",
	}
	h, err := parseHunk(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(h.Lines))
	}
	if h.Lines[0].Type != LineContext {
		t.Errorf("line 0: want LineContext, got %v", h.Lines[0].Type)
	}
	if h.Lines[1].Type != LineRemoved {
		t.Errorf("line 1: want LineRemoved, got %v", h.Lines[1].Type)
	}
	if h.Lines[2].Type != LineAdded {
		t.Errorf("line 2: want LineAdded, got %v", h.Lines[2].Type)
	}
}

func TestParseHunk_EmptyContextLine(t *testing.T) {
	t.Parallel()
	lines := []string{
		"@@ -1,3 +1,3 @@",
		"+first",
		"",
		"+third",
	}
	h, err := parseHunk(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(h.Lines))
	}
	if h.Lines[1].Type != LineContext {
		t.Errorf("blank line should be LineContext, got %v", h.Lines[1].Type)
	}
}

func TestHunkNewContent(t *testing.T) {
	t.Parallel()
	h := Hunk{
		Lines: []DiffLine{
			{Type: LineContext, Content: "ctx"},
			{Type: LineRemoved, Content: "old"},
			{Type: LineAdded, Content: "new"},
		},
	}
	got := h.NewContent()
	want := "ctx\nnew\n"
	if got != want {
		t.Errorf("NewContent: want %q, got %q", want, got)
	}
}

func TestDiffArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mode string
		want []string
	}{
		{"staged", []string{"diff", "--staged"}},
		{"unstaged", []string{"diff"}},
		{"head", []string{"diff", "HEAD~1"}},
		{"branch", []string{"diff", "main...HEAD"}},
		{"", []string{"diff", "--staged"}},
		{"unknown", []string{"diff", "--staged"}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.mode, func(t *testing.T) {
			t.Parallel()
			got := diffArgs(tc.mode)
			if len(got) != len(tc.want) {
				t.Fatalf("mode %q: want %v, got %v", tc.mode, tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("mode %q arg[%d]: want %q, got %q", tc.mode, i, tc.want[i], got[i])
				}
			}
		})
	}
}
