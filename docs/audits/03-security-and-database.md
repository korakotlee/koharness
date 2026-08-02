# Phase 3: Infrastructure, Security & Performance Audit

**Repository**: `github.com/korakotlee/koharness`  
**Date**: 2026-07-29  
**Auditor**: Antigravity Principal Software Architect  

---

## Executive Overview
Phase 3 reviews security posture, file permission safety, error handling robustness, and performance metrics across the `koharness` codebase.

---

## 1. Security & File System Safety Audit

### 1.1 Unmasked Terminal Output of Sensitive API Secrets
- **Location**: [pkg/tui/harvest_view.go:218-220](file:///Users/korakot/dev/koharness/pkg/tui/harvest_view.go#L218-L220), [pkg/mcp/validator.go:63-69](file:///Users/korakot/dev/koharness/pkg/mcp/validator.go#L63-L69)
- **Severity**: High
- **Confidence**: High
- **Description**: 
  - `mcp.ValidateConfig` detects secrets matching regex patterns (e.g. `sk-` OpenAI tokens, `ghp_` GitHub tokens) and stores the exact secret string in `ValidationIssue.Value`.
  - When printed or logged in TUI / debug outputs, raw secret tokens could be leaked into terminal scrollbacks or log files.
- **Evidence**:
  - `ValidationIssue.Value` set to raw secret: [pkg/mcp/validator.go:67](file:///Users/korakot/dev/koharness/pkg/mcp/validator.go#L67)
- **Recommendation**: Redact or obscure secret token values in `ValidationIssue` (e.g., `sk-***...`) before displaying them or returning validation messages.

### 1.2 Fixed File Permission Mode (`0755` / `0644`) for Sensitive Configuration Directories
- **Location**: 
  - [pkg/harvester/creator.go:86](file:///Users/korakot/dev/koharness/pkg/harvester/creator.go#L86)
  - [pkg/symlink/backup.go:63](file:///Users/korakot/dev/koharness/pkg/symlink/backup.go#L63)
- **Severity**: Medium
- **Confidence**: High
- **Description**: 
  - Directories created in `~/.koharness/repo` and `~/.koharness/backups` use standard `0755` permissions, making backup files readable by other local users on shared systems.
  - Secret overrides (`mcp.local.json`) containing API keys are written with `0644` permissions ([pkg/harvester/creator.go:187](file:///Users/korakot/dev/koharness/pkg/harvester/creator.go#L187)).
- **Evidence**:
  - Backup dir permissions: [pkg/symlink/backup.go:63](file:///Users/korakot/dev/koharness/pkg/symlink/backup.go#L63)
  - `mcp.local.json` write permissions: [pkg/harvester/creator.go:187](file:///Users/korakot/dev/koharness/pkg/harvester/creator.go#L187)
- **Recommendation**: Restrict directory permissions for backup sessions and local secret override files to `0700` / `0600` (`os.FileMode(0600)`).

---

## 2. Performance & Error Handling Audit

### 2.1 Synchronous External Process Call during UI Rendering (`GetUserEmail`)
- **Location**: [pkg/tui/banner.go:18-32](file:///Users/korakot/dev/koharness/pkg/tui/banner.go#L18-L32)
- **Severity**: Low
- **Confidence**: High
- **Description**: 
  - `RenderBanner()` calls `GetUserEmail()`, which synchronously executes `git config user.email` via `exec.Command`.
  - Although wrapped in `sync.Once`, the first call blocks TUI startup until the `git` process completes. If `git` hangs or takes long to initialize on slow disk/NFS, command responsiveness drops.
- **Evidence**: [pkg/tui/banner.go:20](file:///Users/korakot/dev/koharness/pkg/tui/banner.go#L20)
- **Recommendation**: Execute git config reading asynchronously or pass cached user context into banner renderer.

### 2.2 Atomic Symlink Replacement Fallback Behavior
- **Location**: [pkg/symlink/linker.go:207-213](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L207-L213)
- **Severity**: Medium
- **Confidence**: High
- **Description**: When `afero.Fs` does not implement `Symlinker`, `LinkerEngine.CreateSymlink` falls back to writing the source path string into a plain text file. While suitable for mock testing, on a live filesystem that fails symlinking, this creates plain text files instead of erroring out.
- **Evidence**: [pkg/symlink/linker.go:207-213](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L207-L213)
- **Recommendation**: Explicitly return an error if OS symlinking is unsupported rather than silently writing a path string file.

---

## 3. Summary Scorecard (Phase 3)
- Secret Protection: **Medium** (Potential secret exposure in log objects, standard permissions on secret files)
- Process Security: **High** (Safe input parsing, no shell injection vectors in exec commands)
- Performance: **High** (Sub-second execution, lightweight dependencies)
