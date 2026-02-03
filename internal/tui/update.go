package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		// Handle edit mode first
		if m.editMode != EditNone {
			switch msg.String() {
			case "esc":
				m.editMode = EditNone
				m.editValue = ""
			case "enter":
				// Save edit
				if m.editMode == EditName && m.editValue != "" {
					if len(m.accounts) == 0 || m.selectedIdx >= len(m.accounts) {
						// New account
						m.config.Accounts[m.editValue] = config.Account{}
						m.accounts = append(m.accounts, m.editValue)
						m.selectedIdx = len(m.accounts) - 1
					}
				}
				m.editMode = EditNone
				m.editValue = ""
			case "backspace":
				if len(m.editValue) > 0 {
					m.editValue = m.editValue[:len(m.editValue)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.editValue += msg.String()
				}
			}
			return m, nil
		}

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
		case "e":
			if m.activeTab == TabConfig && m.editMode == EditNone && len(m.accounts) > 0 {
				m.editMode = EditName
				m.editValue = m.accounts[m.selectedIdx]
			}
		case "n":
			if m.activeTab == TabConfig && m.editMode == EditNone {
				m.editMode = EditName
				m.editValue = ""
			}
		case "d":
			if m.activeTab == TabConfig && m.editMode == EditNone && len(m.accounts) > 0 {
				// Delete selected account
				name := m.accounts[m.selectedIdx]
				delete(m.config.Accounts, name)
				m.accounts = append(m.accounts[:m.selectedIdx], m.accounts[m.selectedIdx+1:]...)
				if m.selectedIdx >= len(m.accounts) {
					m.selectedIdx = len(m.accounts) - 1
				}
				if m.selectedIdx < 0 {
					m.selectedIdx = 0
				}
			}
		}
	}
	return m, nil
}

func (m *Model) updateLayout() {
	// Layout breakpoints:
	// - Unsupported: < 40 width or < 10 height (minimum usable)
	// - Single: 40-79 width (enough for one panel)
	//   - Account list needs ~25 chars (name + status)
	//   - Preview needs ~35 chars (command examples)
	// - Split: 80+ width (left 30 + right 40 + gap 2 = 72 minimum)
	switch {
	case m.width < 40 || m.height < 10:
		m.layout = LayoutUnsupported
	case m.width < 80:
		m.layout = LayoutSingle
	default:
		m.layout = LayoutSplit
	}
}
