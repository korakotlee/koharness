package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Brand Adaptive Color Palette (High Contrast across Dark & Light Terminal Themes)
	ColorViolet  = lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#7D56F4"}
	ColorMagenta = lipgloss.AdaptiveColor{Light: "#7B1FA2", Dark: "#9B51E0"}
	ColorCyan    = lipgloss.AdaptiveColor{Light: "#00796B", Dark: "#00F5D4"}
	ColorGreen   = lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#00E676"}
	ColorAmber   = lipgloss.AdaptiveColor{Light: "#D84315", Dark: "#FFB300"}
	ColorRed     = lipgloss.AdaptiveColor{Light: "#C62828", Dark: "#FF3D00"}
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#616161", Dark: "#767676"}

	// Base Text Styles
	StyleBold   = lipgloss.NewStyle().Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleViolet = lipgloss.NewStyle().Foreground(ColorViolet)
	StyleCyan   = lipgloss.NewStyle().Foreground(ColorCyan)

	// Status Badge Pill Styles
	StyleBadgeSuccess = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorGreen).
				Padding(0, 1)

	StyleBadgeWarn = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorAmber).
			Padding(0, 1)

	StyleBadgeError = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorRed).
			Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
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
