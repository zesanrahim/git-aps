package rules

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type DeepNestingRule struct {
	MaxDepth int
}

func (r *DeepNestingRule) Name() string { return "deep_nesting" }

func (r *DeepNestingRule) Check(file git.FileDiff) []analyzer.Finding {
	lang := DetectLanguage(file.Path)
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		inDeepBlock := false
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if IsBlankOrComment(line.Content, lang) {
				continue
			}
			trimmed := strings.TrimSpace(line.Content)
			if isClosingBrace(trimmed) {
				continue
			}

			depth := nestingDepth(line.Content)
			if depth > r.MaxDepth {
				if !inDeepBlock {
					findings = append(findings, analyzer.Finding{
						File:        file.Path,
						Line:        line.NewNum,
						Severity:    analyzer.SeverityMedium,
						Rule:        r.Name(),
						Description: "Deeply nested code — consider early returns or extracting functions",
						Suggestion:  "Refactor to reduce nesting depth",
					})
					inDeepBlock = true
				}
			} else {
				inDeepBlock = false
			}
		}
	}
	return findings
}

func isClosingBrace(trimmed string) bool {
	cleaned := strings.TrimRight(trimmed, " \t;,")
	return cleaned == "}" || cleaned == "})" || cleaned == "});"
}

func nestingDepth(line string) int {
	trimmed := strings.TrimRight(line, " \t")
	if len(trimmed) == 0 {
		return 0
	}
	indent := len(line) - len(strings.TrimLeft(line, "\t"))
	if indent == 0 {
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		indent = spaces / 4
	}
	return indent
}
