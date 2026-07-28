package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/korakotlee/koharness/pkg/tui"
)

func TestInitModel_NavigationAndToggle(t *testing.T) {
	items := []tui.InitCapabilityItem{
		{Type: "skill", Name: "skill-one", HarnessID: "gemini", Selected: true, HasConflict: false},
		{Type: "prompt", Name: "prompt-two", HarnessID: "claude", Selected: false, HasConflict: true},
	}

	model := tui.NewInitModel(items)

	// Verify initial state
	if model.Cursor != 0 {
		t.Errorf("expected initial cursor 0, got %d", model.Cursor)
	}

	// Send KeyDown msg
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(*tui.InitModel)

	if m.Cursor != 1 {
		t.Errorf("expected cursor 1 after down key, got %d", m.Cursor)
	}

	// Toggle item 1
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(*tui.InitModel)

	if !m.Items[1].Selected {
		t.Errorf("expected item 1 to be selected after space toggle")
	}

	// Check View string output contains CONFLICT badge
	viewOutput := m.View()
	if !strings.Contains(viewOutput, "CONFLICT") {
		t.Errorf("expected view output to render CONFLICT badge")
	}
	if !strings.Contains(viewOutput, "INIT SETUP & CAPABILITY LINKING") {
		t.Errorf("expected view output to render title header")
	}
}
