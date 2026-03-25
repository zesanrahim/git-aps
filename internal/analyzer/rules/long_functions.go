package rules

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type LongFunctionRule struct {
	MaxLines int
}

func (r *LongFunctionRule) Name() string { return "long_functions" }

func (r *LongFunctionRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		funcStart := -1
		funcLine := 0
		lineCount := 0

		for _, line := range hunk.Lines {
			if line.Type == git.LineRemoved {
				continue
			}
			content := strings.TrimSpace(line.Content)
			if strings.HasPrefix(content, "func ") {
				if funcStart > 0 && lineCount > r.MaxLines {
					findings = append(findings, makeLongFuncFinding(file.Path, funcStart, r))
				}
				funcStart = line.NewNum
				funcLine = line.NewNum
				lineCount = 0
			}
			if funcStart > 0 {
				lineCount++
			}
			_ = funcLine
		}
		if funcStart > 0 && lineCount > r.MaxLines {
			findings = append(findings, makeLongFuncFinding(file.Path, funcStart, r))
		}
	}
	return findings
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
