package rules

import (
	"regexp"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type SQLInjectionRule struct{}

func (r *SQLInjectionRule) Name() string { return "sql_injection" }

var sqlConcatPattern = regexp.MustCompile(`(?i)["'](?:SELECT|INSERT|UPDATE|DELETE|DROP|ALTER)\s.*["']\s*\+`)
var sqlSprintfPattern = regexp.MustCompile(`(?i)fmt\.Sprintf\(\s*["'].*(?:SELECT|INSERT|UPDATE|DELETE|DROP|ALTER)\s.*%[sv]`)
var parameterizedPattern = regexp.MustCompile(`\?\s*[,)]|\$\d`)

func (r *SQLInjectionRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	lang := DetectLanguage(file.Path)
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if IsComment(line.Content, lang) {
				continue
			}
			if parameterizedPattern.MatchString(line.Content) {
				continue
			}
			if sqlConcatPattern.MatchString(line.Content) || sqlSprintfPattern.MatchString(line.Content) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityHigh,
					Rule:        r.Name(),
					Description: "Possible SQL injection — query built with string concatenation or formatting",
					Suggestion:  "Use parameterized queries with ? or $1 placeholders",
				})
			}
		}
	}
	return findings
}
