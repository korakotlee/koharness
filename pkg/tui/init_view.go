// Package tui provides terminal user interface components for KoHarness,
// including styled banners, badges, and interactive Bubbletea views.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InitCapabilityItem represents a capability discovered in the repository to link into a client harness.
type InitCapabilityItem struct {
	// Type specifies the capability category (skill, prompt, mcp).
	Type string
	// Name specifies the identifier or file name of the capability.
	Name string
	// HarnessID identifies the target client harness (gemini, claude, codex).
	HarnessID string
	// RepoPath is the absolute path to the asset within the cloned repository.
	RepoPath string
	// TargetPath is the target destination path in the local client harness.
	TargetPath string
	// HasConflict indicates whether a local file or directory already exists at TargetPath.
	HasConflict bool
	// Selected indicates whether this item is toggled for symlinking during initialization.
	Selected bool
}

// InitModel represents the interactive Bubbletea TUI model for repository setup and capability linking.
type InitModel struct {
	Items        []InitCapabilityItem
	Cursor       int
	Offset       int
	MaxVisible   int
	Confirmed    bool
	Canceled     bool
	WindowHeight int
}

// NewInitModel constructs a new InitModel pointer with the provided repository capability items.
func NewInitModel(items []InitCapabilityItem) *InitModel {
	return &InitModel{
		Items:      items,
		Cursor:     0,
		Offset:     0,
		MaxVisible: 10,
		Confirmed:  false,
		Canceled:   false,
	}
}

// Init initializes the Bubbletea model lifecycle.
func (m *InitModel) Init() tea.Cmd {
	return nil
}

func (m *InitModel) setCursor(index int) {
	if len(m.Items) == 0 {
		m.Cursor = 0
		m.Offset = 0
		return
	}

	if index < 0 {
		index = 0
	}
	if index >= len(m.Items) {
		index = len(m.Items) - 1
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

// Update processes terminal input events and updates cursor and selection states.
func (m *InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowHeight = msg.Height
		maxVis := msg.Height - 12
		if maxVis < 5 {
			maxVis = 5
		}
		m.MaxVisible = maxVis
		m.setCursor(m.Cursor)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.Canceled = true
			return m, tea.Quit
		case tea.KeyUp:
			m.setCursor(m.Cursor - 1)
			return m, nil
		case tea.KeyDown:
			m.setCursor(m.Cursor + 1)
			return m, nil
		case tea.KeyPgUp:
			m.setCursor(m.Cursor - m.MaxVisible)
			return m, nil
		case tea.KeyPgDown:
			m.setCursor(m.Cursor + m.MaxVisible)
			return m, nil
		case tea.KeyHome:
			m.setCursor(0)
			return m, nil
		case tea.KeyEnd:
			m.setCursor(len(m.Items) - 1)
			return m, nil
		case tea.KeySpace:
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
			}
			return m, nil
		case tea.KeyEnter:
			m.Confirmed = true
			return m, tea.Quit
		}

		key := strings.ToLower(msg.String())
		switch key {
		case "q", "esc", "ctrl+c":
			m.Canceled = true
			return m, tea.Quit
		case "up", "k":
			m.setCursor(m.Cursor - 1)
		case "down", "j":
			m.setCursor(m.Cursor + 1)
		case "pgup", "b":
			m.setCursor(m.Cursor - m.MaxVisible)
		case "pgdown", "f":
			m.setCursor(m.Cursor + m.MaxVisible)
		case "g", "home":
			m.setCursor(0)
		case "G", "end":
			m.setCursor(len(m.Items) - 1)
		case "space", " ":
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
			}
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
		case "enter":
			m.Confirmed = true
			return m, tea.Quit
		}
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
