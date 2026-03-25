package rules

import (
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type LongFunctionRule struct {
	MaxLines int
}

func (r *LongFunctionRule) Name() string { return "long_functions" }

var funcStartPattern = regexp.MustCompile(`^func[\s(]`)

func (r *LongFunctionRule) Check(file git.FileDiff) []analyzer.Finding {
	lang := DetectLanguage(file.Path)
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		funcStart := -1
		codeLines := 0

		for _, line := range hunk.Lines {
			if line.Type == git.LineRemoved {
				continue
			}
			content := strings.TrimSpace(line.Content)
			if isFuncStart(content) {
				if funcStart > 0 && codeLines > r.MaxLines {
					findings = append(findings, makeLongFuncFinding(file.Path, funcStart, r))
				}
				funcStart = line.NewNum
				codeLines = 0
			}
			if funcStart > 0 {
				if !IsBlankOrComment(line.Content, lang) {
					codeLines++
				}
			}
		}
		if funcStart > 0 && codeLines > r.MaxLines {
			findings = append(findings, makeLongFuncFinding(file.Path, funcStart, r))
		}
	}
	return findings
}

func isFuncStart(content string) bool {
	return funcStartPattern.MatchString(content)
}

func makeLongFuncFinding(path string, line int, r *LongFunctionRule) analyzer.Finding {
	return analyzer.Finding{
		File:        path,
		Line:        line,
		Severity:    analyzer.SeverityMedium,
		Rule:        r.Name(),
		Description: "Function exceeds maximum line count — consider breaking it up",
		Suggestion:  "Split into smaller, focused functions",
	}
}
