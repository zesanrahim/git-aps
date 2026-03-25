package rules

import (
	"regexp"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type TypeAssertRule struct{}

func (r *TypeAssertRule) Name() string { return "type_assert" }

var unsafeAssertPattern = regexp.MustCompile(`\w+\.\([^)]+\)`)
var commaOkPattern = regexp.MustCompile(`\w+\s*,\s*\w+\s*:?=\s*\w+\.\(`)
var typeSwitchPattern = regexp.MustCompile(`\.\(type\)`)

func (r *TypeAssertRule) Check(file git.FileDiff) []analyzer.Finding {
	if DetectLanguage(file.Path) != LangGo {
		return nil
	}
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if IsComment(line.Content, LangGo) {
				continue
			}
			if typeSwitchPattern.MatchString(line.Content) {
				continue
			}
			if commaOkPattern.MatchString(line.Content) {
				continue
			}
			if unsafeAssertPattern.MatchString(line.Content) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityMedium,
					Rule:        r.Name(),
					Description: "Type assertion without ok check — will panic if assertion fails",
					Suggestion:  "Use the comma-ok pattern: v, ok := x.(Type)",
				})
			}
		}
	}
	return findings
}
