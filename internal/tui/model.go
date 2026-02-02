package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fakecore/aim/internal/config"
)

type Layout int

const (
	LayoutUnsupported Layout = iota
	LayoutSingle
	LayoutSplit
)

type Tab int

const (
	TabConfig Tab = iota
	TabStatus
	TabRoutes
	TabUsage
	TabLogs
)

type Model struct {
	width       int
	height      int
	layout      Layout
	activeTab   Tab
	config      *config.Config
	err         error
	accounts    []string
	selectedIdx int
	showPreview bool
}

func New(cfg *config.Config) Model {
	accounts := make([]string, 0, len(cfg.Accounts))
	for name := range cfg.Accounts {
		accounts = append(accounts, name)
	}
	return Model{
		config:      cfg,
		accounts:    accounts,
		selectedIdx: 0,
		activeTab:   TabConfig,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
