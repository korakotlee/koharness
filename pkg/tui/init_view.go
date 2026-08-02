// Package tui provides terminal user interface components for KoHarness,
// including styled banners, badges, and interactive Bubbletea views.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harvester"
)

// InitCapabilityItem represents a capability discovered in the repository to link into a client harness.
type InitCapabilityItem = harvester.RepoCapabilityItem

// InitModel represents the interactive Bubbletea TUI model for repository setup and capability linking.
type InitModel struct {
	*SelectableListModel
	Items []InitCapabilityItem
}

// NewInitModel constructs a new InitModel pointer with the provided repository capability items.
func NewInitModel(items []InitCapabilityItem) *InitModel {
	return &InitModel{
		SelectableListModel: NewSelectableListModel(len(items)),
		Items:               items,
	}
}

// Init initializes the Bubbletea model lifecycle.
func (m *InitModel) Init() tea.Cmd {
	return nil
}

// Update processes terminal input events and updates cursor and selection states.
func (m *InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.TotalItems = len(m.Items)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeySpace {
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
			}
			return m, nil
		}

		key := strings.ToLower(keyMsg.String())
		switch key {
		case "space", " ":
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
			}
			return m, nil
		case "a":
			allSelected := true
			for _, item := range m.Items {
				if !item.Selected {
					allSelected = false
					break
				}
			}
			for i := range m.Items {
				m.Items[i].Selected = !allSelected
			}
			return m, nil
		}
	}

	if handled, cmd := m.HandleMessage(msg); handled {
		return m, cmd
	}

	return m, nil
}

// View renders the interactive terminal user interface layout for repository initialization.
func (m *InitModel) View() string {
	var b strings.Builder

	b.WriteString(RenderBanner() + "\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorViolet).Render("INIT SETUP & CAPABILITY LINKING") + "\n")

	selectedCount := 0
	conflictCount := 0
	for _, item := range m.Items {
		if item.Selected {
			selectedCount++
		}
		if item.HasConflict {
			conflictCount++
		}
	}

	summaryStr := fmt.Sprintf("Selected: %d/%d capabilities", selectedCount, len(m.Items))
	if conflictCount > 0 {
		summaryStr += fmt.Sprintf(" (%d conflict(s) highlighted - will create .bak backups)", conflictCount)
	}
	b.WriteString(StyleMuted.Render(summaryStr) + "\n")
	b.WriteString(StyleMuted.Render("Controls: [↑/↓ or j/k] Navigate | [space] Toggle | [a] Toggle All | [enter] Confirm | [q] Cancel") + "\n\n")

	if len(m.Items) == 0 {
		b.WriteString(StyleMuted.Render("No capabilities discovered in repository.") + "\n\n")
		return b.String()
	}

	if m.MaxVisible <= 0 {
		m.MaxVisible = 10
	}

	if m.Offset > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(fmt.Sprintf("  ▲ %d more item(s) above...", m.Offset)) + "\n")
	} else {
		b.WriteString("\n")
	}

	end := m.Offset + m.MaxVisible
	if end > len(m.Items) {
		end = len(m.Items)
	}

	for i := m.Offset; i < end; i++ {
		item := m.Items[i]
		isFocused := m.Cursor == i

		cursorStr := "  "
		if isFocused {
			cursorStr = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("> ")
		}

		checkStr := "[ ]"
		if item.Selected {
			checkStr = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen).Render("[x]")
		}

		conflictBadge := ""
		if item.HasConflict {
			conflictBadge = " " + StyleBadgeWarn.Render("CONFLICT")
		}

		typeBadge := lipgloss.NewStyle().Foreground(ColorViolet).Render(fmt.Sprintf("[%s]", item.Type))
		harnessBadge := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("(%s)", item.HarnessID))

		nameStr := item.Name
		if isFocused {
			nameStr = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Underline(true).Render(item.Name)
		}

		line := fmt.Sprintf("%s%s %s %s %s%s\n", cursorStr, checkStr, typeBadge, nameStr, harnessBadge, conflictBadge)
		b.WriteString(line)
	}

	remaining := len(m.Items) - end
	if remaining > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(fmt.Sprintf("  ▼ %d more item(s) below...", remaining)) + "\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("Press ENTER to proceed with repository setup and symlinking.") + "\n")
	return b.String()
}

// RunInitView launches the interactive Bubbletea TUI loop for repository initialization.
func RunInitView(items []InitCapabilityItem) ([]InitCapabilityItem, bool, error) {
	model := NewInitModel(items)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	m, ok := finalModel.(*InitModel)
	if !ok {
		return nil, false, fmt.Errorf("unexpected model type returned from Bubbletea")
	}

	if m.Canceled {
		return nil, false, nil
	}

	return m.Items, m.Confirmed, nil
}
