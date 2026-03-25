package analyzer

import (
	"sort"
	"strconv"
)

func MergeFindings(sources ...[]Finding) []Finding {
	seen := map[string]bool{}
	var merged []Finding

	for _, findings := range sources {
		for _, f := range findings {
			key := f.File + ":" + f.Rule + ":" + strconv.Itoa(f.Line)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, f)
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Severity != merged[j].Severity {
			return merged[i].Severity > merged[j].Severity
		}
		if merged[i].File != merged[j].File {
			return merged[i].File < merged[j].File
		}
		return merged[i].Line < merged[j].Line
	})

	return merged
}
