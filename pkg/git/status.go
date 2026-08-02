// Package git provides tools for checking repository status, inspecting uncommitted local changes, and executing Git operations safely.
package git

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrGitNotFound indicates that the git executable is not available in the system PATH.
var ErrGitNotFound = errors.New("git executable not found in PATH")

// DirtyFile represents a file in the repository with uncommitted changes or untracked state.
type DirtyFile struct {
	// Path is the relative path of the dirty file within the repository.
	Path string
	// Status describes the porcelain status flags (e.g. "M", "A", "D", "??", "UU").
	Status string
}

// RepoStatus contains summary information regarding the Git state of a repository.
type RepoStatus struct {
	// IsDirty indicates whether the repository has uncommitted modifications, deletions, or untracked files.
	IsDirty bool
	// Branch is the current active Git branch name.
	Branch string
	// DirtyFiles lists all files with uncommitted changes or untracked status.
	DirtyFiles []DirtyFile
}

// InspectStatus executes git status commands to inspect the state of the target repository at repoPath.
// It returns a RepoStatus struct describing the branch and dirty state.
func InspectStatus(ctx context.Context, repoPath string) (*RepoStatus, error) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrGitNotFound
	}

	cleanRepoPath := filepath.Clean(repoPath)

	// Run git status --porcelain=v1 -b to retrieve branch and dirty files
	cmd := exec.CommandContext(ctx, gitBin, "-C", cleanRepoPath, "status", "--porcelain=v1", "-b")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git status failed in %s: %w (output: %s)", cleanRepoPath, err, strings.TrimSpace(string(outputBytes)))
	}

	status := ParsePorcelain(string(outputBytes))
	return status, nil
}

// ParsePorcelain parses the string output of `git status --porcelain=v1 -b` into a RepoStatus struct.
func ParsePorcelain(output string) *RepoStatus {
	status := &RepoStatus{
		DirtyFiles: make([]DirtyFile, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			// Branch header line: "## main...origin/main" or "## HEAD (no branch)"
			branchInfo := strings.TrimPrefix(line, "## ")
			if idx := strings.Index(branchInfo, "..."); idx != -1 {
				status.Branch = branchInfo[:idx]
			} else if idx := strings.Index(branchInfo, " "); idx != -1 {
				status.Branch = branchInfo[:idx]
			} else {
				status.Branch = branchInfo
			}
			continue
		}

		// Porcelain line format: XY PATH or XY PATH -> PATH2
		if len(line) < 3 {
			continue
		}

		statusCode := strings.TrimSpace(line[:2])
		filePath := strings.TrimSpace(line[3:])

		// Handle renamed files (e.g. "R  file1 -> file2")
		if idx := strings.Index(filePath, " -> "); idx != -1 {
			filePath = filePath[idx+4:]
		}

		// Remove surrounding quotes if git output quoted path
		filePath = strings.Trim(filePath, "\"")

		status.DirtyFiles = append(status.DirtyFiles, DirtyFile{
			Path:   filePath,
			Status: statusCode,
		})
	}

	status.IsDirty = len(status.DirtyFiles) > 0
	return status
}

// PullRebase executes `git pull --rebase` in the specified repository directory.
func PullRebase(ctx context.Context, repoPath string) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return ErrGitNotFound
	}

	cleanRepoPath := filepath.Clean(repoPath)
	cmd := exec.CommandContext(ctx, gitBin, "-C", cleanRepoPath, "pull", "--rebase")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(outputBytes))
		return fmt.Errorf("git pull --rebase failed in %s: %w\nGit output:\n%s", cleanRepoPath, err, outputStr)
	}

	return nil
}
