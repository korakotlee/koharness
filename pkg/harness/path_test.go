package harness_test

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
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
