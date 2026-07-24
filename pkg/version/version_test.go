package version

import (
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Errorf("expected non-empty version")
	}
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Errorf("expected non-empty version string")
	}
}

func TestCustomVersionOverride(t *testing.T) {
	oldVer := Version
	defer func() { Version = oldVer }()

	Version = "v1.2.3"
	if GetVersion() != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", GetVersion())
	}
}
