package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Back       key.Binding
	Apply      key.Binding
	BatchApply key.Binding
	Skip       key.Binding
	Filter     key.Binding
	Quit       key.Binding
	Confirm    key.Binding
	Deny       key.Binding
}

var keys = keyMap{
	Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
	Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Apply:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "apply fix")),
	BatchApply: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "apply all")),
	Skip:       key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "skip")),
	Filter:     key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Confirm:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
	Deny:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "cancel")),
}
