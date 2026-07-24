package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Brand Color Palette
	ColorViolet  = lipgloss.Color("#7D56F4")
	ColorMagenta = lipgloss.Color("#9B51E0")
	ColorCyan    = lipgloss.Color("#00F5D4")
	ColorGreen   = lipgloss.Color("#00E676")
	ColorAmber   = lipgloss.Color("#FFB300")
	ColorRed     = lipgloss.Color("#FF3D00")
	ColorMuted   = lipgloss.Color("#767676")

	// Base Text Styles
	StyleBold   = lipgloss.NewStyle().Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleViolet = lipgloss.NewStyle().Foreground(ColorViolet)
	StyleCyan   = lipgloss.NewStyle().Foreground(ColorCyan)

	// Status Badge Pill Styles
	StyleBadgeSuccess = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(ColorGreen).
				Padding(0, 1)

	StyleBadgeWarn = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorAmber).
			Padding(0, 1)

	StyleBadgeError = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorRed).
			Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorCyan).
			Padding(0, 1)
)

// Badge helpers
func BadgeSuccess(msg string) string {
	if msg == "" {
		msg = "SUCCESS"
	}
	return StyleBadgeSuccess.Render(msg)
}

func BadgeWarn(msg string) string {
	if msg == "" {
		msg = "WARN"
	}
	return StyleBadgeWarn.Render(msg)
}

func BadgeError(msg string) string {
	if msg == "" {
		msg = "ERROR"
	}
	return StyleBadgeError.Render(msg)
}

func BadgeInfo(msg string) string {
	if msg == "" {
		msg = "INFO"
	}
	return StyleBadgeInfo.Render(msg)
}
