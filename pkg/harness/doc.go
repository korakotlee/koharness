// Package harness provides cross-client AI harness discovery, status tracking,
// and adapter path resolution for local developer workstations.
//
// It supports Google Antigravity (~/.gemini), Claude Code (~/.claude, ~/.claude.json),
// and OpenAI Codex (~/.codex), offering an abstracted filesystem layer via afero.Fs
// to enable isolated testing without modifying workstation configuration files.
package harness
