package harness

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	// MemoryStartMarker demarcates the beginning of the koharness-managed memory navigation block.
	MemoryStartMarker = "<!-- KOHARNESS:MEMORY_START -->"
	// MemoryEndMarker demarcates the end of the koharness-managed memory navigation block.
	MemoryEndMarker = "<!-- KOHARNESS:MEMORY_END -->"
)

// GenerateMemoryBlock formats the managed memory navigation instruction block.
func GenerateMemoryBlock(repoMemoryAgentsPath string) string {
	return fmt.Sprintf("%s\n# Memory Navigation\n\nRefer to repository agent memory steering rules at `%s`:\n- Core Map: Read `wiki/INDEX.md` first to locate specific topics.\n- Steering Reference: `%s`\n%s",
		MemoryStartMarker,
		repoMemoryAgentsPath,
		repoMemoryAgentsPath,
		MemoryEndMarker,
	)
}

// InjectOrUpdateMemorySection updates or appends the managed memory block in a client harness AGENTS.md file.
func InjectOrUpdateMemorySection(fs afero.Fs, targetFilePath string, repoMemoryAgentsPath string) error {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	var existingContent string
	if exists, _ := afero.Exists(fs, targetFilePath); exists {
		data, err := afero.ReadFile(fs, targetFilePath)
		if err != nil {
			return fmt.Errorf("failed to read target harness AGENTS.md at %s: %w", targetFilePath, err)
		}
		existingContent = string(data)
	}

	newBlock := GenerateMemoryBlock(repoMemoryAgentsPath)

	startIdx := strings.Index(existingContent, MemoryStartMarker)
	endIdx := strings.Index(existingContent, MemoryEndMarker)

	var updatedContent string
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		// Replace existing managed block
		prefix := existingContent[:startIdx]
		suffix := existingContent[endIdx+len(MemoryEndMarker):]
		updatedContent = strings.TrimRight(prefix, "\r\n") + "\n" + newBlock + "\n" + strings.TrimLeft(suffix, "\r\n")
	} else {
		// Append managed block to file
		if strings.TrimSpace(existingContent) == "" {
			updatedContent = newBlock + "\n"
		} else {
			updatedContent = strings.TrimRight(existingContent, "\r\n") + "\n\n" + newBlock + "\n"
		}
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(targetFilePath)
	if err := fs.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed creating harness directory %s: %w", targetDir, err)
	}

	return afero.WriteFile(fs, targetFilePath, []byte(updatedContent), 0644)
}

// SyncMemoryNavigationRules updates the managed memory section across all detected client harness AGENTS.md files.
func SyncMemoryNavigationRules(fs afero.Fs, homeDir string, repoPath string) error {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	repoMemoryAgentsPath := filepath.Join(repoPath, "memory", "AGENTS.md")

	// Target client harness directories
	harnessDirs := []string{
		filepath.Join(homeDir, ".gemini"),
		filepath.Join(homeDir, ".claude"),
		filepath.Join(homeDir, ".codex"),
	}

	for _, dir := range harnessDirs {
		// Only inject if harness directory exists or is standard
		targetFile := filepath.Join(dir, "AGENTS.md")
		if err := InjectOrUpdateMemorySection(fs, targetFile, repoMemoryAgentsPath); err != nil {
			return fmt.Errorf("failed injecting memory navigation into %s: %w", targetFile, err)
		}
	}

	return nil
}
