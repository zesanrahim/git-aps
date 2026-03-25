package analyzer

import (
	"strings"

	"github.com/zesanrahim/git-aps/internal/git"
)

type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
)

func (s Severity) String() string {
	switch s {
	case SeverityHigh:
		return "HIGH"
	case SeverityMedium:
		return "MED"
	default:
		return "LOW"
	}
}

func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "high":
		return SeverityHigh
	case "med", "medium":
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func FilterBySeverity(findings []Finding, minSeverity Severity) []Finding {
	if minSeverity <= SeverityLow {
		return findings
	}
	var filtered []Finding
	for _, f := range findings {
		if f.Severity >= minSeverity {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

type Finding struct {
	File         string
	Line         int
	EndLine      int
	Severity     Severity
	Rule         string
	Description  string
	Suggestion   string
	OriginalCode string
	FixCode      string
}

type Analyzer interface {
	Analyze(diffs []git.FileDiff) ([]Finding, error)
}
