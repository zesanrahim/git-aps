package rules

import "path/filepath"

type Language int

const (
	LangUnknown Language = iota
	LangGo
	LangPython
	LangJavaScript
	LangTypeScript
	LangJava
	LangRust
)

func DetectLanguage(path string) Language {
	switch filepath.Ext(path) {
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".js", ".jsx", ".mjs":
		return LangJavaScript
	case ".ts", ".tsx":
		return LangTypeScript
	case ".java":
		return LangJava
	case ".rs":
		return LangRust
	default:
		return LangUnknown
	}
}
