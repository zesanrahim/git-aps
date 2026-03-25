package rules

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/git"
)

type MagicNumberRule struct {
	Threshold int
}

func (r *MagicNumberRule) Name() string { return "magic_numbers" }

func (r *MagicNumberRule) Check(file git.FileDiff) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if containsMagicNumber(line.Content, r.Threshold) {
				findings = append(findings, analyzer.Finding{
					File:        file.Path,
					Line:        line.NewNum,
					Severity:    analyzer.SeverityLow,
					Rule:        r.Name(),
					Description: "Magic number detected — consider extracting to a named constant",
					Suggestion:  "Replace with a descriptive constant",
				})
			}
		}
	}
	return findings
}

func containsMagicNumber(line string, threshold int) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "const") || strings.HasPrefix(trimmed, "var") {
		return false
	}

	words := strings.FieldsFunc(trimmed, func(r rune) bool {
		return !unicode.IsDigit(r) && r != '.' && r != '-'
	})

	for _, w := range words {
		if w == "" || w == "." || w == "-" {
			continue
		}
		n, err := strconv.ParseFloat(w, 64)
		if err != nil {
			continue
		}
		if n > float64(threshold) || n < -float64(threshold) {
			return true
		}
	}
	return false
}
