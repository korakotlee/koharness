package doctor

import (
	"fmt"
	"os"

	"github.com/korakotlee/koharness/pkg/harness"
)

// HarnessHealthStatus represents the diagnostic report for a specific client AI harness.
type HarnessHealthStatus struct {
	// ID is the unique identifier of the harness (antigravity, claude, codex).
	ID harness.HarnessID
	// Name is the human-readable display name.
	Name string
	// Installed indicates if the harness configuration or home root was found.
	Installed bool
	// ConfigDir holds the main config directory path.
	ConfigDir string
	// ConfigDirAccessible is true if ConfigDir exists and has read/write access.
	ConfigDirAccessible bool
	// MissingPaths lists expected subdirectories or config files absent on disk.
	MissingPaths []string
	// FoundPaths lists existing subdirectories or config files on disk.
	FoundPaths []string
	// StatusMessage provides diagnostic details or warning notes.
	StatusMessage string
}

// InspectHarnesses evaluates local configuration directory permissions and health across supported client harnesses.
func InspectHarnesses(homeDir string) ([]HarnessHealthStatus, error) {
	opts := []harness.DetectorOption{}
	if homeDir != "" {
		opts = append(opts, harness.WithHomeDir(homeDir))
	}

	detector, err := harness.NewDetector(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed initializing harness detector: %w", err)
	}

	adapters := detector.GetAdapters()
	statuses := make([]HarnessHealthStatus, 0, len(adapters))

	for _, adapter := range adapters {
		status := adapter.GetStatus()
		paths := adapter.GetConfigPaths()

		health := HarnessHealthStatus{
			ID:           adapter.ID(),
			Name:         adapter.Name(),
			Installed:    status.Installed,
			ConfigDir:    paths.ConfigDir,
			MissingPaths: status.PathsMissing,
			FoundPaths:   status.PathsFound,
		}

		if status.Installed {
			// Check config directory accessibility
			if info, err := os.Stat(paths.ConfigDir); err == nil && info.IsDir() {
				// Test readability
				if _, readErr := os.ReadDir(paths.ConfigDir); readErr == nil {
					health.ConfigDirAccessible = true
					health.StatusMessage = "Config directory accessible with read permissions"
				} else {
					health.StatusMessage = fmt.Sprintf("Config directory unreadable: %v", readErr)
				}
			} else {
				health.StatusMessage = "Config directory missing or inaccessible"
			}
		} else {
			health.StatusMessage = "Harness not detected on workstation"
		}

		statuses = append(statuses, health)
	}

	return statuses, nil
}
