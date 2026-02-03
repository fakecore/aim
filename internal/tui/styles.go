package tui

import "github.com/charmbracelet/lipgloss"

// Tokyo Night theme colors
var (
	background  = lipgloss.Color("#24283b")
	foreground  = lipgloss.Color("#c0caf5")
	selectionBg = lipgloss.Color("#2e3c64")
	red         = lipgloss.Color("#f7768e")
	green       = lipgloss.Color("#9ece6a")
	yellow      = lipgloss.Color("#e0af68")
	blue        = lipgloss.Color("#7aa2f7")
	magenta     = lipgloss.Color("#bb9af7")
	cyan        = lipgloss.Color("#7dcfff")
	brightBlack = lipgloss.Color("#414868")

	// Base style with background
	baseStyle = lipgloss.NewStyle().Background(background)

	titleStyle       = baseStyle.Copy().Bold(true).Foreground(blue)
	tabStyle         = baseStyle.Copy().Padding(0, 2).Foreground(brightBlack)
	activeTabStyle   = baseStyle.Copy().Padding(0, 2).Bold(true).Foreground(blue).Underline(true)
	selectedStyle    = baseStyle.Copy().Bold(true).Foreground(foreground).Background(selectionBg)
	helpStyle        = baseStyle.Copy().Foreground(brightBlack).Italic(true)
	footerStyle      = baseStyle.Copy().Foreground(brightBlack).Padding(1, 0)
	leftPanelStyle   = baseStyle.Copy().Width(30).Padding(1)
	rightPanelStyle  = baseStyle.Copy().Padding(1)
	placeholderStyle = baseStyle.Copy().Foreground(brightBlack).Padding(2)

	// Status indicators
	statusOKStyle    = baseStyle.Copy().Foreground(green)
	statusWarnStyle  = baseStyle.Copy().Foreground(yellow)
	statusErrorStyle = baseStyle.Copy().Foreground(red)

	// Key/value display
	keyStyle   = baseStyle.Copy().Foreground(cyan)
	valueStyle = baseStyle.Copy().Foreground(foreground)

	// Edit mode
	editLabelStyle = baseStyle.Copy().Foreground(magenta)
	editInputStyle = baseStyle.Copy().Foreground(foreground).Background(selectionBg)
)
