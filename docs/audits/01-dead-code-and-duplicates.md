# Phase 1: Structural Hygiene & Redundancy Audit

**Repository**: `github.com/korakotlee/koharness`  
**Date**: 2026-07-29  
**Auditor**: Antigravity Principal Software Architect  

---

## Executive Overview
Phase 1 evaluates the codebase for dead code, unreferenced constants/types, unused struct fields, and duplicated implementation logic across packages.

---

## 1. Dead Code & Unused Entities

### 1.1 Unused Transaction Field and Operational Status Lifecycle
- **Location**: [pkg/symlink/linker.go:39](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L39), [pkg/symlink/linker.go:18-20](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L18-L20)
- **Severity**: Low
- **Confidence**: High
- **Description**: 
  - `Transaction.ID` is generated via `generateRandomID()` upon transaction initialization, but `ID` is never read, serialized, or logged across the entire codebase.
  - `OpStatusPending` and `OpStatusExecuted` are defined as constants and assigned to `OperationLog.Status`, but `op.Status` is never queried except when skipping `OpStatusRolledBack` during `Rollback()`.
- **Evidence**:
  - `ID` declaration: [pkg/symlink/linker.go:39](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L39)
  - `OpStatus` constants: [pkg/symlink/linker.go:18-20](file:///Users/korakot/dev/koharness/pkg/symlink/linker.go#L18-L20)
- **Recommendation**: Either utilize `Transaction.ID` and `OpStatus` in diagnostic logs/UI reporting, or remove unused struct fields and constants to simplify state management.

---

## 2. Duplicate & Conflicting Logic

### 2.1 Duplicated File & Directory Copy Routines
- **Location**: 
  - [pkg/harvester/creator.go:202-253](file:///Users/korakot/dev/koharness/pkg/harvester/creator.go#L202-L253)
  - [pkg/symlink/backup.go:177-227](file:///Users/korakot/dev/koharness/pkg/symlink/backup.go#L177-L227)
- **Severity**: Medium
- **Confidence**: High
- **Description**: Both `harvester.Creator` and `symlink.BackupManager` implement private `copyFile` and `copyDir` helper functions working on `afero.Fs`. The implementations duplicate file opening, mode preservation, buffer copying (`io.Copy`), and directory recursive iteration logic line-by-line.
- **Evidence**:
  - `harvester.Creator.copyFile` / `copyDir`: [pkg/harvester/creator.go:202-253](file:///Users/korakot/dev/koharness/pkg/harvester/creator.go#L202-L253)
  - `symlink.BackupManager.copyFile` / `copyDir`: [pkg/symlink/backup.go:177-227](file:///Users/korakot/dev/koharness/pkg/symlink/backup.go#L177-L227)
- **Recommendation**: Refactor filesystem copy utilities into a shared `pkg/fsutil` or `pkg/symlink` helper function (e.g., `CopyFile(fs afero.Fs, src, dst string) error`) to avoid maintenance divergence when permissions or file attributes are handled.

### 2.2 Duplicated `scanDirectory` Logic in Harvester Scanner
- **Location**: [pkg/harvester/scanner.go:131-158](file:///Users/korakot/dev/koharness/pkg/harvester/scanner.go#L131-L158), [pkg/harvester/scanner.go:171-173](file:///Users/korakot/dev/koharness/pkg/harvester/scanner.go#L171-L173)
- **Severity**: Low
- **Confidence**: High
- **Description**: `scanner.go` delegates MCP directory scanning to `scanDirectory`, but duplicates directory existence checking and entry iteration.
- **Evidence**: [pkg/harvester/scanner.go:171-173](file:///Users/korakot/dev/koharness/pkg/harvester/scanner.go#L171-L173)
- **Recommendation**: Standardize folder scanning delegates across capability types (`skills`, `prompts`, `mcp`).

---

## 3. Summary Scorecard (Phase 1)
- Dead Code Density: **Low** (Clean overall codebase with minimal unused symbols)
- Duplication Level: **Medium** (Filesystem copying duplicated across packages)
- Actionable Cleanups: 2 high-value refactoring opportunities identified.
