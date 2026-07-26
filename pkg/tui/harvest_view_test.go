package tui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/korakotlee/koharness/pkg/tui"
)

func TestHarvestModel_KeyNavigationAndToggle(t *testing.T) {
	items := []harvester.DiscoveredCapability{
		{
			HarnessID:  harness.HarnessAntigravity,
			Type:       harness.CapabilitySkill,
			Name:       "a11y",
			SourcePath: "/path/to/a11y",
			Selected:   true,
			IsSecret:   false,
		},
		{
			HarnessID:  harness.HarnessClaude,
			Type:       harness.CapabilityMCP,
			Name:       "github",
			SourcePath: "/path/to/claude.json",
			Selected:   false,
			IsSecret:   true,
		},
		{
			HarnessID:  harness.HarnessCodex,
			Type:       harness.CapabilityPrompt,
			Name:       "review.md",
			SourcePath: "/path/to/review.md",
			Selected:   true,
			IsSecret:   false,
		},
	}

	model := tui.NewHarvestModel(items)

	if model.Cursor != 0 {
		t.Errorf("expected initial cursor 0, got %d", model.Cursor)
	}

	// Initial view check
	v0 := model.View()
	if !strings.Contains(v0, "Selected: 2/3 capabilities") {
		t.Errorf("expected header counter 'Selected: 2/3 capabilities' in view output")
	}

	// Move down using KeyDown
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(*tui.HarvestModel)
	if m.Cursor != 1 {
		t.Errorf("expected cursor 1 after KeyDown, got %d", m.Cursor)
	}

	// Move down using 'j'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(*tui.HarvestModel)
	if m.Cursor != 2 {
		t.Errorf("expected cursor 2 after 'j', got %d", m.Cursor)
	}

	// Move up using 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(*tui.HarvestModel)
	if m.Cursor != 1 {
		t.Errorf("expected cursor 1 after 'k', got %d", m.Cursor)
	}

	// Toggle selection on item 1 with space key
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(*tui.HarvestModel)
	if !m.Items[1].Selected {
		t.Errorf("expected item 1 Selected to be true after space")
	}

	// Header should update to 3/3
	v1 := m.View()
	if !strings.Contains(v1, "Selected: 3/3 capabilities") {
		t.Errorf("expected header counter 'Selected: 3/3 capabilities' after toggle")
	}

	// Toggle secret override with 's'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(*tui.HarvestModel)
	if m.Items[1].IsSecret {
		t.Errorf("expected item 1 IsSecret to be false after 's'")
	}

	// Confirm with enter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*tui.HarvestModel)
	if !m.Confirmed {
		t.Errorf("expected model Confirmed to be true after Enter")
	}
}

func TestHarvestModel_LongListViewportScrolling(t *testing.T) {
	// Create 25 capabilities
	var items []harvester.DiscoveredCapability
	for i := 0; i < 25; i++ {
		items = append(items, harvester.DiscoveredCapability{
			HarnessID:  harness.HarnessAntigravity,
			Type:       harness.CapabilityWorkflow,
			Name:       fmt.Sprintf("workflow-%d.md", i),
			SourcePath: fmt.Sprintf("/path/to/workflow-%d.md", i),
			Selected:   true,
			IsSecret:   false,
		})
	}

	model := tui.NewHarvestModel(items)
	model.MaxVisible = 5

	if model.Offset != 0 {
		t.Errorf("expected initial offset 0, got %d", model.Offset)
	}

	// Move cursor down past MaxVisible (5)
	for i := 0; i < 6; i++ {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updated.(*tui.HarvestModel)
	}

	if model.Cursor != 6 {
		t.Errorf("expected cursor 6, got %d", model.Cursor)
	}
	if model.Offset != 2 {
		t.Errorf("expected offset 2 after moving cursor to 6 with MaxVisible 5, got %d", model.Offset)
	}

	viewStr := model.View()
	if !strings.Contains(viewStr, "▲ 2 more item(s) above...") {
		t.Errorf("expected top scroll indicator '▲ 2 more item(s) above...' in view")
	}
	if !strings.Contains(viewStr, "▼ 18 more item(s) below...") {
		t.Errorf("expected bottom scroll indicator '▼ 18 more item(s) below...' in view")
	}

	// Test toggle all with 'a'
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(*tui.HarvestModel)
	for _, item := range model.Items {
		if item.Selected {
			t.Errorf("expected all items to be unselected after pressing 'a'")
		}
	}
}
