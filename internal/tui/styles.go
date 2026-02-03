package tui

import "github.com/charmbracelet/lipgloss"

// Tokyo Night Terminal Color Scheme
// https://github.com/folke/tokyonight.nvim
var (
	// ANSI Colors (0-15)
	ansiBlack         = lipgloss.Color("#1d202f")  // palette 0
	ansiRed           = lipgloss.Color("#f7768e")  // palette 1
	ansiGreen         = lipgloss.Color("#9ece6a")  // palette 2
	ansiYellow        = lipgloss.Color("#e0af68")  // palette 3
	ansiBlue          = lipgloss.Color("#7aa2f7")  // palette 4
	ansiMagenta       = lipgloss.Color("#bb9af7")  // palette 5
	ansiCyan          = lipgloss.Color("#7dcfff")  // palette 6
	ansiWhite         = lipgloss.Color("#a9b1d6")  // palette 7
	ansiBrightBlack   = lipgloss.Color("#414868")  // palette 8
	ansiBrightRed     = lipgloss.Color("#f7768e")  // palette 9
	ansiBrightGreen   = lipgloss.Color("#9ece6a")  // palette 10
	ansiBrightYellow  = lipgloss.Color("#e0af68")  // palette 11
	ansiBrightBlue    = lipgloss.Color("#7aa2f7")  // palette 12
	ansiBrightMagenta = lipgloss.Color("#bb9af7")  // palette 13
	ansiBrightCyan    = lipgloss.Color("#7dcfff")  // palette 14
	ansiBrightWhite   = lipgloss.Color("#c0caf5")  // palette 15

	// Terminal Colors
	terminalBackground       = lipgloss.Color("#24283b")
	terminalForeground       = lipgloss.Color("#c0caf5")
	terminalSelectionBg      = lipgloss.Color("#2e3c64")
	terminalSelectionFg      = lipgloss.Color("#c0caf5")
	terminalCursorColor      = lipgloss.Color("#c0caf5")
	terminalCursorText       = lipgloss.Color("#24283b")

	// UI Colors (Tokyo Night Storm/Day variants available)
	bgDark       = lipgloss.Color("#1f2335")  // Darker background for contrast
	bgFloat      = lipgloss.Color("#1f2335")  // Floating window background
	bgHighlight  = lipgloss.Color("#292e42")  // Highlight background
	bgPopup      = lipgloss.Color("#1f2335")  // Popup menu background
	bgSearch     = lipgloss.Color("#3d59a1")  // Search highlight
	bgVisual     = lipgloss.Color("#2e3c64")  // Visual selection
	bgStatusline = lipgloss.Color("#1f2335")  // Status line background

	// Foreground variants
	fgDark       = lipgloss.Color("#a9b1d6")  // Darker foreground
	fgGutter     = lipgloss.Color("#3b4261")  // Line number foreground
	fgComment    = lipgloss.Color("#565f89")  // Comment color

	// Semantic Colors (Tokyo Night naming)
	red       = ansiRed       // #f7768e - Errors, deletes
	orange    = lipgloss.Color("#ff9e64") // Warnings, constants
	yellow    = ansiYellow    // #e0af68 - Warnings, functions
	green     = ansiGreen     // #9ece6a - Success, strings
	teal      = lipgloss.Color("#1abc9c") // Types, special
	cyan      = ansiCyan      // #7dcfff - Info, fields
	blue      = ansiBlue      // #7aa2f7 - Keywords, links
	magenta   = ansiMagenta   // #bb9af7 - Keywords, special
	purple    = lipgloss.Color("#9d7cd8") // Functions, methods

	// Diff colors
	diffAdd    = lipgloss.Color("#283b4d")
	diffChange = lipgloss.Color("#394b70")
	diffDelete = lipgloss.Color("#3f2d3d")
	diffText   = lipgloss.Color("#2e3c64")

	// Border colors
	borderHighlight = lipgloss.Color("#29a4bd")
	borderFocus     = blue
	borderMuted     = lipgloss.Color("#273849")

	// Git colors
	gitAdd        = green
	gitChange     = yellow
	gitDelete     = red
	gitConflict   = orange
	gitIgnored    = fgComment
	gitUntracked  = fgDark

	// Special
	none       = lipgloss.Color("")
	normal     = terminalForeground
	comment    = fgComment
	whitespace = lipgloss.Color("#363b54")
)

// UI Styles - Tokyo Night Theme
var (
	// Background fills the entire terminal
	backgroundStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(terminalForeground)

	// Title styles
	titleStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Bold(true).
		Foreground(blue)

	subtitleStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(cyan)

	// Tab styles
	tabStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Padding(0, 2).
		Foreground(fgComment)

	activeTabStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Padding(0, 2).
		Bold(true).
		Foreground(blue).
		Underline(true)

	tabSeparatorStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgGutter)

	// Selection styles
	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(terminalForeground).
		Background(terminalSelectionBg)

	selectedInactiveStyle = lipgloss.NewStyle().
		Foreground(terminalForeground).
		Background(bgHighlight)

	// Help and footer
	helpStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Italic(true)

	helpKeyStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Bold(true).
		Foreground(yellow)

	footerStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Padding(1, 0)

	// Panel styles
	leftPanelStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Width(30).
		Padding(1)

	rightPanelStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Padding(1)

	panelBorderStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderMuted)

	panelBorderActiveStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(blue)

	// Placeholder and empty states
	placeholderStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Padding(2)

	emptyStateStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Italic(true)

	// Status indicators
	statusOKStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(green)

	statusWarnStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(yellow)

	statusErrorStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(red)

	statusInfoStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(cyan)

	statusPendingStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment)

	// Key/value display
	keyStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(cyan)

	valueStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(terminalForeground)

	valueMutedStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment)

	// Edit mode
	editLabelStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(magenta)

	editInputStyle = lipgloss.NewStyle().
		Background(bgHighlight).
		Foreground(terminalForeground).
		Padding(0, 1)

	editCursorStyle = lipgloss.NewStyle().
		Background(terminalCursorColor).
		Foreground(terminalCursorText)

	// List styles
	listItemStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(terminalForeground)

	listItemSelectedStyle = lipgloss.NewStyle().
		Background(terminalSelectionBg).
		Foreground(terminalSelectionFg).
		Bold(true)

	listItemDimStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
		Background(bgHighlight).
		Foreground(terminalForeground).
		Padding(0, 2).
		Margin(0, 1)

	buttonActiveStyle = lipgloss.NewStyle().
		Background(blue).
		Foreground(terminalBackground).
		Padding(0, 2).
		Margin(0, 1).
		Bold(true)

	// Scrollbar
	scrollbarStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgGutter)

	scrollbarThumbStyle = lipgloss.NewStyle().
		Background(fgComment).
		Foreground(terminalForeground)

	// Syntax highlighting (for code preview)
	syntaxKeywordStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(magenta)

	syntaxFunctionStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(blue)

	syntaxStringStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(green)

	syntaxCommentStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Italic(true)

	syntaxNumberStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(orange)

	syntaxOperatorStyle = lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(red)
)

// Icons
const (
	successIcon  = "✓"
	warningIcon  = "⚠"
	errorIcon    = "✗"
	infoIcon     = "ℹ"
	pendingIcon  = "○"
	selectedIcon = ">"
	unselectedIcon = " "
)

// Helper functions for dynamic styles

// WithBackground returns a style with the terminal background
func WithBackground(s lipgloss.Style) lipgloss.Style {
	return s.Background(terminalBackground)
}

// Dim returns dimmed text style
func Dim(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(fgComment).
		Render(s)
}

// Highlight returns highlighted text style
func Highlight(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(yellow).
		Bold(true).
		Render(s)
}

// Success returns success text style
func Success(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(green).
		Render(s)
}

// Error returns error text style
func Error(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(red).
		Render(s)
}

// Warning returns warning text style
func Warning(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(yellow).
		Render(s)
}

// Info returns info text style
func Info(s string) string {
	return lipgloss.NewStyle().
		Background(terminalBackground).
		Foreground(cyan).
		Render(s)
}
