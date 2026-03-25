package rules

import (
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type StringConcatRule struct{}

func (r *StringConcatRule) Name() string { return "perf_string_concat" }

var stringConcatPattern = regexp.MustCompile(`\w+\s*\+=\s*`)

func (r *StringConcatRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	lang := DetectLanguage(file.Path)
	for _, hunk := range file.Hunks {
		var loopIndents []int
		for _, line := range hunk.Lines {
			if line.Type == git.LineRemoved {
				continue
			}
			trimmed := strings.TrimSpace(line.Content)
			indent := nestingDepth(line.Content)

			for len(loopIndents) > 0 && indent <= loopIndents[len(loopIndents)-1] && trimmed != "" {
				loopIndents = loopIndents[:len(loopIndents)-1]
			}

			if strings.HasPrefix(trimmed, "for ") || strings.HasPrefix(trimmed, "for{") {
				loopIndents = append(loopIndents, indent)
			}

			if line.Type == git.LineAdded && len(loopIndents) > 0 {
				if IsComment(line.Content, lang) {
					continue
				}
				if stringConcatPattern.MatchString(trimmed) {
					findings = append(findings, analyzer.Finding{
						File:        file.Path,
						Line:        line.NewNum,
						Severity:    analyzer.SeverityLow,
						Rule:        r.Name(),
						Description: "String concatenation inside a loop — consider using strings.Builder",
						Suggestion:  "Use strings.Builder for efficient string building in loops",
					})
				}
			}
		}
	}
	return findings
}

type RegexLoopRule struct{}

func (r *RegexLoopRule) Name() string { return "perf_regex_loop" }

var regexCompilePattern = regexp.MustCompile(`regexp\.(MustCompile|Compile)\(`)

func (r *RegexLoopRule) Check(file git.FileDiff) []analyzer.Finding {
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
			trimmed := strings.TrimSpace(line.Content)
			indent := nestingDepth(line.Content)

			for len(loopIndents) > 0 && indent <= loopIndents[len(loopIndents)-1] && trimmed != "" {
				loopIndents = loopIndents[:len(loopIndents)-1]
			}

			if strings.HasPrefix(trimmed, "for ") || strings.HasPrefix(trimmed, "for{") {
				loopIndents = append(loopIndents, indent)
			}

			if line.Type == git.LineAdded && len(loopIndents) > 0 {
				if regexCompilePattern.MatchString(trimmed) {
					findings = append(findings, analyzer.Finding{
						File:        file.Path,
						Line:        line.NewNum,
						Severity:    analyzer.SeverityMedium,
						Rule:        r.Name(),
						Description: "Regex compiled inside a loop — compile once outside the loop",
						Suggestion:  "Move regexp.Compile/MustCompile to package level or before the loop",
					})
				}
			}
		}
	}
	return findings
}
