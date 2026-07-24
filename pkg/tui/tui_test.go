package tui

import (
	"strings"
	"testing"
)

func TestBadges(t *testing.T) {
	if !strings.Contains(BadgeSuccess(""), "SUCCESS") {
		t.Errorf("expected SUCCESS in badge")
	}
	if !strings.Contains(BadgeWarn(""), "WARN") {
		t.Errorf("expected WARN in badge")
	}
	if !strings.Contains(BadgeError(""), "ERROR") {
		t.Errorf("expected ERROR in badge")
	}
	if !strings.Contains(BadgeInfo(""), "INFO") {
		t.Errorf("expected INFO in badge")
	}
}

func TestRenderBanner(t *testing.T) {
	banner := RenderBanner()
	if banner == "" {
		t.Errorf("expected non-empty banner output")
	}
	if !strings.Contains(banner, "KoHarness CLI") {
		t.Errorf("expected KoHarness CLI in banner output")
	}
}
