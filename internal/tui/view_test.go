package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
)

// TestWindowSizeMsgHandling tests that WindowSizeMsg correctly updates model dimensions
func TestWindowSizeMsgHandling(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	m := New(cfg)

	// Verify initial state
	if m.width != 0 {
		t.Errorf("Expected initial width to be 0, got %d", m.width)
	}
	if m.height != 0 {
		t.Errorf("Expected initial height to be 0, got %d", m.height)
	}
	if m.layout != LayoutSingle {
		t.Errorf("Expected initial layout to be LayoutSingle, got %d", m.layout)
	}

	// Send WindowSizeMsg
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	m = newM.(Model)

	// Verify state after WindowSizeMsg
	if m.width != 60 {
		t.Errorf("Expected width to be 60, got %d", m.width)
	}
	if m.height != 20 {
		t.Errorf("Expected height to be 20, got %d", m.height)
	}
	if m.layout != LayoutSingle {
		t.Errorf("Expected layout to be LayoutSingle for 60x20, got %d", m.layout)
	}

	// Test view renders correctly
	view := m.View()
	if !strings.Contains(view, "Config") {
		t.Error("Expected View to contain 'Config' tab")
	}
	if !strings.Contains(view, "ACCOUNTS") {
		t.Error("Expected View to contain 'ACCOUNTS'")
	}
}

// TestWindowSizeMsgSplitLayout tests split layout at 80+ width
func TestWindowSizeMsgSplitLayout(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	m := New(cfg)

	// Send WindowSizeMsg for split layout
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = newM.(Model)

	// Verify state
	if m.width != 100 {
		t.Errorf("Expected width to be 100, got %d", m.width)
	}
	if m.height != 30 {
		t.Errorf("Expected height to be 30, got %d", m.height)
	}
	if m.layout != LayoutSplit {
		t.Errorf("Expected layout to be LayoutSplit for 100x30, got %d", m.layout)
	}

	// Test view renders correctly with both panels
	view := m.View()
	if !strings.Contains(view, "ACCOUNTS") {
		t.Error("Expected View to contain 'ACCOUNTS'")
	}
	if !strings.Contains(view, "LIVE PREVIEW") {
		t.Error("Expected View to contain 'LIVE PREVIEW' in split layout")
	}
}

// TestWindowSizeMsgUnsupportedLayout tests unsupported layout handling
func TestWindowSizeMsgUnsupportedLayout(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	m := New(cfg)

	// Send WindowSizeMsg for unsupported layout (too small)
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 8})
	m = newM.(Model)

	// Verify state
	if m.width != 30 {
		t.Errorf("Expected width to be 30, got %d", m.width)
	}
	if m.height != 8 {
		t.Errorf("Expected height to be 8, got %d", m.height)
	}
	if m.layout != LayoutUnsupported {
		t.Errorf("Expected layout to be LayoutUnsupported for 30x8, got %d", m.layout)
	}

	// Test view shows unsupported message
	view := m.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Error("Expected View to show 'Terminal too small' message")
	}
}

// TestTabNavigation tests switching between tabs
func TestTabNavigation(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"deepseek": {Key: "sk-test", Vendor: "deepseek"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Verify initial tab
	if m.activeTab != TabConfig {
		t.Errorf("Expected initial tab to be TabConfig, got %d", m.activeTab)
	}

	// Navigate right to Status tab
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = newM.(Model)
	if m.activeTab != TabStatus {
		t.Errorf("Expected tab to be TabStatus after right arrow, got %d", m.activeTab)
	}

	// Navigate right to Routes tab
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = newM.(Model)
	if m.activeTab != TabRoutes {
		t.Errorf("Expected tab to be TabRoutes after right arrow, got %d", m.activeTab)
	}

	// Navigate left back to Status
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = newM.(Model)
	if m.activeTab != TabStatus {
		t.Errorf("Expected tab to be TabStatus after left arrow, got %d", m.activeTab)
	}

	// Navigate with 'h' key (vim left)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = newM.(Model)
	if m.activeTab != TabConfig {
		t.Errorf("Expected tab to be TabConfig after 'h' key, got %d", m.activeTab)
	}

	// Navigate with 'l' key (vim right)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = newM.(Model)
	if m.activeTab != TabStatus {
		t.Errorf("Expected tab to be TabStatus after 'l' key, got %d", m.activeTab)
	}
}

// TestTabNavigationBoundary tests tab navigation boundaries
func TestTabNavigationBoundary(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Navigate left from first tab (should stay at Config)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = newM.(Model)
	if m.activeTab != TabConfig {
		t.Errorf("Expected tab to stay at TabConfig when at left boundary, got %d", m.activeTab)
	}

	// Navigate to last tab (Logs)
	m.activeTab = TabLogs

	// Navigate right from last tab (should stay at Logs)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = newM.(Model)
	if m.activeTab != TabLogs {
		t.Errorf("Expected tab to stay at TabLogs when at right boundary, got %d", m.activeTab)
	}
}

// TestAccountNavigation tests navigating the account list
func TestAccountNavigation(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
			"account2": {Key: "key2", Vendor: "vendor2"},
			"account3": {Key: "key3", Vendor: "vendor3"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Verify initial selection
	if m.selectedIdx != 0 {
		t.Errorf("Expected initial selectedIdx to be 0, got %d", m.selectedIdx)
	}

	// Navigate down with 'j' key
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newM.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx to be 1 after 'j', got %d", m.selectedIdx)
	}

	// Navigate down with arrow key
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	if m.selectedIdx != 2 {
		t.Errorf("Expected selectedIdx to be 2 after down arrow, got %d", m.selectedIdx)
	}

	// Navigate up with 'k' key
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newM.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx to be 1 after 'k', got %d", m.selectedIdx)
	}

	// Navigate up with arrow key
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx to be 0 after up arrow, got %d", m.selectedIdx)
	}
}

// TestAccountNavigationBoundary tests account navigation boundaries
func TestAccountNavigationBoundary(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
			"account2": {Key: "key2", Vendor: "vendor2"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Navigate up from first account (should stay at 0)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx to stay at 0 when at top boundary, got %d", m.selectedIdx)
	}

	// Navigate to last account
	m.selectedIdx = 1

	// Navigate down from last account (should stay at last index)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	if m.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx to stay at 1 when at bottom boundary, got %d", m.selectedIdx)
	}
}

// TestQuitKey tests quitting with 'q' key
func TestQuitKey(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)

	// Press 'q' to quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Verify quit command is returned
	if cmd == nil {
		t.Error("Expected quit command to be returned for 'q' key")
	}
}

// TestCtrlCQuit tests quitting with Ctrl+C
func TestCtrlCQuit(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)

	// Press Ctrl+C to quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Verify quit command is returned
	if cmd == nil {
		t.Error("Expected quit command to be returned for Ctrl+C")
	}
}

// TestCreateAccount tests creating a new account
func TestCreateAccount(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Press 'n' to enter edit mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newM.(Model)

	if m.editMode != EditName {
		t.Errorf("Expected editMode to be EditName, got %d", m.editMode)
	}

	// Type account name (each character separately)
	for _, ch := range []rune{'a', 'b', 'c'} {
		newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = newM.(Model)
	}

	if m.editValue != "abc" {
		t.Errorf("Expected editValue to be 'abc', got '%s'", m.editValue)
	}

	// Press Enter to confirm
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)

	if m.editMode != EditNone {
		t.Errorf("Expected editMode to be EditNone after confirm, got %d", m.editMode)
	}

	// Verify account was created
	if _, exists := m.config.Accounts["abc"]; !exists {
		t.Error("Expected account 'abc' to be created")
	}

	// Verify account is in list
	found := false
	for _, name := range m.accounts {
		if name == "abc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected account 'abc' to be in accounts list")
	}
}

// TestCancelEdit tests canceling edit mode with Escape
func TestCancelEdit(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Enter edit mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newM.(Model)

	// Type something (each character separately)
	for _, ch := range []rune{'t', 'e', 's', 't'} {
		newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = newM.(Model)
	}

	if m.editValue != "test" {
		t.Errorf("Expected editValue to be 'test', got '%s'", m.editValue)
	}

	// Press Escape to cancel
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newM.(Model)

	if m.editMode != EditNone {
		t.Errorf("Expected editMode to be EditNone after cancel, got %d", m.editMode)
	}

	if m.editValue != "" {
		t.Errorf("Expected editValue to be cleared after cancel, got '%s'", m.editValue)
	}
}

// TestBackspaceInEditMode tests backspace functionality
func TestBackspaceInEditMode(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)

	// Enter edit mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = newM.(Model)

	// Type something (each character separately)
	for _, ch := range []rune{'h', 'e', 'l', 'l', 'o'} {
		newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = newM.(Model)
	}

	if m.editValue != "hello" {
		t.Errorf("Expected editValue to be 'hello', got '%s'", m.editValue)
	}

	// Press backspace
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = newM.(Model)

	if m.editValue != "hell" {
		t.Errorf("Expected editValue to be 'hell' after backspace, got '%s'", m.editValue)
	}

	// Press backspace again
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = newM.(Model)

	if m.editValue != "hel" {
		t.Errorf("Expected editValue to be 'hel' after backspace, got '%s'", m.editValue)
	}
}

// TestDeleteAccount tests deleting an account
func TestDeleteAccount(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
			"account2": {Key: "key2", Vendor: "vendor2"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Verify initial state
	if len(m.accounts) != 2 {
		t.Errorf("Expected 2 accounts initially, got %d", len(m.accounts))
	}

	// Press 'd' to delete selected account
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = newM.(Model)

	// Verify account was deleted
	if len(m.accounts) != 1 {
		t.Errorf("Expected 1 account after delete, got %d", len(m.accounts))
	}

	if _, exists := m.config.Accounts["account1"]; exists {
		t.Error("Expected account1 to be deleted from config")
	}
}

// TestDeleteLastAccount tests deleting the last account
func TestDeleteLastAccount(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"only": {Key: "key", Vendor: "vendor"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	// Press 'd' to delete the only account
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = newM.(Model)

	// Verify account was deleted
	if len(m.accounts) != 0 {
		t.Errorf("Expected 0 accounts after delete, got %d", len(m.accounts))
	}

	// Verify selectedIdx is still valid (0)
	if m.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx to be 0, got %d", m.selectedIdx)
	}
}

// TestSingleLayoutPreviewToggle tests toggling preview in single layout
func TestSingleLayoutPreviewToggle(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.activeTab = TabConfig

	// Verify initial state
	if m.showPreview {
		t.Error("Expected showPreview to be false initially")
	}

	// Press Tab to toggle preview
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if !m.showPreview {
		t.Error("Expected showPreview to be true after Tab")
	}

	// Press Tab again to toggle back
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if m.showPreview {
		t.Error("Expected showPreview to be false after second Tab")
	}
}

// TestTabDoesNotToggleInSplitLayout tests that Tab doesn't toggle in split layout
func TestTabDoesNotToggleInSplitLayout(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
		},
	}

	m := New(cfg)
	m.width = 100
	m.height = 30
	m.layout = LayoutSplit
	m.activeTab = TabConfig

	// Press Tab - should not toggle preview in split layout
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if m.showPreview {
		t.Error("Expected showPreview to stay false in split layout")
	}
}

// TestTabDoesNotToggleInOtherTabs tests that Tab doesn't toggle in non-Config tabs
func TestTabDoesNotToggleInOtherTabs(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{
			"account1": {Key: "key1", Vendor: "vendor1"},
		},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.activeTab = TabStatus

	// Press Tab - should not toggle preview when not on Config tab
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if m.showPreview {
		t.Error("Expected showPreview to stay false when not on Config tab")
	}
}

// TestViewRendersTabs tests that view renders all tabs
func TestViewRendersTabs(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	view := m.View()

	tabs := []string{"Config", "Status", "Routes", "Usage", "Logs"}
	for _, tab := range tabs {
		if !strings.Contains(view, tab) {
			t.Errorf("Expected view to contain tab '%s'", tab)
		}
	}
}

// TestViewRendersFooter tests that view renders footer
func TestViewRendersFooter(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	view := m.View()

	if !strings.Contains(view, "? Help") {
		t.Error("Expected view to contain '? Help' in footer")
	}
	if !strings.Contains(view, "v Vendors") {
		t.Error("Expected view to contain 'v Vendors' in footer")
	}
	if !strings.Contains(view, "q Quit") {
		t.Error("Expected view to contain 'q Quit' in footer")
	}
}

// TestViewRendersAccountListHelp tests that view renders account list help
func TestViewRendersAccountListHelp(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle

	view := m.View()

	if !strings.Contains(view, "n: new") {
		t.Error("Expected view to contain 'n: new' help text")
	}
	if !strings.Contains(view, "e: edit") {
		t.Error("Expected view to contain 'e: edit' help text")
	}
	if !strings.Contains(view, "d: delete") {
		t.Error("Expected view to contain 'd: delete' help text")
	}
}

// TestEmptyAccountsView tests view with no accounts
func TestEmptyAccountsView(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.showPreview = true // Show preview to see empty state message

	view := m.View()

	if !strings.Contains(view, "No accounts configured") {
		t.Error("Expected view to show 'No accounts configured' message")
	}
	if !strings.Contains(view, "Press 'n' to create one") {
		t.Error("Expected view to show prompt to create account")
	}
}

// TestStatusTabView tests Status tab view
func TestStatusTabView(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.activeTab = TabStatus

	view := m.View()

	if !strings.Contains(view, "STATUS") {
		t.Error("Expected view to show 'STATUS' title")
	}
}

// TestPlaceholderTabView tests placeholder tabs view
func TestPlaceholderTabView(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.activeTab = TabRoutes

	view := m.View()

	if !strings.Contains(view, "Coming soon...") {
		t.Error("Expected view to show 'Coming soon...' placeholder")
	}
}

// TestEditModeView tests view in edit mode
func TestEditModeView(t *testing.T) {
	cfg := &config.Config{
		Version: "2",
		Accounts: map[string]config.Account{},
	}

	m := New(cfg)
	m.width = 60
	m.height = 20
	m.layout = LayoutSingle
	m.editMode = EditName
	m.editValue = "test-account"

	view := m.View()

	if !strings.Contains(view, "New account name:") {
		t.Error("Expected view to show 'New account name:' prompt")
	}
	if !strings.Contains(view, "test-account") {
		t.Error("Expected view to show edit value")
	}
}

// TestTruncateFunction tests the truncate helper function
func TestTruncateFunction(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactlyten", 10, "exactlyten"},
		{"this is a long string", 5, "this "},
		{"", 5, ""},
		{"test", 0, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
