package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectableListModel encapsulates reusable cursor navigation, scrolling, window sizing,
// and key event routing for interactive Bubbletea item checklists.
type SelectableListModel struct {
	TotalItems   int
	Cursor       int
	Offset       int
	MaxVisible   int
	Confirmed    bool
	Canceled     bool
	WindowHeight int
}

// NewSelectableListModel constructs a new SelectableListModel pointer.
func NewSelectableListModel(totalItems int) *SelectableListModel {
	return &SelectableListModel{
		TotalItems: totalItems,
		Cursor:     0,
		Offset:     0,
		MaxVisible: 10,
		Confirmed:  false,
		Canceled:   false,
	}
}

// SetCursor updates the cursor position and adjusts the scrolling viewport offset.
func (m *SelectableListModel) SetCursor(index int) {
	if m.TotalItems == 0 {
		m.Cursor = 0
		m.Offset = 0
		return
	}

	if index < 0 {
		index = 0
	}
	if index >= m.TotalItems {
		index = m.TotalItems - 1
	}
	m.Cursor = index

	if m.MaxVisible <= 0 {
		m.MaxVisible = 10
	}

	if m.Cursor < m.Offset {
		m.Offset = m.Cursor
	} else if m.Cursor >= m.Offset+m.MaxVisible {
		m.Offset = m.Cursor - m.MaxVisible + 1
	}

	if m.Offset < 0 {
		m.Offset = 0
	}
}

// HandleMessage updates window height or cursor position based on tea.Msg navigation key events.
// Returns handled (bool) indicating whether a standard navigation/confirm/cancel key was processed.
func (m *SelectableListModel) HandleMessage(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowHeight = msg.Height
		maxVis := msg.Height - 12
		if maxVis < 5 {
			maxVis = 5
		}
		m.MaxVisible = maxVis
		m.SetCursor(m.Cursor)
		return true, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.Canceled = true
			return true, tea.Quit
		case tea.KeyUp:
			m.SetCursor(m.Cursor - 1)
			return true, nil
		case tea.KeyDown:
			m.SetCursor(m.Cursor + 1)
			return true, nil
		case tea.KeyPgUp:
			m.SetCursor(m.Cursor - m.MaxVisible)
			return true, nil
		case tea.KeyPgDown:
			m.SetCursor(m.Cursor + m.MaxVisible)
			return true, nil
		case tea.KeyHome:
			m.SetCursor(0)
			return true, nil
		case tea.KeyEnd:
			m.SetCursor(m.TotalItems - 1)
			return true, nil
		case tea.KeyEnter:
			m.Confirmed = true
			return true, tea.Quit
		}

		key := strings.ToLower(msg.String())
		switch key {
		case "q", "esc", "ctrl+c":
			m.Canceled = true
			return true, tea.Quit
		case "up", "k":
			m.SetCursor(m.Cursor - 1)
			return true, nil
		case "down", "j":
			m.SetCursor(m.Cursor + 1)
			return true, nil
		case "pgup", "b":
			m.SetCursor(m.Cursor - m.MaxVisible)
			return true, nil
		case "pgdown", "f":
			m.SetCursor(m.Cursor + m.MaxVisible)
			return true, nil
		case "g", "home":
			m.SetCursor(0)
			return true, nil
		case "G", "end":
			m.SetCursor(m.TotalItems - 1)
			return true, nil
		case "enter":
			m.Confirmed = true
			return true, tea.Quit
		}
	}
	return false, nil
}
