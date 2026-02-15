package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the Bubble Tea model for the bltctl TUI.
type Model struct {
	version string
}

// New creates a new TUI model.
func New(version string) Model {
	return Model{version: version}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	return "bltctl " + m.version + " — loading..."
}
