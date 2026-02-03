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

	titleStyle       = lipgloss.NewStyle().Background(background).Bold(true).Foreground(blue)
	tabStyle         = lipgloss.NewStyle().Background(background).Padding(0, 2).Foreground(brightBlack)
	activeTabStyle   = lipgloss.NewStyle().Background(background).Padding(0, 2).Bold(true).Foreground(blue).Underline(true)
	selectedStyle    = lipgloss.NewStyle().Background(background).Bold(true).Foreground(foreground).Background(selectionBg)
	helpStyle        = lipgloss.NewStyle().Background(background).Foreground(brightBlack).Italic(true)
	footerStyle      = lipgloss.NewStyle().Background(background).Foreground(brightBlack).Padding(1, 0)
	leftPanelStyle   = lipgloss.NewStyle().Background(background).Width(30).Padding(1)
	rightPanelStyle  = lipgloss.NewStyle().Background(background).Padding(1)
	placeholderStyle = lipgloss.NewStyle().Background(background).Foreground(brightBlack).Padding(2)

	// Status indicators
	statusOKStyle    = lipgloss.NewStyle().Background(background).Foreground(green)
	statusWarnStyle  = lipgloss.NewStyle().Background(background).Foreground(yellow)
	statusErrorStyle = lipgloss.NewStyle().Background(background).Foreground(red)

	// Key/value display
	keyStyle   = lipgloss.NewStyle().Background(background).Foreground(cyan)
	valueStyle = lipgloss.NewStyle().Background(background).Foreground(foreground)

	// Edit mode
	editLabelStyle = lipgloss.NewStyle().Background(background).Foreground(magenta)
	editInputStyle = lipgloss.NewStyle().Background(background).Foreground(foreground).Background(selectionBg)
)
