package rules

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type ErrorIgnoredRule struct{}

func (r *ErrorIgnoredRule) Name() string { return "error_ignored" }

func (r *ErrorIgnoredRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			trimmed := strings.TrimSpace(line.Content)
			if isIgnoredError(trimmed) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityHigh,
					Rule:        r.Name(),
					Description: "Error return value is being discarded",
					Suggestion:  "Handle the error or explicitly assign to _ with a comment explaining why",
				})
			}
		}
	}
	return findings
}

func isIgnoredError(line string) bool {
	patterns := []string{
		"_ = ",
		"_ =\t",
	}
	for _, p := range patterns {
		if strings.Contains(line, p) && (strings.Contains(line, "err") || strings.Contains(line, "Err")) {
			return true
		}
	}
	return false
}
