package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zesanrahim/git-aps/internal/analyzer"
	"github.com/zesanrahim/git-aps/internal/fixer"
)

type view int

const (
	viewList view = iota
	viewDetail
	viewConfirm
	viewBatchConfirm
)

type severityFilter int

const (
	filterAll severityFilter = iota
	filterMed
	filterHigh
)

type Model struct {
	allFindings    []analyzer.Finding
	findings       []analyzer.Finding
	cursor         int
	view           view
	message        string
	width          int
	height         int
	severityFilter severityFilter
	appliedCount   int
	skippedCount   int
}

func New(findings []analyzer.Finding) Model {
	all := make([]analyzer.Finding, len(findings))
	copy(all, findings)
	return Model{
		allFindings: all,
		findings:    findings,
		view:        viewList,
	}
}

func (m Model) Summary() (applied, skipped, remaining int) {
	return m.appliedCount, m.skippedCount, len(m.allFindings)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		m.message = ""
		switch m.view {
		case viewList:
			return m.updateList(msg)
		case viewDetail:
			return m.updateDetail(msg)
		case viewConfirm:
			return m.updateConfirm(msg)
		case viewBatchConfirm:
			return m.updateBatchConfirm(msg)
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.findings)-1 {
			m.cursor++
		}
	case key.Matches(msg, keys.Enter):
		if len(m.findings) > 0 {
			m.view = viewDetail
		}
	case key.Matches(msg, keys.Apply):
		if len(m.findings) > 0 {
			f := m.findings[m.cursor]
			if f.FixCode != "" || f.OriginalCode != "" {
				m.view = viewConfirm
			} else {
				m.view = viewDetail
			}
		}
	case key.Matches(msg, keys.BatchApply):
		if m.fixableCount() > 0 {
			m.view = viewBatchConfirm
		}
	case key.Matches(msg, keys.Skip):
		if len(m.findings) > 0 {
			m.skippedCount++
			cur := m.findings[m.cursor]
			m.findings = append(m.findings[:m.cursor], m.findings[m.cursor+1:]...)
			m.removeFromAll(cur)
			if m.cursor >= len(m.findings) && m.cursor > 0 {
				m.cursor--
			}
		}
	case key.Matches(msg, keys.Filter):
		m.severityFilter = (m.severityFilter + 1) % 3
		m.applyFilter()
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.view = viewList
	case key.Matches(msg, keys.Apply):
		f := m.findings[m.cursor]
		if f.FixCode != "" || f.OriginalCode != "" {
			m.view = viewConfirm
		}
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		err := fixer.Apply(m.findings[m.cursor])
		if err != nil {
			m.message = errorStyle.Render("Fix failed: " + err.Error())
		} else {
			m.appliedCount++
			m.message = successStyle.Render("Fix applied to " + m.findings[m.cursor].File)
			cur := m.findings[m.cursor]
			m.findings = append(m.findings[:m.cursor], m.findings[m.cursor+1:]...)
			m.removeFromAll(cur)
			if m.cursor >= len(m.findings) && m.cursor > 0 {
				m.cursor--
			}
		}
		m.view = viewList
	case key.Matches(msg, keys.Deny), key.Matches(msg, keys.Back):
		m.view = viewDetail
	}
	return m, nil
}

func (m Model) updateBatchConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		applied, failed := m.applyAllFixes()
		if failed > 0 {
			m.message = errorStyle.Render(fmt.Sprintf("Applied %d fixes, %d failed", applied, failed))
		} else {
			m.message = successStyle.Render(fmt.Sprintf("Applied %d fixes", applied))
		}
		m.cursor = 0
		m.view = viewList
	case key.Matches(msg, keys.Deny), key.Matches(msg, keys.Back):
		m.view = viewList
	}
	return m, nil
}

func (m *Model) applyAllFixes() (applied, failed int) {
	type fixTarget struct {
		finding analyzer.Finding
		index   int
	}

	byFile := make(map[string][]fixTarget)
	for i, f := range m.findings {
		if f.FixCode == "" && f.OriginalCode == "" {
			continue
		}
		byFile[f.File] = append(byFile[f.File], fixTarget{finding: f, index: i})
	}

	var toRemove []int
	for _, targets := range byFile {
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].finding.Line > targets[j].finding.Line
		})
		for _, t := range targets {
			if err := fixer.Apply(t.finding); err != nil {
				failed++
			} else {
				applied++
				toRemove = append(toRemove, t.index)
			}
		}
	}

	m.appliedCount += applied

	sort.Sort(sort.Reverse(sort.IntSlice(toRemove)))
	for _, idx := range toRemove {
		if idx < len(m.findings) {
			m.removeFromAll(m.findings[idx])
			m.findings = append(m.findings[:idx], m.findings[idx+1:]...)
		}
	}

	return applied, failed
}

func (m *Model) removeFromAll(finding analyzer.Finding) {
	for i, f := range m.allFindings {
		if f.File == finding.File && f.Line == finding.Line && f.Rule == finding.Rule {
			m.allFindings = append(m.allFindings[:i], m.allFindings[i+1:]...)
			return
		}
	}
}

func (m *Model) applyFilter() {
	switch m.severityFilter {
	case filterMed:
		m.findings = analyzer.FilterBySeverity(m.allFindings, analyzer.SeverityMedium)
	case filterHigh:
		m.findings = analyzer.FilterBySeverity(m.allFindings, analyzer.SeverityHigh)
	default:
		m.findings = make([]analyzer.Finding, len(m.allFindings))
		copy(m.findings, m.allFindings)
	}
	m.cursor = 0
}

func (m Model) fixableCount() int {
	count := 0
	for _, f := range m.findings {
		if f.FixCode != "" || f.OriginalCode != "" {
			count++
		}
	}
	return count
}

func (m Model) filterLabel() string {
	switch m.severityFilter {
	case filterMed:
		return " (MED+)"
	case filterHigh:
		return " (HIGH)"
	default:
		return ""
	}
}

func (m Model) View() string {
	if len(m.findings) == 0 {
		msg := titleStyle.Render("git-aps") + "\n\n  No anti-patterns found.\n"
		if m.appliedCount > 0 || m.skippedCount > 0 {
			msg += fmt.Sprintf("\n  %d applied, %d skipped\n", m.appliedCount, m.skippedCount)
		}
		msg += "\n" + helpStyle.Render("q quit")
		return msg
	}

	switch m.view {
	case viewDetail:
		return m.renderDetail()
	case viewConfirm:
		return m.renderConfirm()
	case viewBatchConfirm:
		return m.renderBatchConfirm()
	default:
		return m.renderList()
	}
}

func (m Model) renderList() string {
	var b strings.Builder

	header := fmt.Sprintf(" git-aps  %d findings%s", len(m.findings), m.filterLabel())
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	for i, f := range m.findings {
		sev := severityTag(f.Severity)
		line := fmt.Sprintf("%s  %s:%d  %s", sev, f.File, f.Line, f.Description)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n" + m.message + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate • enter details • a apply • A apply all • s skip • f filter • q quit"))

	return b.String()
}

func (m Model) renderDetail() string {
	f := m.findings[m.cursor]
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf(" %s:%d ", f.File, f.Line)))
	b.WriteString("\n")
	b.WriteString(detailLabelStyle.Render("Severity: ") + severityTag(f.Severity) + "\n")
	b.WriteString(detailLabelStyle.Render("Rule:     ") + f.Rule + "\n")
	b.WriteString(detailLabelStyle.Render("Issue:    ") + f.Description + "\n")
	b.WriteString(detailLabelStyle.Render("Suggest:  ") + f.Suggestion + "\n")

	if f.OriginalCode != "" {
		b.WriteString("\n" + detailLabelStyle.Render("Before:") + "\n")
		b.WriteString(beforeStyle.Render(f.OriginalCode) + "\n")
	}
	if f.FixCode != "" {
		b.WriteString("\n" + detailLabelStyle.Render("After:") + "\n")
		b.WriteString(afterStyle.Render(f.FixCode) + "\n")
	} else if f.OriginalCode != "" {
		b.WriteString("\n" + detailLabelStyle.Render("After:") + "\n")
		b.WriteString(afterStyle.Render("(remove this code)") + "\n")
	}

	b.WriteString("\n")
	if f.FixCode != "" || f.OriginalCode != "" {
		b.WriteString(helpStyle.Render("a apply fix • esc back • q quit"))
	} else {
		b.WriteString(helpStyle.Render("esc back • q quit"))
	}

	return b.String()
}

func (m Model) renderConfirm() string {
	f := m.findings[m.cursor]
	var b strings.Builder

	b.WriteString(headerStyle.Render(" Confirm fix "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Apply fix to %s:%d?\n\n", f.File, f.Line))

	if f.OriginalCode != "" {
		b.WriteString(detailLabelStyle.Render("Before:") + "\n")
		b.WriteString(beforeStyle.Render(f.OriginalCode) + "\n\n")
	}
	if f.FixCode != "" {
		b.WriteString(detailLabelStyle.Render("After:") + "\n")
		b.WriteString(afterStyle.Render(f.FixCode) + "\n\n")
	} else {
		b.WriteString(detailLabelStyle.Render("After:") + "\n")
		b.WriteString(afterStyle.Render("(delete lines)") + "\n\n")
	}

	b.WriteString(helpStyle.Render("y confirm • n cancel"))

	return b.String()
}

func (m Model) renderBatchConfirm() string {
	var b strings.Builder

	fixable := m.fixableCount()
	b.WriteString(headerStyle.Render(" Batch apply "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Apply %d fixable findings?\n\n", fixable))

	files := make(map[string]int)
	for _, f := range m.findings {
		if f.FixCode != "" || f.OriginalCode != "" {
			files[f.File]++
		}
	}
	for file, count := range files {
		b.WriteString(fmt.Sprintf("  %s (%d fixes)\n", file, count))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("y confirm • n cancel"))

	return b.String()
}

func severityTag(s analyzer.Severity) string {
	tag := fmt.Sprintf("[%s]", s)
	switch s {
	case analyzer.SeverityHigh:
		return highStyle.Render(tag)
	case analyzer.SeverityMedium:
		return medStyle.Render(tag)
	default:
		return lowStyle.Render(tag)
	}
}
