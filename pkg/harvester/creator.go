package harvester

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/symlink"
	"github.com/spf13/afero"
)

// CreatorOptions configures execution settings for repository bootstrapping and capability harvesting.
type CreatorOptions struct {
	Fs         afero.Fs
	HomeDir    string
	RepoPath   string
	BackupRoot string
	InitGit    bool
}

// Creator orchestrates dotfile repository bootstrapping, asset copying, backup creation, and symlinking.
type Creator struct {
	fs        afero.Fs
	homeDir   string
	repoPath  string
	initGit   bool
	backupMgr *symlink.BackupManager
	linker    *symlink.LinkerEngine
}

// NewCreator initializes a Creator instance with the provided options.
func NewCreator(opts CreatorOptions) *Creator {
	fs := opts.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	if opts.RepoPath == "" {
		opts.RepoPath = filepath.Join(opts.HomeDir, ".koharness", "repo")
	}

	bm := symlink.NewBackupManager(fs, opts.HomeDir, opts.BackupRoot)
	linker := symlink.NewLinkerEngine(symlink.LinkerConfig{
		Fs:            fs,
		HomeDir:       opts.HomeDir,
		BackupManager: bm,
		DryRun:        false,
	})

	return &Creator{
		fs:        fs,
		homeDir:   opts.HomeDir,
		repoPath:  opts.RepoPath,
		initGit:   opts.InitGit,
		backupMgr: bm,
		linker:    linker,
	}
}

// IsRepoExisting checks if the target repository directory already exists and contains files.
func (c *Creator) IsRepoExisting() (bool, error) {
	exists, err := afero.Exists(c.fs, c.repoPath)
	if err != nil || !exists {
		return false, nil
	}
	entries, err := afero.ReadDir(c.fs, c.repoPath)
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
}

// ScaffoldRepo initializes the target repository structure, .gitignore, and .koharness.yaml configuration.
func (c *Creator) ScaffoldRepo() error {
	dirs := []string{
		c.repoPath,
		filepath.Join(c.repoPath, "skills"),
		filepath.Join(c.repoPath, "prompts"),
		filepath.Join(c.repoPath, "mcp"),
		filepath.Join(c.repoPath, "memory"),
		filepath.Join(c.repoPath, "memory", "raw"),
		filepath.Join(c.repoPath, "memory", "wiki"),
	}

	for _, d := range dirs {
		if err := c.fs.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Scaffold 3-layer agent memory template assets
	memoryAgentsPath := filepath.Join(c.repoPath, "memory", "AGENTS.md")
	memoryAgentsContent := `# Agent Memory Navigation

## Core Map
- Read ` + "`wiki/INDEX.md`" + ` first to locate specific topics before searching files.

## Triggering Rules
- **Tech Architecture**: Check ` + "`wiki/architecture.md`" + ` before answering infrastructure or tech stack questions.
- **Raw Files**: Only inspect original assets in ` + "`raw/`" + ` if ` + "`wiki/`" + ` lacks detailed figures or visual layout details.
`
	if err := afero.WriteFile(c.fs, memoryAgentsPath, []byte(memoryAgentsContent), 0644); err != nil {
		return fmt.Errorf("failed to write memory/AGENTS.md: %w", err)
	}

	memoryIndexPath := filepath.Join(c.repoPath, "memory", "wiki", "INDEX.md")
	memoryIndexContent := `# Memory Index

- ` + "`architecture.md`" + `: System configurations, API contracts, deployment steps.
`
	if err := afero.WriteFile(c.fs, memoryIndexPath, []byte(memoryIndexContent), 0644); err != nil {
		return fmt.Errorf("failed to write memory/wiki/INDEX.md: %w", err)
	}

	memoryArchPath := filepath.Join(c.repoPath, "memory", "wiki", "architecture.md")
	memoryArchContent := `# System Architecture

High-level component details and environment configurations.
`
	if err := afero.WriteFile(c.fs, memoryArchPath, []byte(memoryArchContent), 0644); err != nil {
		return fmt.Errorf("failed to write memory/wiki/architecture.md: %w", err)
	}

	gitkeepPath := filepath.Join(c.repoPath, "memory", "raw", ".gitkeep")
	if err := afero.WriteFile(c.fs, gitkeepPath, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write memory/raw/.gitkeep: %w", err)
	}

	gitignorePath := filepath.Join(c.repoPath, ".gitignore")
	gitignoreContent := `.DS_Store
mcp.local.json
*.log
credentials.json
`
	if err := afero.WriteFile(c.fs, gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	manifestPath := filepath.Join(c.repoPath, ".koharness.yaml")
	manifestContent := `version: "1.0"
managed_by: koharness
`
	if err := afero.WriteFile(c.fs, manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("failed to write .koharness.yaml: %w", err)
	}

	if c.initGit {
		_ = c.runGitInit(c.repoPath)
	}

	return nil
}

// HarvestCapabilities copies selected capabilities into the repository, archives backups, and establishes symlinks.
func (c *Creator) HarvestCapabilities(items []DiscoveredCapability) error {
	if err := c.ScaffoldRepo(); err != nil {
		return err
	}

	for _, item := range items {
		if !item.Selected {
			continue
		}

		var targetSubDir string
		switch item.Type {
		case harness.CapabilitySkill:
			targetSubDir = "skills"
		case harness.CapabilityWorkflow, harness.CapabilityPrompt:
			targetSubDir = "prompts"
		case harness.CapabilityMCP:
			targetSubDir = "mcp"
		default:
			targetSubDir = "skills"
		}

		destRepoPath := filepath.Join(c.repoPath, targetSubDir, item.Name)

		if item.Type == harness.CapabilityMCP && item.IsSecret {
			localMCPPath := filepath.Join(c.repoPath, targetSubDir, "mcp.local.json")
			if err := c.appendMCPOverride(item, localMCPPath); err != nil {
				return fmt.Errorf("failed writing local MCP override for %s: %w", item.Name, err)
			}
			continue
		}

		// 1. Copy source asset into repo directory
		if err := c.copyAsset(item.SourcePath, destRepoPath); err != nil {
			return fmt.Errorf("failed copying harvested asset %s to repo: %w", item.Name, err)
		}

		// 2. Create atomic symlink from repo asset back to original harness location (with automatic backup)
		if _, err := c.linker.CreateSymlink(destRepoPath, item.SourcePath); err != nil {
			return fmt.Errorf("failed establishing symlink for %s: %w", item.Name, err)
		}
	}

	return nil
}

func (c *Creator) appendMCPOverride(item DiscoveredCapability, targetFile string) error {
	existingData := make(map[string]interface{})
	if exists, _ := afero.Exists(c.fs, targetFile); exists {
		data, err := afero.ReadFile(c.fs, targetFile)
		if err == nil {
			_ = json.Unmarshal(data, &existingData)
		}
	}

	servers, ok := existingData["mcpServers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		existingData["mcpServers"] = servers
	}

	servers[item.Name] = map[string]interface{}{
		"source": item.SourcePath,
	}

	updatedBytes, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return err
	}

	return afero.WriteFile(c.fs, targetFile, updatedBytes, 0644)
}

func (c *Creator) copyAsset(src, dst string) error {
	info, err := c.fs.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return c.copyDir(src, dst)
	}
	return c.copyFile(src, dst)
}

func (c *Creator) copyFile(src, dst string) error {
	if err := c.fs.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := c.fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := c.fs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (c *Creator) copyDir(src, dst string) error {
	if err := c.fs.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := afero.ReadDir(c.fs, src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcChild := filepath.Join(src, entry.Name())
		dstChild := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := c.copyDir(srcChild, dstChild); err != nil {
				return err
			}
		} else {
			if err := c.copyFile(srcChild, dstChild); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Creator) runGitInit(dir string) error {
	cmd := exec.Command("git", "init", dir)
	return cmd.Run()
}
