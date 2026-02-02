package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.layout == LayoutSingle && m.activeTab == TabConfig {
				m.showPreview = !m.showPreview
			}
		case "right", "l":
			if m.activeTab < TabLogs {
				m.activeTab++
			}
		case "left", "h":
			if m.activeTab > TabConfig {
				m.activeTab--
			}
		case "up", "k":
			if m.activeTab == TabConfig && m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.activeTab == TabConfig && m.selectedIdx < len(m.accounts)-1 {
				m.selectedIdx++
			}
		}
	}
	return m, nil
}

func (m *Model) updateLayout() {
	switch {
	case m.width < 60 || m.height < 15:
		m.layout = LayoutUnsupported
	case m.width < 100:
		m.layout = LayoutSingle
	default:
		m.layout = LayoutSplit
	}
}
