package fixer

import (
	"fmt"
	"os"
	"strings"

	"github.com/zesanrahim/git-aps/internal/analyzer"
)

func Apply(finding analyzer.Finding) error {
	if finding.FixCode == "" && finding.OriginalCode == "" {
		return fmt.Errorf("no fix available for this finding")
	}

	data, err := os.ReadFile(finding.File)
	if err != nil {
		return fmt.Errorf("reading %s: %w", finding.File, err)
	}

	lines := strings.Split(string(data), "\n")
	if finding.Line < 1 || finding.Line > len(lines) {
		return fmt.Errorf("line %d out of range (file has %d lines)", finding.Line, len(lines))
	}

	end := finding.Line
	if finding.EndLine > finding.Line {
		end = finding.EndLine
	}
	if end > len(lines) {
		end = len(lines)
	}

	newLines := make([]string, 0, len(lines))
	newLines = append(newLines, lines[:finding.Line-1]...)
	if finding.FixCode != "" {
		fixLines := strings.Split(finding.FixCode, "\n")
		newLines = append(newLines, fixLines...)
	}
	newLines = append(newLines, lines[end:]...)

	return os.WriteFile(finding.File, []byte(strings.Join(newLines, "\n")), 0644)
}
