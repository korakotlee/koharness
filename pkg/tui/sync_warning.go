// Package tui provides terminal user interface components, views, styles, and interactive screens for KoHarness.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/git"
)

// SyncWarningModel represents the Bubbletea model for rendering dirty repository warning screens.
type SyncWarningModel struct {
	RepoPath   string
	DirtyFiles []git.DirtyFile
	Quitted    bool
}

// NewSyncWarningModel initializes a new SyncWarningModel.
func NewSyncWarningModel(repoPath string, dirtyFiles []git.DirtyFile) SyncWarningModel {
	return SyncWarningModel{
		RepoPath:   repoPath,
		DirtyFiles: dirtyFiles,
		Quitted:    false,
	}
}

// Init implements tea.Model.
func (m SyncWarningModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SyncWarningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			m.Quitted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m SyncWarningModel) View() string {
	var s strings.Builder

	s.WriteString(RenderBanner())
	s.WriteString("\n")
	s.WriteString(BadgeWarn("UNCOMMITTED LOCAL CHANGES DETECTED"))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Cannot safely sync upstream updates to repository:"))
	s.WriteString("\n  " + StyleViolet.Render(m.RepoPath) + "\n\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render("DIRTY FILES:"))
	s.WriteString("\n")

	fileBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAmber).
		Padding(0, 1)

	var filesList strings.Builder
	maxDisplay := 10
	for i, f := range m.DirtyFiles {
		if i >= maxDisplay {
			remaining := len(m.DirtyFiles) - maxDisplay
			filesList.WriteString(StyleMuted.Render(fmt.Sprintf("... and %d more file(s)", remaining)) + "\n")
			break
		}
		statusStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorRed)
		filesList.WriteString(fmt.Sprintf("  [%s] %s\n", statusStyle.Render(f.Status), f.Path))
	}

	s.WriteString(fileBoxStyle.Render(strings.TrimRight(filesList.String(), "\n")))
	s.WriteString("\n\n")

	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("RECOMMENDED ACTIONS:"))
	s.WriteString("\n")
	s.WriteString("  1. Commit your local changes:\n")
	s.WriteString("     " + StyleMuted.Render("cd "+m.RepoPath+" && git add . && git commit -m \"local changes\"") + "\n")
	s.WriteString("  2. Create a Pull Request or push your branch before re-running sync.\n\n")

	s.WriteString(StyleMuted.Render("Press [Enter] or [q] to exit without pulling updates."))
	s.WriteString("\n")

	return s.String()
}

// RunSyncWarningView launches the interactive Bubbletea TUI warning view.
func RunSyncWarningView(repoPath string, dirtyFiles []git.DirtyFile) error {
	m := NewSyncWarningModel(repoPath, dirtyFiles)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// RenderSyncWarningText outputs a non-interactive plain text warning for non-TTY or automated execution modes.
func RenderSyncWarningText(repoPath string, dirtyFiles []git.DirtyFile) string {
	var s strings.Builder
	s.WriteString(BadgeWarn("UNCOMMITTED LOCAL CHANGES DETECTED") + "\n")
	s.WriteString(fmt.Sprintf("Repository path: %s\n", repoPath))
	s.WriteString("Dirty Files:\n")
	for _, f := range dirtyFiles {
		s.WriteString(fmt.Sprintf("  [%s] %s\n", f.Status, f.Path))
	}
	s.WriteString("\nAction Required: Commit or stash local changes before running sync.\n")
	return s.String()
}
