package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/korakotlee/koharness/pkg/memory"
)

// MemoryDiagnostic holds health diagnostic results for the 3-layer agent memory architecture.
type MemoryDiagnostic struct {
	// MemoryDirExists indicates if the memory/ root directory exists.
	MemoryDirExists bool
	// RawDirExists indicates if memory/raw/ exists.
	RawDirExists bool
	// WikiDirExists indicates if memory/wiki/ exists.
	WikiDirExists bool
	// AgentsFileExists indicates if memory/AGENTS.md exists.
	AgentsFileExists bool
	// IndexFileExists indicates if memory/wiki/INDEX.md exists.
	IndexFileExists bool
	// BrokenWikiLinks lists broken Markdown link descriptions found in wiki files.
	BrokenWikiLinks []string
	// InvalidTriggers lists invalid trigger path descriptions found in AGENTS.md.
	InvalidTriggers []string
}

// IsHealthy returns true if all 3-layer memory components are present and no broken links or triggers are found.
func (d *MemoryDiagnostic) IsHealthy() bool {
	return d.MemoryDirExists && d.RawDirExists && d.WikiDirExists && d.AgentsFileExists && d.IndexFileExists &&
		len(d.BrokenWikiLinks) == 0 && len(d.InvalidTriggers) == 0
}

// InspectMemory performs health checks on the repository's 3-layer agent memory layout.
func InspectMemory(repoRoot string) (*MemoryDiagnostic, error) {
	memDir := filepath.Join(repoRoot, "memory")
	diag := &MemoryDiagnostic{}

	if info, err := os.Stat(memDir); err == nil && info.IsDir() {
		diag.MemoryDirExists = true
	} else {
		return diag, nil
	}

	status, err := memory.InspectMemoryLayout(repoRoot)
	if err == nil {
		diag.RawDirExists = status.RawDirExists
		diag.WikiDirExists = status.WikiDirExists
		diag.AgentsFileExists = status.AgentsFileExists
		diag.IndexFileExists = status.IndexFileExists
	}

	// Inspect wiki links
	wikiDir := filepath.Join(memDir, "wiki")
	if diag.WikiDirExists {
		_ = filepath.Walk(wikiDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}

			broken, bErr := memory.ValidateMarkdownLinks(path)
			if bErr == nil {
				for _, b := range broken {
					relPath, _ := filepath.Rel(repoRoot, path)
					diag.BrokenWikiLinks = append(diag.BrokenWikiLinks, fmt.Sprintf("%s (line %d) -> %s", relPath, b.LineNumber, b.Target))
				}
			}
			return nil
		})
	}

	// Inspect steering trigger paths
	agentsFile := filepath.Join(memDir, "AGENTS.md")
	if diag.AgentsFileExists {
		invalid, tErr := memory.ValidateTriggerPaths(repoRoot, agentsFile)
		if tErr == nil {
			for _, inv := range invalid {
				diag.InvalidTriggers = append(diag.InvalidTriggers, fmt.Sprintf("AGENTS.md (line %d) -> %s", inv.LineNumber, inv.TargetPath))
			}
		}
	}

	return diag, nil
}
