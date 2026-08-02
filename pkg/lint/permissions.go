package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ValidateScriptPermissions verifies that all files inside skills/*/scripts/ have the executable bit (+x) set.
func ValidateScriptPermissions(repoRoot string) ([]LintIssue, error) {
	var issues []LintIssue

	skillsDir := filepath.Join(repoRoot, "skills")
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return nil, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed reading skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		scriptsDir := filepath.Join(skillsDir, entry.Name(), "scripts")
		sInfo, sErr := os.Stat(scriptsDir)
		if os.IsNotExist(sErr) || !sInfo.IsDir() {
			continue
		}

		err = filepath.Walk(scriptsDir, func(path string, fileInfo os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if fileInfo.IsDir() {
				return nil
			}

			// Ignore hidden files like .DS_Store or .gitkeep
			if filepath.Base(path)[0] == '.' {
				return nil
			}

			relPath, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				relPath = path
			}

			// On Windows, permission modes do not have POSIX executable bits (+x).
			if runtime.GOOS != "windows" {
				if fileInfo.Mode().Perm()&0111 == 0 {
					issues = append(issues, LintIssue{
						Path:     relPath,
						Category: "Executable Bit",
						Message:  "script file lacks executable permission (chmod +x required)",
					})
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking scripts directory %s: %w", scriptsDir, err)
		}
	}

	return issues, nil
}
