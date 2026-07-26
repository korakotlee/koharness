package symlink

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// OpStatus defines the lifecycle state of a filesystem operation log.
type OpStatus string

const (
	// OpStatusPending indicates an operation is scheduled but not yet executed.
	OpStatusPending OpStatus = "pending"
	// OpStatusExecuted indicates an operation successfully modified the target path.
	OpStatusExecuted OpStatus = "executed"
	// OpStatusCommitted indicates an operation is finalized and verified.
	OpStatusCommitted OpStatus = "committed"
	// OpStatusRolledBack indicates an operation was safely restored from backup.
	OpStatusRolledBack OpStatus = "rolled_back"
)

// OperationLog records state changes performed during a symlink creation transaction.
type OperationLog struct {
	SourcePath   string
	TargetPath   string
	TempPath     string
	BackupRecord *BackupRecord
	Status       OpStatus
}

// Transaction tracks a sequence of filesystem operations for safe rollback upon failure.
type Transaction struct {
	fs         afero.Fs
	ID         string
	Operations []*OperationLog
	DryRun     bool
}

// NewTransaction initializes a new transactional session.
func NewTransaction(fs afero.Fs, dryRun bool) *Transaction {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return &Transaction{
		fs:     fs,
		ID:     generateRandomID(),
		DryRun: dryRun,
	}
}

// Rollback restores backed-up files and removes temporary or newly created symlinks in reverse order.
func (tx *Transaction) Rollback() error {
	if tx.DryRun {
		return nil
	}

	var rollbackErrs []string

	for i := len(tx.Operations) - 1; i >= 0; i-- {
		op := tx.Operations[i]
		if op.Status == OpStatusRolledBack {
			continue
		}

		// Remove newly created target symlink if created
		if op.TargetPath != "" {
			_ = tx.fs.Remove(op.TargetPath)
		}

		// Remove remaining temp symlink if present
		if op.TempPath != "" {
			_ = tx.fs.Remove(op.TempPath)
		}

		// Restore original backed-up asset
		if op.BackupRecord != nil {
			if err := restoreFromBackup(tx.fs, op.BackupRecord); err != nil {
				rollbackErrs = append(rollbackErrs, fmt.Sprintf("failed to restore %s from %s: %v", op.BackupRecord.OriginalPath, op.BackupRecord.BackupPath, err))
			}
		}
		op.Status = OpStatusRolledBack
	}

	if len(rollbackErrs) > 0 {
		return fmt.Errorf("transaction rollback encountered errors: %s", rollbackErrs)
	}
	return nil
}

// Commit finalizes the transaction, removing temporary artifacts without altering target symlinks.
func (tx *Transaction) Commit() {
	for _, op := range tx.Operations {
		if op.TempPath != "" {
			_ = tx.fs.Remove(op.TempPath)
		}
		op.Status = OpStatusCommitted
	}
}

// LinkerConfig configures execution options for LinkerEngine.
type LinkerConfig struct {
	Fs            afero.Fs
	HomeDir       string
	BackupManager *BackupManager
	DryRun        bool
}

// LinkerEngine executes atomic symlink creation, broken link detection, and dry-run previews.
type LinkerEngine struct {
	fs            afero.Fs
	homeDir       string
	backupManager *BackupManager
	dryRun        bool
}

// NewLinkerEngine constructs a LinkerEngine using the provided config parameters.
func NewLinkerEngine(cfg LinkerConfig) *LinkerEngine {
	fs := cfg.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}
	bm := cfg.BackupManager
	if bm == nil {
		bm = NewBackupManager(fs, cfg.HomeDir, "")
	}
	return &LinkerEngine{
		fs:            fs,
		homeDir:       cfg.HomeDir,
		backupManager: bm,
		dryRun:        cfg.DryRun,
	}
}

// CreateSymlink performs atomic symlink creation from sourcePath to targetPath.
// If targetPath exists, it is backed up before being atomically replaced.
func (le *LinkerEngine) CreateSymlink(sourcePath, targetPath string) (*Transaction, error) {
	tx := NewTransaction(le.fs, le.dryRun)

	op := &OperationLog{
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Status:     OpStatusPending,
	}
	tx.Operations = append(tx.Operations, op)

	if le.dryRun {
		op.Status = OpStatusExecuted
		return tx, nil
	}

	// 1. Verify source exists
	srcExists, err := afero.Exists(le.fs, sourcePath)
	if err != nil {
		return tx, fmt.Errorf("failed to verify source path %s: %w", sourcePath, err)
	}
	if !srcExists {
		return tx, fmt.Errorf("symlink source path does not exist: %s", sourcePath)
	}

	// 2. Check if target path exists (or is a broken symlink)
	targetExists := false
	lstater, isLstater := le.fs.(afero.Lstater)
	if isLstater {
		fi, _, lerr := lstater.LstatIfPossible(targetPath)
		if lerr == nil {
			targetExists = true
			_ = fi
		}
	}
	if !targetExists {
		targetExists, _ = afero.Exists(le.fs, targetPath)
	}

	// 3. Backup existing target if present
	if targetExists {
		rec, berr := le.backupManager.Backup(targetPath)
		if berr != nil {
			return tx, fmt.Errorf("failed to backup existing file at %s: %w", targetPath, berr)
		}
		op.BackupRecord = rec
	}

	// Ensure parent directory of target exists
	if err := le.fs.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return tx, fmt.Errorf("failed to create target parent directory for %s: %w", targetPath, err)
	}

	// 4. Create temp symlink
	tempPath := fmt.Sprintf("%s.tmp.%s", targetPath, generateRandomID())
	op.TempPath = tempPath

	if linker, ok := le.fs.(afero.Symlinker); ok {
		if err := linker.SymlinkIfPossible(sourcePath, tempPath); err != nil {
			_ = tx.Rollback()
			return tx, fmt.Errorf("failed to create temporary symlink from %s to %s: %w", sourcePath, tempPath, err)
		}
	} else {
		// Fallback for non-symlink supporting afero filesystems: write path text
		if err := afero.WriteFile(le.fs, tempPath, []byte(sourcePath), 0644); err != nil {
			_ = tx.Rollback()
			return tx, fmt.Errorf("failed writing fallback symlink path: %w", err)
		}
	}

	// 5. Atomic Rename tempPath -> targetPath
	if err := le.fs.Rename(tempPath, targetPath); err != nil {
		_ = tx.Rollback()
		return tx, fmt.Errorf("failed to atomically replace target path %s: %w", targetPath, err)
	}

	op.Status = OpStatusExecuted
	tx.Commit()
	return tx, nil
}

// IsBrokenSymlink checks if targetPath is a dangling symlink whose target source does not exist.
func (le *LinkerEngine) IsBrokenSymlink(targetPath string) (bool, error) {
	lstater, ok := le.fs.(afero.Lstater)
	if !ok {
		return false, nil
	}

	fi, _, err := lstater.LstatIfPossible(targetPath)
	if err != nil {
		return false, nil
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return false, nil // Not a symlink
	}

	// Symlink exists; check if target exists
	exists, _ := afero.Exists(le.fs, targetPath)
	return !exists, nil
}

// RepairSymlink repairs a broken or dangling symlink by re-creating it pointing to newSourcePath.
func (le *LinkerEngine) RepairSymlink(newSourcePath, targetPath string) (*Transaction, error) {
	isBroken, err := le.IsBrokenSymlink(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed checking symlink status for %s: %w", targetPath, err)
	}

	if !isBroken {
		// If not broken, still proceed with standard atomic symlink replacement
		return le.CreateSymlink(newSourcePath, targetPath)
	}

	// Remove broken link cleanly and recreate
	if !le.dryRun {
		_ = le.fs.Remove(targetPath)
	}
	return le.CreateSymlink(newSourcePath, targetPath)
}

func restoreFromBackup(fs afero.Fs, rec *BackupRecord) error {
	if rec == nil || rec.BackupPath == "" || rec.OriginalPath == "" {
		return nil
	}

	_ = fs.RemoveAll(rec.OriginalPath)

	bm := NewBackupManager(fs, "", "")
	info, err := fs.Stat(rec.BackupPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return bm.copyDir(rec.BackupPath, rec.OriginalPath)
	}
	return bm.copyFile(rec.BackupPath, rec.OriginalPath)
}

func generateRandomID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
