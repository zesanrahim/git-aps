package rules

import (
	"regexp"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type CustomRule struct {
	RuleName    string
	Pattern     *regexp.Regexp
	Sev         analyzer.Severity
	Desc        string
	Suggestion_ string
}

func (r *CustomRule) Name() string { return r.RuleName }

func (r *CustomRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding

	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if r.Pattern.MatchString(line.Content) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					EndLine:     line.NewNum,
					Severity:    r.Sev,
					Rule:        r.RuleName,
					Description: r.Desc,
					Suggestion:  r.Suggestion_,
				})
			}
		}
	}

	return findings
}
