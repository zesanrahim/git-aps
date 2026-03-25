package rules

import (
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type GoroutineCtxRule struct{}

func (r *GoroutineCtxRule) Name() string { return "goroutine_ctx" }

var goAnonPattern = regexp.MustCompile(`go\s+func\s*\(`)
var goNamedPattern = regexp.MustCompile(`go\s+\w+\(`)

func (r *GoroutineCtxRule) Check(file git.FileDiff) []analyzer.Finding {
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
			trimmed := strings.TrimSpace(line.Content)

			if goAnonPattern.MatchString(trimmed) {
				if !strings.Contains(trimmed, "ctx") && !strings.Contains(trimmed, "context.Context") {
					findings = append(findings, makeGoroutineCtxFinding(file.Path, line.NewNum))
				}
				continue
			}

			if goNamedPattern.MatchString(trimmed) && !strings.HasPrefix(trimmed, "go func") {
				parenIdx := strings.Index(trimmed, "(")
				if parenIdx >= 0 {
					args := trimmed[parenIdx:]
					if !strings.Contains(args, "ctx") {
						findings = append(findings, makeGoroutineCtxFinding(file.Path, line.NewNum))
					}
				}
			}
		}
	}
	return findings
}

func makeGoroutineCtxFinding(path string, lineNum int) analyzer.Finding {
	return analyzer.Finding{
		File:        path,
		Line:        lineNum,
		Severity:    analyzer.SeverityLow,
		Rule:        "goroutine_ctx",
		Description: "Goroutine launched without context — cancellation and timeout won't propagate",
		Suggestion:  "Pass a context.Context to enable cancellation and timeout propagation",
	}
}
