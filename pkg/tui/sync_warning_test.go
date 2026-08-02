package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/korakotlee/koharness/pkg/git"
)

func TestSyncWarningModel_Update(t *testing.T) {
	m := NewSyncWarningModel("/path/to/repo", []git.DirtyFile{
		{Path: "pkg/git/status.go", Status: "M"},
	})

	// Test quitting via "q" key
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Errorf("expected tea.Quit command on key 'q'")
	}
	resModel, ok := updatedModel.(SyncWarningModel)
	if !ok {
		t.Fatalf("unexpected model type returned")
	}
	if !resModel.Quitted {
		t.Errorf("expected Quitted to be true after pressing 'q'")
	}
}

func TestSyncWarningModel_View(t *testing.T) {
	dirtyFiles := []git.DirtyFile{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.txt", Status: "??"},
	}
	m := NewSyncWarningModel("/tmp/test-repo", dirtyFiles)
	view := m.View()

	if !strings.Contains(view, "UNCOMMITTED LOCAL CHANGES DETECTED") {
		t.Errorf("expected view to contain header badge")
	}
	if !strings.Contains(view, "/tmp/test-repo") {
		t.Errorf("expected view to contain repo path")
	}
	if !strings.Contains(view, "file1.go") || !strings.Contains(view, "file2.txt") {
		t.Errorf("expected view to list dirty files")
	}
}

func TestRenderSyncWarningText(t *testing.T) {
	dirtyFiles := []git.DirtyFile{
		{Path: "file1.go", Status: "M"},
	}
	text := RenderSyncWarningText("/tmp/test-repo", dirtyFiles)

	if !strings.Contains(text, "UNCOMMITTED LOCAL CHANGES DETECTED") {
		t.Errorf("expected warning header in text output")
	}
	if !strings.Contains(text, "file1.go") {
		t.Errorf("expected file path in text output")
	}
}
