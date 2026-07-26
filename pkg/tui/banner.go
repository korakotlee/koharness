package tui

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/version"
)

var (
	cachedEmail string
	emailOnce   sync.Once
)

// GetUserEmail retrieves user.email from local git config with fallback, cached after first call.
func GetUserEmail() string {
	emailOnce.Do(func() {
		cmd := exec.Command("git", "config", "user.email")
		out, err := cmd.Output()
		if err == nil {
			email := strings.TrimSpace(string(out))
			if email != "" {
				cachedEmail = email
				return
			}
		}
		cachedEmail = "kleemakdej@gmail.com"
	})
	return cachedEmail
}

// RenderBanner outputs the 5-line block-art logo and metadata panel.
func RenderBanner() string {
	// 5-line block-art KH logo
	lines := []string{
		"█  █ █  █",
		"███  ████",
		"█  █ █  █",
		"█  █ █  █",
	}

	// Gradient colors per line
	lineColors := []lipgloss.Color{
		lipgloss.Color("#06BC38"),
		lipgloss.Color("#FFB901"),
		lipgloss.Color("#AD3B9B"),
		lipgloss.Color("#009EDF"),
	}

	styledLines := make([]string, len(lines))
	for i, line := range lines {
		styledLines[i] = lipgloss.NewStyle().Foreground(lineColors[i]).Bold(true).Render(line)
	}

	logoView := lipgloss.JoinVertical(lipgloss.Left, styledLines...)

	ver := version.GetVersion()
	email := GetUserEmail()

	metaLines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render("KoHarness CLI ") + lipgloss.NewStyle().Foreground(ColorMuted).Render(ver),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0")).Render(email),
		lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render("Central AI Harness Manager"),
	}

	metaView := lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinVertical(lipgloss.Left, metaLines...))

	banner := lipgloss.JoinHorizontal(lipgloss.Center, logoView, metaView)
	return lipgloss.NewStyle().MarginTop(1).MarginBottom(1).Render(banner)
}
