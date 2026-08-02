# Phase 2: Architectural Consistency & AI Drift Audit

**Repository**: `github.com/korakotlee/koharness`  
**Date**: 2026-07-29  
**Auditor**: Antigravity Principal Software Architect  

---

## Executive Overview
Phase 2 inspects architectural boundaries, layer separation, UI leaks, and code consistency across `cmd`, `pkg/harness`, `pkg/harvester`, `pkg/symlink`, and `pkg/tui`.

---

## 1. Architectural Inconsistencies & AI Drift

### 1.1 UI Presentation Leak into CLI Domain logic (`cmd/init.go`)
- **Location**: [cmd/init.go:117-180](file:///Users/korakot/dev/koharness/cmd/init.go#L117-L180)
- **Severity**: High
- **Confidence**: High
- **Description**: 
  - `cmd/init.go` defines a private function `scanRepoCapabilities` which directly constructs and returns `[]tui.InitCapabilityItem`.
  - The CLI command layer (`cmd`) bypasses domain abstractions in `pkg/harness` or `pkg/harvester` and projects directory entries directly into presentation structs (`tui.InitCapabilityItem`).
  - This couples domain repository scanning directly to UI presentation models.
- **Evidence**:
  - `scanRepoCapabilities` signature: [cmd/init.go:117](file:///Users/korakot/dev/koharness/cmd/init.go#L117)
  - Direct construction of `tui.InitCapabilityItem`: [cmd/init.go:166-174](file:///Users/korakot/dev/koharness/cmd/init.go#L166-L174)
- **Recommendation**: Move repository capability scanning to `pkg/harvester` or `pkg/harness`, returning a domain struct (e.g., `[]harness.CapabilityAsset`), and map to TUI models inside `tui` or right before invoking `tui.RunInitView`.

### 1.2 Divergent Keyboard Key Handling in TUI Views
- **Location**: 
  - [pkg/tui/harvest_view.go:87-161](file:///Users/korakot/dev/koharness/pkg/tui/harvest_view.go#L87-L161)
  - [pkg/tui/init_view.go:102-170](file:///Users/korakot/dev/koharness/pkg/tui/init_view.go#L102-L170)
- **Severity**: Medium
- **Confidence**: High
- **Description**: 
  - `HarvestModel.Update` handles secret toggling with `"s"` key ([pkg/tui/harvest_view.go:141-144](file:///Users/korakot/dev/koharness/pkg/tui/harvest_view.go#L141-L144)). `InitModel.Update` does not support secret toggling.
  - Both views duplicate ~90 lines of identical Bubbletea keyboard navigation logic (`up`, `down`, `pgup`, `pgdown`, `home`, `end`, `space`, `a`, `enter`, `q`), but duplicate key event handling across two separate structs instead of using a shared list/table component.
- **Evidence**:
  - `HarvestModel.Update`: [pkg/tui/harvest_view.go:87-161](file:///Users/korakot/dev/koharness/pkg/tui/harvest_view.go#L87-L161)
  - `InitModel.Update`: [pkg/tui/init_view.go:102-170](file:///Users/korakot/dev/koharness/pkg/tui/init_view.go#L102-L170)
- **Recommendation**: Create a unified `tui.SelectableList` component to handle scrolling, item rendering, and keyboard interactions, reducing code duplication by over 300 lines across `harvest_view.go` and `init_view.go`.

### 1.3 Inconsistent Subshell Fallback across Commands
- **Location**: [cmd/repo.go:73-80](file:///Users/korakot/dev/koharness/cmd/repo.go#L73-L80)
- **Severity**: Low
- **Confidence**: High
- **Description**: `cmd/repo.go` falls back to `/bin/zsh` when `SHELL` is unset on Unix platforms. On macOS/Linux systems where `/bin/zsh` is missing or user runs `bash`/`fish`, hardcoding `/bin/zsh` without checking binary existence can cause runtime execution failures.
- **Evidence**: [cmd/repo.go:78](file:///Users/korakot/dev/koharness/cmd/repo.go#L78)
- **Recommendation**: Use `exec.LookPath` to verify shell binary existence or fallback gracefully to `/bin/sh`.

---

## 2. Summary Scorecard (Phase 2)
- Layer Separation: **Medium** (`cmd` contains domain scanning logic creating TUI types)
- UI Consistency: **Medium** (Duplicated TUI view controller logic)
- Architectural Drift: **Low** (Clean package separation overall)
