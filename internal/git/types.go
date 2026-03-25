package git

type FileDiff struct {
	Path    string
	OldPath string
	Hunks   []Hunk
	IsNew   bool
	Deleted bool
}

type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []DiffLine
}

type DiffLine struct {
	Type    LineType
	Content string
	OldNum  int
	NewNum  int
}

type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
)

func (h *Hunk) NewContent() string {
	var s string
	for _, l := range h.Lines {
		if l.Type != LineRemoved {
			s += l.Content + "\n"
		}
	}
	return s
}
