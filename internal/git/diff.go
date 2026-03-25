package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetDiff(mode string) ([]FileDiff, error) {
	args := diffArgs(mode)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	return ParseDiff(string(out))
}

func diffArgs(mode string) []string {
	switch mode {
	case "unstaged":
		return []string{"diff"}
	case "head":
		return []string{"diff", "HEAD~1"}
	case "branch":
		return []string{"diff", "main...HEAD"}
	default:
		return []string{"diff", "--staged"}
	}
}

func ParseDiff(raw string) ([]FileDiff, error) {
	var files []FileDiff
	chunks := splitFileDiffs(raw)

	for _, chunk := range chunks {
		fd, err := parseFileDiff(chunk)
		if err != nil {
			return nil, err
		}
		files = append(files, fd)
	}
	return files, nil
}

func splitFileDiffs(raw string) []string {
	var chunks []string
	lines := strings.Split(raw, "\n")
	start := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			if start >= 0 {
				chunks = append(chunks, strings.Join(lines[start:i], "\n"))
			}
			start = i
		}
	}
	if start >= 0 {
		chunks = append(chunks, strings.Join(lines[start:], "\n"))
	}
	return chunks
}

func parseFileDiff(chunk string) (FileDiff, error) {
	lines := strings.Split(chunk, "\n")
	fd := FileDiff{}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				fd.Path = strings.TrimPrefix(parts[3], "b/")
			}
		}
		if strings.HasPrefix(line, "new file") {
			fd.IsNew = true
		}
		if strings.HasPrefix(line, "deleted file") {
			fd.Deleted = true
		}
		if strings.HasPrefix(line, "rename from ") {
			fd.OldPath = strings.TrimPrefix(line, "rename from ")
		}
	}

	hunkChunks := splitHunks(lines)
	for _, hunkLines := range hunkChunks {
		hunk, err := parseHunk(hunkLines)
		if err != nil {
			continue
		}
		fd.Hunks = append(fd.Hunks, hunk)
	}

	return fd, nil
}

var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

func splitHunks(lines []string) [][]string {
	var hunks [][]string
	start := -1

	for i, line := range lines {
		if hunkHeaderRe.MatchString(line) {
			if start >= 0 {
				hunks = append(hunks, lines[start:i])
			}
			start = i
		}
	}
	if start >= 0 {
		hunks = append(hunks, lines[start:])
	}
	return hunks
}

func parseHunk(lines []string) (Hunk, error) {
	if len(lines) == 0 {
		return Hunk{}, fmt.Errorf("empty hunk")
	}

	h := Hunk{}
	matches := hunkHeaderRe.FindStringSubmatch(lines[0])
	if matches == nil {
		return h, fmt.Errorf("invalid hunk header: %s", lines[0])
	}

	h.OldStart, _ = strconv.Atoi(matches[1])
	if matches[2] != "" {
		h.OldCount, _ = strconv.Atoi(matches[2])
	}
	h.NewStart, _ = strconv.Atoi(matches[3])
	if matches[4] != "" {
		h.NewCount, _ = strconv.Atoi(matches[4])
	}

	oldNum := h.OldStart
	newNum := h.NewStart
	for _, line := range lines[1:] {
		if len(line) == 0 {
			h.Lines = append(h.Lines, DiffLine{
				Type:    LineContext,
				Content: "",
				OldNum:  oldNum,
				NewNum:  newNum,
			})
			oldNum++
			newNum++
			continue
		}
		dl := DiffLine{Content: line[1:]}
		switch line[0] {
		case '+':
			dl.Type = LineAdded
			dl.NewNum = newNum
			newNum++
		case '-':
			dl.Type = LineRemoved
			dl.OldNum = oldNum
			oldNum++
		default:
			dl.Type = LineContext
			dl.OldNum = oldNum
			dl.NewNum = newNum
			oldNum++
			newNum++
		}
		h.Lines = append(h.Lines, dl)
	}

	return h, nil
}
