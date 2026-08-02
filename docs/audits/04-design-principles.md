# Phase 4: Software Design Principles Audit

**Repository**: `github.com/korakotlee/koharness`  
**Date**: 2026-07-29  
**Auditor**: Antigravity Principal Software Architect  

---

## Executive Overview
Phase 4 evaluates `koharness` against foundational software engineering principles: SOLID, DRY, KISS, YAGNI, Law of Demeter, and Defensive Programming.

---

## 1. Principle Compliance & Breach Analysis

### 1.1 Single Responsibility Principle (SRP)
- **Violation**: `cmd/init.go` combines CLI arg parsing, repository git cloning, filesystem scanning, TUI orchestration, symlinking loop, and global config persistence in a single function ([cmd/init.go:29-115](file:///Users/korakot/dev/koharness/cmd/init.go#L29-L115)).
- **Impact**: High. Modifying how repositories are scanned or symlinked requires touching `cmd/init.go`, making unit testing of orchestration logic difficult.
- **Refactoring Strategy**: Extract an `InitializerService` inside `pkg/harness` or `pkg/harvester` that accepts options and returns operation summaries. `InitCmd.RunE` should only validate CLI flags and pass data to `InitializerService`.

### 1.2 Don't Repeat Yourself (DRY)
- **Violation**: 
  - File/Directory copy routines duplicated in `pkg/harvester/creator.go` and `pkg/symlink/backup.go`.
  - Bubbletea list navigation & window resize handling duplicated between `pkg/tui/harvest_view.go` and `pkg/tui/init_view.go`.
- **Impact**: Medium. Logic changes in file permission copying or keybindings must be performed in multiple files, increasing drift risk.
- **Refactoring Strategy**:
  - Extract common file/folder copying to a shared `afero` utility helper.
  - Implement a reusable `tui.CapabilityListView` for TUI check-list management.

### 1.3 Interface Segregation Principle (ISP) & Open/Closed Principle (OCP)
- **Compliance**: **Excellent**. `harness.HarnessAdapter` interface ([pkg/harness/detector.go](file:///Users/korakot/dev/koharness/pkg/harness/detector.go)) allows adding new AI harness adapters (`AntigravityAdapter`, `ClaudeAdapter`, `CodexAdapter`) without modifying existing detector or scanner implementations.

### 1.4 Fail Fast & Defensive Programming
- **Compliance**: **High**. `symlink.Transaction` implement transactional rollbacks for filesystem operations, ensuring that if atomic renaming fails mid-operation, previous files are restored cleanly from backup records.

---

## 2. Summary Scorecard (Phase 4)
- SOLID Compliance: **8/10** (Great interface segregation, SRP can be improved in `cmd`)
- DRY Compliance: **7/10** (TUI views and copying routines have minor duplication)
- Defensive Resilience: **9/10** (Robust transaction rollback and atomic filesystem operations)
