package rules

import (
	"regexp"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type TodoCommentRule struct{}

func (r *TodoCommentRule) Name() string { return "todo_comments" }

var todoInCommentPattern = regexp.MustCompile(`(?i)(?://|/\*|#)\s*.*\b(TODO|FIXME|HACK|XXX)\b`)

func (r *TodoCommentRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			matches := todoInCommentPattern.FindStringSubmatch(line.Content)
			if matches != nil {
				marker := matches[1]
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityLow,
					Rule:        r.Name(),
					Description: marker + " comment found — track this as a ticket instead",
					Suggestion:  "Create an issue and remove the comment",
				})
			}
		}
	}
	return findings
}
