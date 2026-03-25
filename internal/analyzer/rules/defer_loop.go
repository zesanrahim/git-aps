package rules

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type DeferLoopRule struct{}

func (r *DeferLoopRule) Name() string { return "defer_loop" }

func (r *DeferLoopRule) Check(file git.FileDiff) []analyzer.Finding {
	if DetectLanguage(file.Path) != LangGo {
		return nil
	}
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		var loopIndents []int
		for _, line := range hunk.Lines {
			if line.Type == git.LineRemoved {
				continue
			}
			content := line.Content
			trimmed := strings.TrimSpace(content)
			indent := nestingDepth(content)

			for len(loopIndents) > 0 && indent <= loopIndents[len(loopIndents)-1] && trimmed != "" {
				loopIndents = loopIndents[:len(loopIndents)-1]
			}

			if line.Type == git.LineAdded && strings.HasPrefix(trimmed, "for ") || strings.HasPrefix(trimmed, "for{") {
				loopIndents = append(loopIndents, indent)
			}

			if line.Type == git.LineAdded && len(loopIndents) > 0 && strings.HasPrefix(trimmed, "defer ") {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityHigh,
					Rule:        r.Name(),
					Description: "defer inside a loop — deferred calls run at function exit, not loop iteration end",
					Suggestion:  "Move the deferred operation into a separate function or call it directly",
				})
			}
		}
	}
	return findings
}
