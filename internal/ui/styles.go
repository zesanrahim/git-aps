package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			PaddingLeft(1)

	highStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	medStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	lowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			PaddingLeft(1).
			PaddingRight(1)

	normalStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("63"))

	fixCodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Background(lipgloss.Color("234")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	beforeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Padding(0, 1)

	afterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40")).
			Background(lipgloss.Color("22")).
			Padding(0, 1)
)
