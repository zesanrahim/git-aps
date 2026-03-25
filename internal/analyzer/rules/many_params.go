package rules

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type ManyParamsRule struct {
	MaxParams int
}

func (r *ManyParamsRule) Name() string { return "many_params" }

func (r *ManyParamsRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			content := strings.TrimSpace(line.Content)
			if !strings.HasPrefix(content, "func ") {
				continue
			}
			openParen := strings.Index(content, "(")
			closeParen := strings.Index(content, ")")
			if openParen < 0 || closeParen < 0 || closeParen <= openParen+1 {
				continue
			}
			params := strings.Split(content[openParen+1:closeParen], ",")
			if len(params) > r.MaxParams {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityMedium,
					Rule:        r.Name(),
					Description: "Too many function parameters — consider using a struct",
					Suggestion:  "Group related parameters into an options struct",
				})
			}
		}
	}
	return findings
}
