package harness_test

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

func TestGetRepoPath(t *testing.T) {
	fakeHome := "/Users/testuser"

	tests := []struct {
		name       string
		customPath string
		envVal     string
		expected   string
	}{
		{
			name:       "Default fallback path",
			customPath: "",
			envVal:     "",
			expected:   filepath.Join(fakeHome, ".koharness/repo"),
		},
		{
			name:       "Environment variable override",
			customPath: "",
			envVal:     "~/my-custom-repo",
			expected:   filepath.Join(fakeHome, "my-custom-repo"),
		},
		{
			name:       "Custom parameter flag takes highest precedence",
			customPath: "/opt/dotfiles",
			envVal:     "~/ignored-repo",
			expected:   "/opt/dotfiles",
		},
		{
			name:       "Custom parameter flag with tilde expansion",
			customPath: "~/personal/repo",
			envVal:     "",
			expected:   filepath.Join(fakeHome, "personal/repo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := harness.PathOptions{
				HomeDir: fakeHome,
				GetEnv: func(key string) string {
					if key == harness.EnvKoharnessRepo {
						return tt.envVal
					}
					return ""
				},
			}

			resolved, err := harness.GetRepoPath(tt.customPath, opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resolved != tt.expected {
				t.Errorf("got %q, want %q", resolved, tt.expected)
			}
		})
	}
}

func TestGlobalConfigSaveAndLoad(t *testing.T) {
	fakeHome := "/home/testuser"
	memFs := afero.NewMemMapFs()

	opts := harness.PathOptions{
		Fs:      memFs,
		HomeDir: fakeHome,
	}

	// 1. Initially config does not exist, LoadGlobalConfig should return empty config
	cfg, err := harness.LoadGlobalConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error loading empty config: %v", err)
	}
	if cfg.RepoPath != "" {
		t.Errorf("expected empty RepoPath, got %q", cfg.RepoPath)
	}

	// 2. Save config
	cfg.RepoPath = "~/custom/dotfiles"
	cfg.OriginalHarnesses = []string{"antigravity", "claude"}
	if err := harness.SaveGlobalConfig(cfg, opts); err != nil {
		t.Fatalf("failed saving config: %v", err)
	}

	// 3. Load config and verify saved values
	loaded, err := harness.LoadGlobalConfig(opts)
	if err != nil {
		t.Fatalf("failed loading saved config: %v", err)
	}
	if loaded.RepoPath != "~/custom/dotfiles" {
		t.Errorf("expected RepoPath %q, got %q", "~/custom/dotfiles", loaded.RepoPath)
	}
	if len(loaded.OriginalHarnesses) != 2 || loaded.OriginalHarnesses[0] != "antigravity" || loaded.OriginalHarnesses[1] != "claude" {
		t.Errorf("expected OriginalHarnesses [antigravity claude], got %v", loaded.OriginalHarnesses)
	}

	// 4. Test GetRepoPath resolution using config file
	resolved, err := harness.GetRepoPath("", opts)
	if err != nil {
		t.Fatalf("unexpected error resolving path from config: %v", err)
	}
	expected := filepath.Join(fakeHome, "custom/dotfiles")
	if resolved != expected {
		t.Errorf("GetRepoPath with config = %q, want %q", resolved, expected)
	}
}

func TestExpandTilde(t *testing.T) {
	fakeHome := "/home/user"

	tests := []struct {
		input    string
		expected string
	}{
		{"~", fakeHome},
		{"~/sub/dir", filepath.Join(fakeHome, "sub/dir")},
		{"/var/log", "/var/log"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := harness.ExpandTilde(tt.input, fakeHome)
		if got != tt.expected {
			t.Errorf("ExpandTilde(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

