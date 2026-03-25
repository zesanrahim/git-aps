package rules

import (
	"regexp"
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
	lang := DetectLanguage(file.Path)
	var findings []analyzer.Finding
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			if line.Type != git.LineAdded {
				continue
			}
			if containsMagicNumber(line.Content, r.Threshold, lang) {
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

var (
	ipPattern       = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	versionPattern  = regexp.MustCompile(`v?\d+\.\d+(\.\d+)+`)
	hexPattern      = regexp.MustCompile(`0[xX][0-9a-fA-F]+`)
	octalPattern    = regexp.MustCompile(`0[oO][0-7]+`)
	indexPattern    = regexp.MustCompile(`\[\d+\]`)
	portPattern     = regexp.MustCompile(`":\d+"`)
	bitShiftPattern = regexp.MustCompile(`[<>]{2}\s*\d+`)
	timeDurPattern  = regexp.MustCompile(`\d+\s*\*\s*time\.`)
)

func containsMagicNumber(line string, threshold int, lang Language) bool {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "const") || strings.HasPrefix(trimmed, "var") {
		return false
	}
	if IsComment(trimmed, lang) {
		return false
	}
	if strings.Contains(trimmed, "iota") {
		return false
	}

	cleaned := StripComment(trimmed, lang)

	cleaned = ipPattern.ReplaceAllString(cleaned, "")
	cleaned = versionPattern.ReplaceAllString(cleaned, "")
	cleaned = hexPattern.ReplaceAllString(cleaned, "")
	cleaned = octalPattern.ReplaceAllString(cleaned, "")
	cleaned = indexPattern.ReplaceAllString(cleaned, "")
	cleaned = portPattern.ReplaceAllString(cleaned, "")
	cleaned = bitShiftPattern.ReplaceAllString(cleaned, "")

	if timeDurPattern.MatchString(trimmed) {
		return false
	}

	words := strings.FieldsFunc(cleaned, func(r rune) bool {
		return !unicode.IsDigit(r) && r != '.' && r != '-'
	})

	for _, w := range words {
		if w == "" || w == "." || w == "-" {
			continue
		}

		start := strings.Index(cleaned, w)
		if start >= 0 && IsInStringLiteral(cleaned, start) {
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
