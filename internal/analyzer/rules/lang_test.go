package rules

import "testing"

func TestDetectLanguage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		want Language
	}{
		{"go file", "main.go", LangGo},
		{"go nested path", "internal/foo/bar.go", LangGo},
		{"python file", "script.py", LangPython},
		{"js file", "app.js", LangJavaScript},
		{"jsx file", "component.jsx", LangJavaScript},
		{"mjs file", "module.mjs", LangJavaScript},
		{"ts file", "app.ts", LangTypeScript},
		{"tsx file", "component.tsx", LangTypeScript},
		{"java file", "Main.java", LangJava},
		{"rust file", "main.rs", LangRust},
		{"unknown extension", "README.md", LangUnknown},
		{"no extension", "Makefile", LangUnknown},
		{"c file", "main.c", LangUnknown},
		{"yaml file", "config.yaml", LangUnknown},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DetectLanguage(tc.path)
			if got != tc.want {
				t.Errorf("DetectLanguage(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
