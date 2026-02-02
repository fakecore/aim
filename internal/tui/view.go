package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.layout == LayoutUnsupported {
		return m.unsupportedView()
	}

	var sections []string
	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderContent())
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) unsupportedView() string {
	return fmt.Sprintf(
		"Terminal too small\n\nCurrent: %d x %d\nMinimum: 60 x 15\n\nPlease resize and retry",
		m.width, m.height,
	)
}

func (m Model) renderHeader() string {
	tabs := []string{"Config", "Status", "Routes", "Usage", "Logs"}
	var rendered []string
	for i, tab := range tabs {
		style := tabStyle
		if Tab(i) == m.activeTab {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(tab))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, rendered...)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case TabConfig:
		return m.renderConfigTab()
	case TabStatus:
		return m.renderStatusTab()
	default:
		return placeholderStyle.Render("Coming soon...")
	}
}

func (m Model) renderConfigTab() string {
	if m.layout == LayoutSplit {
		left := m.renderAccountList()
		right := m.renderPreview()
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanelStyle.Render(left),
			rightPanelStyle.Render(right),
		)
	}
	if m.showPreview {
		return m.renderPreview()
	}
	return m.renderAccountList()
}

func (m Model) renderAccountList() string {
	var lines []string
	lines = append(lines, titleStyle.Render("ACCOUNTS"))
	lines = append(lines, "")
	for i, name := range m.accounts {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "> "
		}
		acc := m.config.Accounts[name]
		status := "✓"
		if acc.Key == "" {
			status = "⚠"
		}
		line := fmt.Sprintf("%s%s %s", prefix, status, name)
		if i == m.selectedIdx {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	if m.layout == LayoutSingle {
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Tab: switch to preview"))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderPreview() string {
	if len(m.accounts) == 0 {
		return "No accounts configured"
	}
	name := m.accounts[m.selectedIdx]
	var lines []string
	lines = append(lines, titleStyle.Render("PREVIEW"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Account: %s", name))
	lines = append(lines, "")
	lines = append(lines, "Commands:")
	lines = append(lines, fmt.Sprintf("  aim run cc -a %s", name))
	if m.layout == LayoutSingle {
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Tab: switch to accounts"))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderStatusTab() string {
	return placeholderStyle.Render("Status tab - Coming soon")
}

func (m Model) renderFooter() string {
	help := "? Help  v Vendors  q Quit"
	return footerStyle.Render(help)
}
