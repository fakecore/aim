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

	// Build the content
	var content string

	content += m.renderHeader()
	content += m.renderContent()
	content += m.renderFooter()

	// // Fill remaining height with background
	// availableHeight := m.height - lipgloss.Height(content)
	// if availableHeight > 0 {
	// 	filler := backgroundStyle.Height(availableHeight).Render("")
	// 	content = lipgloss.JoinVertical(lipgloss.Left, content, filler)
	// }

	// Ensure full width background (only if width is set)
	if m.width > 0 {
		return backgroundStyle.Width(m.width).Render(content)
	}
	return content
}

func (m Model) unsupportedView() string {
	msg := fmt.Sprintf(
		"Terminal too small\n\nCurrent: %d x %d\nMinimum: 40 x 10\n\nPlease resize and retry",
		m.width, m.height,
	)
	return backgroundStyle.Width(m.width).Height(m.height).Render(msg)
}

func (m Model) renderHeader() string {
	// Use abbreviated tab names on narrow screens
	var tabs []string
	if m.width < 50 {
		// Compact mode: use 2-3 letter abbreviations
		tabs = []string{"Cfg", "Sts", "Rt", "Us", "Lg"}
	} else {
		// Full mode: use complete names
		tabs = []string{"Config", "Status", "Routes", "Usage", "Logs"}
	}

	var rendered []string
	for i, tab := range tabs {
		style := tabStyle
		if Tab(i) == m.activeTab {
			style = activeTabStyle
		}
		rendered = append(rendered, style.Render(tab))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Left, rendered...)

	// Apply width and fixed height (1) to prevent content overflow covering header
	if m.width > 0 {
		header = backgroundStyle.Width(m.width).Render(header)
	}

	return header
}

func (m Model) renderContent() string {
	// Calculate available height for content
	// Header: tab height 1 (padding is included in the block, not extra)
	// Footer: 2 lines (1 content + 1 padding)
	headerHeight := 1
	footerHeight := 4
	availableHeight := m.height - headerHeight - footerHeight
	if availableHeight < 1 {
		availableHeight = 1
	}

	switch m.activeTab {
	case TabConfig:
		return m.renderConfigTab(availableHeight)
	case TabStatus:
		return m.renderStatusTab(availableHeight)
	default:
		return m.renderPlaceholderTab(availableHeight)
	}
}

func (m Model) renderConfigTab(height int) string {
	if m.layout == LayoutSplit {
		// Split layout: left panel fixed width, right panel takes remaining
		leftWidth := 30
		rightWidth := m.width - leftWidth

		left := m.renderAccountList(height)
		right := m.renderPreview(height)

		leftRendered := leftPanelStyle.Width(leftWidth).Height(height).Render(left)
		rightRendered := rightPanelStyle.Width(rightWidth).Height(height).Render(right)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, rightRendered)
	}

	// Single layout
	if m.showPreview {
		return rightPanelStyle.Width(m.width).Height(height).Render(m.renderPreview(height))
	}
	return leftPanelStyle.Width(m.width).Height(height).Render(m.renderAccountList(height))
}

func (m Model) renderAccountList(height int) string {
	var lines []string
	lines = append(lines, titleStyle.Render("ACCOUNTS"))
	lines = append(lines, "")

	for i, name := range m.accounts {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "> "
		}
		acc := m.config.Accounts[name]
		status := successIcon
		if acc.Key == "" {
			status = warningIcon
		}
		line := fmt.Sprintf("%s%s %s", prefix, status, name)
		if i == m.selectedIdx {
			line = selectedStyle.Width(28).Render(line)
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

	content := strings.Join(lines, "\n")

	// Fill remaining height
	contentHeight := strings.Count(content, "\n") + 1
	if contentHeight < height {
		filler := strings.Repeat("\n", height-contentHeight)
		content += filler
	}

	return content
}

func (m Model) renderPreview(height int) string {
	if len(m.accounts) == 0 {
		return placeholderStyle.Width(m.width - 4).Height(height - 2).Render(
			"No accounts configured\n\nPress 'n' to create one",
		)
	}

	name := m.accounts[m.selectedIdx]
	acc := m.config.Accounts[name]

	var lines []string
	lines = append(lines, titleStyle.Render("LIVE PREVIEW"))
	lines = append(lines, "")

	// Account info
	lines = append(lines, keyStyle.Render("Account: ")+valueStyle.Render(name))
	lines = append(lines, keyStyle.Render("Vendor:  ")+valueStyle.Render(acc.Vendor))
	lines = append(lines, "")

	// Supported tools
	lines = append(lines, titleStyle.Render("SUPPORTED TOOLS"))
	lines = append(lines, "")

	// claude-code
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(blue).Render("claude-code"))
	lines = append(lines, fmt.Sprintf("  $ aim run cc -a %s", name))
	if acc.Key != "" {
		key, _ := config.ResolveKey(acc.Key)
		lines = append(lines, fmt.Sprintf("  ANTHROPIC_AUTH_TOKEN=%s...", truncate(key, 16)))
	}
	lines = append(lines, "")

	// codex
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(green).Render("codex"))
	lines = append(lines, fmt.Sprintf("  $ aim run codex -a %s", name))
	lines = append(lines, "")

	if m.layout == LayoutSingle {
		lines = append(lines, helpStyle.Render("Tab: switch to accounts"))
	}

	content := strings.Join(lines, "\n")

	// Fill remaining height
	contentHeight := strings.Count(content, "\n")
	if contentHeight < height {
		filler := strings.Repeat("\n", height-contentHeight)
		content += filler
	}

	return content
}

func (m Model) renderStatusTab(height int) string {
	var lines []string
	lines = append(lines, titleStyle.Render("STATUS"))
	lines = append(lines, "")
	lines = append(lines, "Coming soon...")

	content := strings.Join(lines, "\n")

	// Fill remaining height
	contentHeight := strings.Count(content, "\n") + 1
	if contentHeight < height {
		filler := strings.Repeat("\n", height-contentHeight)
		content += filler
	}

	return leftPanelStyle.Width(m.width).Height(height).Render(content)
}

func (m Model) renderPlaceholderTab(height int) string {

	var lines []string
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "Coming soon...")

	content := strings.Join(lines, "\n")

	// Fill remaining height
	contentHeight := strings.Count(content, "\n") + 1
	if contentHeight < height {
		filler := strings.Repeat("\n", height-contentHeight)
		content += filler
	}

	return leftPanelStyle.Width(m.width).Height(height).Render(content)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (m Model) renderFooter() string {
	help := "? Help  v Vendors  q Quit"
	if m.width > 0 {
		return footerStyle.Width(m.width).Render(help)
	}
	return footerStyle.Render(help)
}
