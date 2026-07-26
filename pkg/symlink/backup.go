package symlink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
)

// BackupManager manages creation of timestamped configuration backups prior to filesystem modifications.
type BackupManager struct {
	fs         afero.Fs
	homeDir    string
	backupRoot string

	mu         sync.Mutex
	sessionDir string
	timestamp  string
}

// BackupRecord tracks metadata for a single backed-up filesystem asset.
type BackupRecord struct {
	OriginalPath string
	BackupPath   string
	IsDirectory  bool
	IsSymlink    bool
	SymlinkTarget string
}

// NewBackupManager creates a new BackupManager instance bound to the provided afero.Fs filesystem,
// user home directory, and custom backup root directory override (if empty, defaults to ~/.koharness/backups).
func NewBackupManager(fs afero.Fs, homeDir string, backupRoot string) *BackupManager {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	if backupRoot == "" {
		backupRoot = filepath.Join(homeDir, ".koharness", "backups")
	}
	return &BackupManager{
		fs:         fs,
		homeDir:    homeDir,
		backupRoot: backupRoot,
	}
}

// EnsureSession initializes and returns the active timestamped backup session directory (~/.koharness/backups/YYYYMMDD-HHMMSS).
func (bm *BackupManager) EnsureSession() (string, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.sessionDir != "" {
		return bm.sessionDir, nil
	}

	bm.timestamp = time.Now().Format("20060102-150405")
	bm.sessionDir = filepath.Join(bm.backupRoot, bm.timestamp)

	if err := bm.fs.MkdirAll(bm.sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup session directory %s: %w", bm.sessionDir, err)
	}

	return bm.sessionDir, nil
}

// SessionDir returns the current active backup session directory path.
func (bm *BackupManager) SessionDir() string {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.sessionDir
}

// Backup archives the specified target path into the current session backup directory while preserving path hierarchy.
func (bm *BackupManager) Backup(targetPath string) (*BackupRecord, error) {
	exists, err := afero.Exists(bm.fs, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check existence of target path %s: %w", targetPath, err)
	}
	if !exists {
		// Check if it's a broken symlink via Lstat
		lstater, ok := bm.fs.(afero.Lstater)
		if ok {
			fi, _, lerr := lstater.LstatIfPossible(targetPath)
			if lerr == nil && (fi.Mode()&os.ModeSymlink != 0) {
				exists = true
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("target path %s does not exist for backup", targetPath)
	}

	sessionDir, err := bm.EnsureSession()
	if err != nil {
		return nil, err
	}

	relPath, err := bm.computeRelativePath(targetPath)
	if err != nil {
		return nil, err
	}

	backupDest := filepath.Join(sessionDir, relPath)

	if err := bm.fs.MkdirAll(filepath.Dir(backupDest), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent backup directory for %s: %w", backupDest, err)
	}

	record := &BackupRecord{
		OriginalPath: targetPath,
		BackupPath:   backupDest,
	}

	lstater, ok := bm.fs.(afero.Lstater)
	if ok {
		fi, _, lerr := lstater.LstatIfPossible(targetPath)
		if lerr == nil && (fi.Mode()&os.ModeSymlink != 0) {
			record.IsSymlink = true
			if reader, ok := bm.fs.(afero.LinkReader); ok {
				target, err := reader.ReadlinkIfPossible(targetPath)
				if err == nil {
					record.SymlinkTarget = target
				}
			}
			if err := bm.copySymlink(targetPath, backupDest, record.SymlinkTarget); err != nil {
				return nil, err
			}
			return record, nil
		}
	}

	info, err := bm.fs.Stat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat target path %s: %w", targetPath, err)
	}

	if info.IsDir() {
		record.IsDirectory = true
		if err := bm.copyDir(targetPath, backupDest); err != nil {
			return nil, err
		}
	} else {
		if err := bm.copyFile(targetPath, backupDest); err != nil {
			return nil, err
		}
	}

	return record, nil
}

func (bm *BackupManager) computeRelativePath(targetPath string) (string, error) {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}

	if bm.homeDir != "" {
		absHome, err := filepath.Abs(bm.homeDir)
		if err == nil && strings.HasPrefix(absTarget, absHome) {
			rel, err := filepath.Rel(absHome, absTarget)
			if err == nil {
				return rel, nil
			}
		}
	}

	// Fallback to stripping root separator
	cleaned := filepath.Clean(absTarget)
	return strings.TrimPrefix(cleaned, string(filepath.Separator)), nil
}

func (bm *BackupManager) copyFile(src, dst string) error {
	srcFile, err := bm.fs.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}

	dstFile, err := bm.fs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination backup file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed copying content from %s to %s: %w", src, dst, err)
	}

	return nil
}

func (bm *BackupManager) copyDir(src, dst string) error {
	if err := bm.fs.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination dir %s: %w", dst, err)
	}

	entries, err := afero.ReadDir(bm.fs, src)
	if err != nil {
		return fmt.Errorf("failed to read source dir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcChild := filepath.Join(src, entry.Name())
		dstChild := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := bm.copyDir(srcChild, dstChild); err != nil {
				return err
			}
		} else {
			if err := bm.copyFile(srcChild, dstChild); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bm *BackupManager) copySymlink(src, dst, target string) error {
	if linker, ok := bm.fs.(afero.Symlinker); ok {
		return linker.SymlinkIfPossible(target, dst)
	}
	// Fallback to plain text copy if filesystem doesn't support symlinks
	return afero.WriteFile(bm.fs, dst, []byte(target), 0644)
}
