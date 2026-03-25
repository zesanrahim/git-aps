package rules

import "strings"

func IsComment(line string, lang Language) bool {
	trimmed := strings.TrimSpace(line)
	switch lang {
	case LangGo, LangJava, LangJavaScript, LangTypeScript, LangRust:
		return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
	case LangPython:
		return strings.HasPrefix(trimmed, "#")
	default:
		return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
	}
}

func IsBlankOrComment(line string, lang Language) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	return IsComment(line, lang)
}

func HasCommentPrefix(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.Contains(trimmed, "//") || strings.Contains(trimmed, "/*") || strings.Contains(trimmed, "#")
}

func StripComment(line string, lang Language) string {
	switch lang {
	case LangGo, LangJava, LangJavaScript, LangTypeScript, LangRust:
		if idx := strings.Index(line, "//"); idx >= 0 {
			return line[:idx]
		}
	case LangPython:
		if idx := strings.Index(line, "#"); idx >= 0 {
			return line[:idx]
		}
	default:
		if idx := strings.Index(line, "//"); idx >= 0 {
			return line[:idx]
		}
		if idx := strings.Index(line, "#"); idx >= 0 {
			return line[:idx]
		}
	}
	return line
}

func IsInStringLiteral(line string, pos int) bool {
	doubleQuotes := 0
	for i := 0; i < pos && i < len(line); i++ {
		if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
			doubleQuotes++
		}
	}
	return doubleQuotes%2 == 1
}
