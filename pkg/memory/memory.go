// Package memory provides parsing, verification, and diagnostic logic for the 3-layer agent memory architecture.
//
// The 3-layer agent memory architecture standardizes team and personal AI documentation using:
// 1. raw/: Immutable original source files (PDFs, PPTs, raw transcripts).
// 2. wiki/: Compiled Markdown knowledge base anchored by wiki/INDEX.md.
// 3. AGENTS.md: Root steering file with core map and context triggering rules.
package memory

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MarkdownLink represents a hyperlinked reference found within a Markdown document.
type MarkdownLink struct {
	// SourceFile is the file containing the link.
	SourceFile string
	// LineNumber is the line on which the link was found (1-indexed).
	LineNumber int
	// Text is the anchor text of the link.
	Text string
	// Target is the target URI or relative filepath of the link.
	Target string
}

// BrokenLink represents a Markdown link whose target file does not exist on disk.
type BrokenLink struct {
	MarkdownLink
	// ResolvedPath is the absolute or relative path evaluated on disk.
	ResolvedPath string
}

// InvalidTrigger represents a trigger path referenced in AGENTS.md that cannot be resolved on disk.
type InvalidTrigger struct {
	// SourceFile is the AGENTS.md steering file path.
	SourceFile string
	// LineNumber is the line where the trigger path was referenced.
	LineNumber int
	// TargetPath is the extracted relative target path.
	TargetPath string
}

// MemoryStatus represents the completeness state of a 3-layer agent memory structure.
type MemoryStatus struct {
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
	// MissingItems lists human-readable names of missing memory components.
	MissingItems []string
}

// IsComplete returns true if all 3-layer memory components exist.
func (s *MemoryStatus) IsComplete() bool {
	return s.MemoryDirExists && s.RawDirExists && s.WikiDirExists && s.AgentsFileExists && s.IndexFileExists
}

// InspectMemoryLayout verifies the existence of the 3-layer memory directory components in a repository.
func InspectMemoryLayout(repoRoot string) (*MemoryStatus, error) {
	memDir := filepath.Join(repoRoot, "memory")
	status := &MemoryStatus{}

	if info, err := os.Stat(memDir); err == nil && info.IsDir() {
		status.MemoryDirExists = true
	} else {
		status.MissingItems = append(status.MissingItems, "memory/")
		return status, nil
	}

	rawDir := filepath.Join(memDir, "raw")
	if info, err := os.Stat(rawDir); err == nil && info.IsDir() {
		status.RawDirExists = true
	} else {
		status.MissingItems = append(status.MissingItems, "memory/raw/")
	}

	wikiDir := filepath.Join(memDir, "wiki")
	if info, err := os.Stat(wikiDir); err == nil && info.IsDir() {
		status.WikiDirExists = true
	} else {
		status.MissingItems = append(status.MissingItems, "memory/wiki/")
	}

	agentsFile := filepath.Join(memDir, "AGENTS.md")
	if info, err := os.Stat(agentsFile); err == nil && !info.IsDir() {
		status.AgentsFileExists = true
	} else {
		status.MissingItems = append(status.MissingItems, "memory/AGENTS.md")
	}

	indexFile := filepath.Join(memDir, "wiki", "INDEX.md")
	if info, err := os.Stat(indexFile); err == nil && !info.IsDir() {
		status.IndexFileExists = true
	} else {
		status.MissingItems = append(status.MissingItems, "memory/wiki/INDEX.md")
	}

	return status, nil
}

// linkRegex matches standard Markdown relative links [text](target).
var linkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// ExtractMarkdownLinks reads a Markdown file and extracts internal file link references.
func ExtractMarkdownLinks(filePath string) ([]MarkdownLink, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Markdown file %s: %w", filePath, err)
	}
	defer file.Close()

	var links []MarkdownLink
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := linkRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			text := strings.TrimSpace(m[1])
			target := strings.TrimSpace(m[2])

			// Ignore external web links, mailto links, or anchor-only links
			if strings.HasPrefix(target, "http://") ||
				strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") ||
				strings.HasPrefix(target, "#") {
				continue
			}

			// Strip section anchor from path if present (e.g. page.md#heading)
			if idx := strings.Index(target, "#"); idx != -1 {
				target = target[:idx]
			}
			if target == "" {
				continue
			}

			links = append(links, MarkdownLink{
				SourceFile: filePath,
				LineNumber: lineNum,
				Text:       text,
				Target:     target,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Markdown file %s: %w", filePath, err)
	}

	return links, nil
}

// ValidateMarkdownLinks checks that all relative Markdown links in a document resolve to existing files.
func ValidateMarkdownLinks(filePath string) ([]BrokenLink, error) {
	links, err := ExtractMarkdownLinks(filePath)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(filePath)
	var broken []BrokenLink

	for _, link := range links {
		resolved := filepath.Join(baseDir, link.Target)
		if _, err := os.Stat(resolved); os.IsNotExist(err) {
			broken = append(broken, BrokenLink{
				MarkdownLink: link,
				ResolvedPath: resolved,
			})
		}
	}

	return broken, nil
}

// codePathRegex matches file references like `wiki/architecture.md`, 'wiki/architecture.md', or "raw/notes.txt".
var codePathRegex = regexp.MustCompile("[\"'`]?((?:wiki|raw|memory)/[a-zA-Z0-9_/.-]+)[\"'`]?")

// ValidateTriggerPaths parses an AGENTS.md steering file and verifies referenced trigger file paths exist.
func ValidateTriggerPaths(repoRoot string, agentsFilePath string) ([]InvalidTrigger, error) {
	file, err := os.Open(agentsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open steering file %s: %w", agentsFilePath, err)
	}
	defer file.Close()

	var invalid []InvalidTrigger
	scanner := bufio.NewScanner(file)
	lineNum := 0

	memDir := filepath.Join(repoRoot, "memory")

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := codePathRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			rawPath := strings.TrimSpace(m[1])

			// Skip wildcard patterns
			if strings.Contains(rawPath, "*") {
				continue
			}

			// Resolve relative path against repoRoot or memory root
			var resolved string
			if strings.HasPrefix(rawPath, "memory/") {
				resolved = filepath.Join(repoRoot, rawPath)
			} else {
				resolved = filepath.Join(memDir, rawPath)
			}

			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				invalid = append(invalid, InvalidTrigger{
					SourceFile: agentsFilePath,
					LineNumber: lineNum,
					TargetPath: rawPath,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading steering file %s: %w", agentsFilePath, err)
	}

	return invalid, nil
}
