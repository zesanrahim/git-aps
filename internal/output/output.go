package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

type jsonFinding struct {
	File         string `json:"file"`
	Line         int    `json:"line"`
	EndLine      int    `json:"end_line,omitempty"`
	Severity     string `json:"severity"`
	Rule         string `json:"rule"`
	Description  string `json:"description"`
	Suggestion   string `json:"suggestion,omitempty"`
	OriginalCode string `json:"original_code,omitempty"`
	FixCode      string `json:"fix_code,omitempty"`
}

func PrintJSON(findings []analyzer.Finding, w io.Writer) error {
	out := make([]jsonFinding, len(findings))
	for i, f := range findings {
		out[i] = jsonFinding{
			File:         f.File,
			Line:         f.Line,
			EndLine:      f.EndLine,
			Severity:     f.Severity.String(),
			Rule:         f.Rule,
			Description:  f.Description,
			Suggestion:   f.Suggestion,
			OriginalCode: f.OriginalCode,
			FixCode:      f.FixCode,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func PrintText(findings []analyzer.Finding, w io.Writer) {
	var high, med, low int
	for _, f := range findings {
		switch f.Severity {
		case analyzer.SeverityHigh:
			high++
		case analyzer.SeverityMedium:
			med++
		default:
			low++
		}
		fmt.Fprintf(w, "[%s] %s:%d — %s (%s)\n", f.Severity, f.File, f.Line, f.Description, f.Rule)
	}
	fmt.Fprintf(w, "\n%d findings (%d high, %d med, %d low)\n", len(findings), high, med, low)
}
