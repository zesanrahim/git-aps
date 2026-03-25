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
			if !strings.HasPrefix(content, "func ") && !strings.HasPrefix(content, "func(") {
				continue
			}
			params := extractParamList(content)
			if params == "" {
				continue
			}
			count := countTopLevelParams(params)
			if count > r.MaxParams {
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

func extractParamList(line string) string {
	idx := strings.Index(line, "func")
	if idx < 0 {
		return ""
	}
	rest := line[idx+4:]

	if len(rest) > 0 && rest[0] == ' ' {
		nameStart := strings.IndexByte(rest, '(')
		if nameStart < 0 {
			return ""
		}
		rest = rest[nameStart:]
	}

	if len(rest) == 0 || rest[0] != '(' {
		return ""
	}

	closeParen := findMatchingParen(rest, 0)
	if closeParen < 0 {
		return ""
	}
	firstGroup := rest[1:closeParen]

	after := rest[closeParen+1:]
	after = strings.TrimSpace(after)
	if len(after) > 0 && after[0] == '(' {
		secondClose := findMatchingParen(after, 0)
		if secondClose > 0 {
			return after[1:secondClose]
		}
	}

	return firstGroup
}

func findMatchingParen(s string, start int) int {
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func countTopLevelParams(params string) int {
	params = strings.TrimSpace(params)
	if params == "" {
		return 0
	}
	count := 1
	depth := 0
	for _, ch := range params {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}
