// Package symlink provides filesystem backup management, atomic symlink creation,
// dangling symlink detection and repair, dry-run previews, and transactional
// rollback mechanics for local AI client harness configuration assets.
//
// The package ensures that pre-existing workstation configuration files (such as
// Google Antigravity ~/.gemini, Claude Code ~/.claude, or OpenAI Codex ~/.codex) are
// safely backed up to timestamped archives under ~/.koharness/backups/YYYYMMDD-HHMMSS/
// before any file or symlink modification occurs. All operations support filesystem
// abstraction via afero.Fs for completely isolated unit and integration testing.
package symlink
