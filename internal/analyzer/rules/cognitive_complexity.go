package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type CognitiveComplexityRule struct {
	Threshold int
}

func (r *CognitiveComplexityRule) Name() string { return "cognitive_complexity" }

var controlFlowPattern = regexp.MustCompile(`\b(if|else\s+if|else|for|while|switch|select)\b`)
var boolOpPattern = regexp.MustCompile(`&&|\|\|`)
var gotoPattern = regexp.MustCompile(`\bgoto\b`)

func (r *CognitiveComplexityRule) Check(file git.FileDiff) []analyzer.Finding {
	lang := DetectLanguage(file.Path)
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		funcStart := -1
		funcIndent := 0
		complexity := 0

		for _, line := range hunk.Lines {
			if line.Type == git.LineRemoved {
				continue
			}
			content := line.Content
			trimmed := strings.TrimSpace(content)

			if isFuncStart(trimmed) {
				if funcStart > 0 && complexity > r.Threshold {
					findings = append(findings, makeCognitiveComplexityFinding(file.Path, funcStart, complexity, r.Threshold))
				}
				funcStart = line.NewNum
				funcIndent = nestingDepth(content)
				complexity = 0
				continue
			}

			if funcStart < 0 {
				continue
			}
			if IsBlankOrComment(content, lang) {
				continue
			}

			relativeDepth := nestingDepth(content) - funcIndent - 1
			if relativeDepth < 0 {
				relativeDepth = 0
			}

			matches := controlFlowPattern.FindAllString(trimmed, -1)
			for _, m := range matches {
				increment := 1
				if strings.HasPrefix(m, "else if") {
					increment = 1
				} else if m == "else" {
					increment = 1
				}
				complexity += increment + relativeDepth
			}

			boolOps := boolOpPattern.FindAllString(trimmed, -1)
			complexity += len(boolOps)

			if gotoPattern.MatchString(trimmed) {
				complexity++
			}
		}

		if funcStart > 0 && complexity > r.Threshold {
			findings = append(findings, makeCognitiveComplexityFinding(file.Path, funcStart, complexity, r.Threshold))
		}
	}
	return findings
}

func makeCognitiveComplexityFinding(path string, line, complexity, threshold int) analyzer.Finding {
	sev := analyzer.SeverityMedium
	if complexity > threshold*2 {
		sev = analyzer.SeverityHigh
	}
	return analyzer.Finding{
		File:        path,
		Line:        line,
		Severity:    sev,
		Rule:        "cognitive_complexity",
		Description: fmt.Sprintf("Cognitive complexity of %d exceeds threshold of %d", complexity, threshold),
		Suggestion:  "Break down complex logic into smaller functions or simplify conditionals",
	}
}
