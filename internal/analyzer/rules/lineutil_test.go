package rules

import "testing"

func TestIsComment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		lang Language
		want bool
	}{
		{"go line comment", "// this is a comment", LangGo, true},
		{"go block comment start", "/* block", LangGo, true},
		{"go doc comment star", "* doc", LangGo, true},
		{"go code line", "x := 5", LangGo, false},
		{"go leading whitespace comment", "    // indented", LangGo, true},
		{"java line comment", "// java comment", LangJava, true},
		{"java block comment", "/* block */", LangJava, true},
		{"js line comment", "// js comment", LangJavaScript, true},
		{"ts line comment", "// ts comment", LangTypeScript, true},
		{"rust line comment", "// rust comment", LangRust, true},
		{"python hash comment", "# python comment", LangPython, true},
		{"python code", "x = 5", LangPython, false},
		{"python leading whitespace", "    # indented", LangPython, true},
		{"unknown slashes", "// unknown", LangUnknown, true},
		{"unknown hash", "# unknown", LangUnknown, true},
		{"unknown block", "/* unknown */", LangUnknown, true},
		{"unknown code", "some code", LangUnknown, false},
		{"empty line", "", LangGo, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsComment(tc.line, tc.lang)
			if got != tc.want {
				t.Errorf("IsComment(%q, %v) = %v, want %v", tc.line, tc.lang, got, tc.want)
			}
		})
	}
}

func TestIsBlankOrComment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		lang Language
		want bool
	}{
		{"blank line", "", LangGo, true},
		{"whitespace only", "   \t  ", LangGo, true},
		{"go comment", "// comment", LangGo, true},
		{"go code", "x := 5", LangGo, false},
		{"python comment", "# comment", LangPython, true},
		{"python code", "x = 5", LangPython, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsBlankOrComment(tc.line, tc.lang)
			if got != tc.want {
				t.Errorf("IsBlankOrComment(%q, %v) = %v, want %v", tc.line, tc.lang, got, tc.want)
			}
		})
	}
}

func TestStripComment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		lang Language
		want string
	}{
		{"go inline comment", `x := 5 // assign`, LangGo, "x := 5 "},
		{"go no comment", "x := 5", LangGo, "x := 5"},
		{"go full line comment", "// full comment", LangGo, ""},
		{"python inline comment", "x = 5 # assign", LangPython, "x = 5 "},
		{"python no comment", "x = 5", LangPython, "x = 5"},
		{"java inline comment", "int x = 5; // assign", LangJava, "int x = 5; "},
		{"unknown slashes", "code // comment", LangUnknown, "code "},
		{"unknown hash", "code # comment", LangUnknown, "code "},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := StripComment(tc.line, tc.lang)
			if got != tc.want {
				t.Errorf("StripComment(%q, %v) = %q, want %q", tc.line, tc.lang, got, tc.want)
			}
		})
	}
}

func TestIsInStringLiteral(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		line string
		pos  int
		want bool
	}{
		{"before first quote", `x := "hello"`, 0, false},
		{"inside string", `x := "hello"`, 7, true},
		{"after closing quote", `x := "hello" + y`, 13, false},
		{"at open quote", `x := "hello"`, 5, false},
		{"empty string position 0", "", 0, false},
		{"escaped quote inside string", `x := "hel\"lo"`, 9, true},
		{"two strings second inside", `"a" + "b"`, 7, true},
		{"two strings between", `"a" + "b"`, 4, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsInStringLiteral(tc.line, tc.pos)
			if got != tc.want {
				t.Errorf("IsInStringLiteral(%q, %d) = %v, want %v", tc.line, tc.pos, got, tc.want)
			}
		})
	}
}
