package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#7D56F4")
	textColor      = lipgloss.Color("#FFFFFF")
	dimColor       = lipgloss.Color("#666666")

	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	tabStyle         = lipgloss.NewStyle().Padding(0, 2).Foreground(dimColor)
	activeTabStyle   = lipgloss.NewStyle().Padding(0, 2).Bold(true).Foreground(primaryColor).Underline(true)
	selectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	helpStyle        = lipgloss.NewStyle().Foreground(dimColor).Italic(true)
	footerStyle      = lipgloss.NewStyle().Foreground(dimColor).Padding(1, 0)
	leftPanelStyle   = lipgloss.NewStyle().Width(30).Padding(1)
	rightPanelStyle  = lipgloss.NewStyle().Padding(1)
	placeholderStyle = lipgloss.NewStyle().Foreground(dimColor).Padding(2)
)
