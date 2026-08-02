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

// HarvestModel represents the interactive Bubbletea TUI model for capability selection.
type HarvestModel struct {
	Items        []harvester.DiscoveredCapability
	Cursor       int
	Offset       int
	MaxVisible   int
	Confirmed    bool
	Canceled     bool
	WindowHeight int
}

// NewHarvestModel constructs a new HarvestModel pointer with the provided discovered capabilities.
func NewHarvestModel(items []harvester.DiscoveredCapability) *HarvestModel {
	return &HarvestModel{
		Items:      items,
		Cursor:     0,
		Offset:     0,
		MaxVisible: 10,
		Confirmed:  false,
		Canceled:   false,
	}
}

// Init initializes the Bubbletea model lifecycle.
func (m *HarvestModel) Init() tea.Cmd {
	return nil
}

func (m *HarvestModel) setCursor(index int) {
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

	// Adjust scroll window offset
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
func (m *HarvestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Explicit KeyType matching
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
				m.Items[m.Cursor].ToggleImportSkip()
			}
			return m, nil
		case tea.KeyEnter:
			m.Confirmed = true
			return m, tea.Quit
		}

		// String representation matching fallback (without strings.TrimSpace)
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
				m.Items[m.Cursor].ToggleImportSkip()
			}
		case "i":
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].ToggleIgnore()
			}
		case "s":
			if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
				m.Items[m.Cursor].IsSecret = !m.Items[m.Cursor].IsSecret
			}
		case "a":
			// Toggle all items between Import and Skip
			allSelected := true
			for _, item := range m.Items {
				if item.GetState() != harvester.StateImport {
					allSelected = false
					break
				}
			}
			for idx := range m.Items {
				if allSelected {
					m.Items[idx].SetState(harvester.StateSkip)
				} else {
					m.Items[idx].SetState(harvester.StateImport)
				}
			}
		case "enter":
			m.Confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the interactive terminal user interface layout.
func (m *HarvestModel) View() string {
	var b strings.Builder

	b.WriteString(RenderBanner() + "\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorViolet).Render("HARVEST DISCOVERED CAPABILITIES") + "\n")

	importCount, skipCount, ignoreCount := 0, 0, 0
	for _, item := range m.Items {
		switch item.GetState() {
		case harvester.StateImport:
			importCount++
		case harvester.StateSkip:
			skipCount++
		case harvester.StateIgnore:
			ignoreCount++
		}
	}

	b.WriteString(StyleMuted.Render(fmt.Sprintf("Selection: %d Import | %d Skip | %d Ignore (Total: %d)", importCount, skipCount, ignoreCount, len(m.Items))) + "\n")
	b.WriteString(StyleMuted.Render("Controls: [↑/↓ or j/k] Navigate | [space] Import/Skip | [i] Ignore | [a] Toggle All | [s] Secret | [enter] Confirm | [q] Cancel") + "\n\n")

	if len(m.Items) == 0 {
		b.WriteString(StyleMuted.Render("No unmanaged capabilities discovered.") + "\n\n")
		return b.String()
	}

	if m.MaxVisible <= 0 {
		m.MaxVisible = 10
	}

	// Scroll indicator top
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
		ignoreBadge := ""
		switch item.GetState() {
		case harvester.StateImport:
			checkStr = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen).Render("[x]")
		case harvester.StateSkip:
			checkStr = lipgloss.NewStyle().Foreground(ColorMuted).Render("[ ]")
		case harvester.StateIgnore:
			checkStr = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render("[i]")
			ignoreBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(" (Ignore - saved to .koharness.local.yaml)")
		}

		secretStr := ""
		if item.IsSecret {
			secretStr = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(" (Local Secret Override)")
		}

		typeBadge := lipgloss.NewStyle().Foreground(ColorViolet).Render(fmt.Sprintf("[%s]", item.Type))
		harnessBadge := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("(%s)", item.HarnessID))

		nameStr := item.Name
		if item.GetState() == harvester.StateIgnore {
			nameStr = lipgloss.NewStyle().Strikethrough(true).Foreground(ColorMuted).Render(item.Name)
		}
		if isFocused {
			nameStr = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Underline(true).Render(item.Name)
		}

		line := fmt.Sprintf("%s%s %s %s %s%s%s\n", cursorStr, checkStr, typeBadge, nameStr, harnessBadge, secretStr, ignoreBadge)
		b.WriteString(line)
	}

	// Scroll indicator bottom
	remaining := len(m.Items) - end
	if remaining > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(fmt.Sprintf("  ▼ %d more item(s) below...", remaining)) + "\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("Press ENTER to proceed with selected items.") + "\n")
	return b.String()
}

// RunHarvestView launches the interactive Bubbletea TUI loop in the terminal.
func RunHarvestView(items []harvester.DiscoveredCapability) ([]harvester.DiscoveredCapability, bool, error) {
	model := NewHarvestModel(items)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	m, ok := finalModel.(*HarvestModel)
	if !ok {
		return nil, false, fmt.Errorf("unexpected model type returned from Bubbletea")
	}

	if m.Canceled {
		return nil, false, nil
	}

	return m.Items, m.Confirmed, nil
}
