package rules

import (
	"regexp"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type ErrorWrapRule struct{}

func (r *ErrorWrapRule) Name() string { return "error_wrap" }

var errfVerbPattern = regexp.MustCompile(`fmt\.Errorf\([^)]*%[vs][^)]*,\s*\w*[eE]rr`)
var errfHasW = regexp.MustCompile(`%w`)

func (r *ErrorWrapRule) Check(file git.FileDiff) []analyzer.Finding {
	if DetectLanguage(file.Path) != LangGo {
		return nil
	}
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if errfVerbPattern.MatchString(line.Content) && !errfHasW.MatchString(line.Content) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityMedium,
					Rule:        r.Name(),
					Description: "Error wrapped with %v/%s instead of %w — errors.Is/As won't work up the chain",
					Suggestion:  "Use %w instead of %v or %s to preserve the error chain",
				})
			}
		}
	}
	return findings
}
