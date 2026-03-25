package ai

import (
	"strconv"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func parseResponse(filePath string, response string) ([]analyzer.Finding, error) {
	if strings.TrimSpace(response) == "NO_ISSUES" {
		return nil, nil
	}

	var findings []analyzer.Finding
	blocks := strings.Split(response, "FINDING")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		f := analyzer.Finding{File: filePath}
		lines := strings.Split(block, "\n")

		inFix := false
		var fixLines []string

		for _, line := range lines {
			if inFix {
				if strings.TrimSpace(line) == "END_FIX" {
					inFix = false
					f.FixCode = strings.Join(fixLines, "\n")
					continue
				}
				fixLines = append(fixLines, line)
				continue
			}

			if strings.HasPrefix(line, "LINE:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "LINE:"))
				f.Line, _ = strconv.Atoi(val)
			} else if strings.HasPrefix(line, "END_LINE:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "END_LINE:"))
				f.EndLine, _ = strconv.Atoi(val)
			} else if strings.HasPrefix(line, "SEVERITY:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "SEVERITY:"))
				f.Severity = parseSeverity(val)
			} else if strings.HasPrefix(line, "RULE:") {
				f.Rule = strings.TrimSpace(strings.TrimPrefix(line, "RULE:"))
			} else if strings.HasPrefix(line, "DESCRIPTION:") {
				f.Description = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
			} else if strings.HasPrefix(line, "SUGGESTION:") {
				f.Suggestion = strings.TrimSpace(strings.TrimPrefix(line, "SUGGESTION:"))
			} else if strings.HasPrefix(line, "FIX:") {
				inFix = true
				fixLines = nil
			}
		}

		if f.Description != "" {
			f.FixCode = stripCodeFences(f.FixCode)
			findings = append(findings, f)
		}
	}

	return findings, nil
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		first := strings.Index(s, "\n")
		if first >= 0 {
			s = s[first+1:]
		}
	}
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	return strings.TrimSpace(s)
}

func parseSeverity(s string) analyzer.Severity {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "HIGH":
		return analyzer.SeverityHigh
	case "MED", "MEDIUM":
		return analyzer.SeverityMedium
	default:
		return analyzer.SeverityLow
	}
}
