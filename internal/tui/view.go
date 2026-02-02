package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fakecore/aim/internal/config"
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

	// Edit mode
	if m.editMode == EditName {
		lines = append(lines, "")
		lines = append(lines, "New account name:")
		lines = append(lines, m.editValue+"_")
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("n: new  e: edit  d: delete  q: quit"))

	return strings.Join(lines, "\n")
}

func (m Model) renderPreview() string {
	if len(m.accounts) == 0 {
		return placeholderStyle.Render("No accounts configured\n\nPress 'n' to create one")
	}

	name := m.accounts[m.selectedIdx]
	acc := m.config.Accounts[name]

	var lines []string
	lines = append(lines, titleStyle.Render("LIVE PREVIEW"))
	lines = append(lines, "")

	// Account info
	lines = append(lines, fmt.Sprintf("Account: %s", name))
	lines = append(lines, fmt.Sprintf("Vendor: %s", acc.Vendor))
	if acc.Vendor == "" {
		lines = append(lines, fmt.Sprintf("  (inferred: %s)", name))
	}
	lines = append(lines, "")

	// Supported tools
	lines = append(lines, "Supported tools:")
	lines = append(lines, "")

	// claude-code
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("claude-code"))
	lines = append(lines, fmt.Sprintf("  $ aim run cc -a %s", name))
	if acc.Key != "" {
		key, _ := config.ResolveKey(acc.Key)
		lines = append(lines, fmt.Sprintf("  ANTHROPIC_AUTH_TOKEN=%s...", truncate(key, 16)))
	}
	lines = append(lines, "")

	// codex
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("codex"))
	lines = append(lines, fmt.Sprintf("  $ aim run codex -a %s", name))
	lines = append(lines, "")

	if m.layout == LayoutSingle {
		lines = append(lines, helpStyle.Render("Tab: switch to accounts"))
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (m Model) renderStatusTab() string {
	return placeholderStyle.Render("Status tab - Coming soon")
}

func (m Model) renderFooter() string {
	help := "? Help  v Vendors  q Quit"
	return footerStyle.Render(help)
}
