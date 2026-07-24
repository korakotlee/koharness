# Code Review Playbook

This playbook defines the code review standards and checklist for the **KoHarness** codebase. All pull requests, automated changes, and task completions must satisfy these criteria before being merged or archived.

---

## 1. Objectives

- **Correctness & Reliability:** Ensure code fulfills functional requirements without regressions or edge-case failures.
- **Maintainability & Readability:** Code must be clear, idiomatic Go, thoroughly documented with Godoc comments, and cleanly structured.
- **Cross-Platform Safety:** Symlinks, file paths, and environment variable expansions must operate seamlessly across macOS, Linux, and Windows.
- **OpenSpec Alignment:** Verify implementation matches delta specs and acceptance criteria.

---

## 2. Code Review Checklist

### A. Architecture & Structural Design
- [ ] **CLI & TUI Separation:** Command handling (Cobra) must be decoupled from UI rendering (Bubbletea / Lip Gloss) and core domain logic.
- [ ] **Filesystem Abstraction:** File operations must use `afero.Fs` abstraction to allow unit testing without mutating host filesystems.
- [ ] **Adapter Pattern:** Client harness logic (Claude Code, Antigravity, Codex) must implement cleanly defined interfaces.

### B. Go Code Quality & Style
- [ ] **Godoc Documentation:** All exported packages, structs, interfaces, functions, and constants MUST have clear Godoc comments.
- [ ] **Error Handling:** Errors must be checked explicitly. Wrap errors with contextual information using `fmt.Errorf("...: %w", err)`.
- [ ] **Zero Ignored Errors:** No swallowed exceptions or unhandled return errors (`_ = functionCall()` must be justified).
- [ ] **Naming Conventions:** Follow standard Go naming conventions (camelCase for unexported, PascalCase for exported, short receiver names).

### C. Security & Path Safety
- [ ] **Path Traversal Protection:** All file paths received from user inputs or config files must be sanitized and cleaned (`filepath.Clean`).
- [ ] **Symlink Safety:** Verify symlink targets before linking or overwriting to prevent pointing to unsafe external locations.
- [ ] **Secret Protection:** Ensure env expansion does not leak sensitive credentials into log outputs or repository commits.

### D. Testing & Quality Assurance
- [ ] **Unit Tests:** New features and bug fixes must include unit tests using `stretchr/testify` assertions.
- [ ] **Mocked I/O:** File system logic must be tested using `afero.NewMemMapFs()` or temporary directories (`t.TempDir()`).
- [ ] **Edge Cases:** Tests must cover missing files, permission errors, dirty git states, and malformed JSON/YAML configs.

### E. OpenSpec Verification
- [ ] **Requirement Coverage:** Verify that every requirement defined in the feature spec has corresponding test coverage.
- [ ] **Documentation Update:** Verify README.md or user-facing documentation has been updated if flags or workflows changed.

---

## 3. Verification Commands

Before approving any review, run the following verification suite:

```bash
# 1. Run all unit tests
go test -v ./...

# 2. Run static code analysis
golangci-lint run

# 3. Run koharness built-in linter
koharness lint
```

---

## 4. Review Verdict

- **APPROVED:** All checklist items pass, tests succeed, and linter is clean.
- **CHANGES REQUESTED:** Blockers identified (e.g. missing tests, unhandled errors, missing Godoc comments, or spec deviation).
