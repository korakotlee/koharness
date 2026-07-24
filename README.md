# KoHarness

Centralize, version-control, and share AI capabilities across Claude Code, Google Antigravity, and OpenAI Codex / Copilot CLI.
Cross-harness manager for prompts, skills, workflows, and Model Context Protocol (MCP) configs.

---

## 1. Problem Statement & HMW

Managing prompt libraries, custom workflows, MCP server configs, and multi-file agent skills across an engineering team using different AI harnesses (Claude Code, Google Antigravity, OpenAI Codex / Copilot CLI) causes configuration drift and fragmented tooling.
Each AI tool stores configuration differently, but underneath the hood they all rely on standardized JSON/YAML configs, markdown files, and local execution paths.

**How Might We** build `koharness` as an open, friction-free CLI tool and dotfiles-style manager that centralizes, version-controls, and seamlessly syncs AI prompts, skills, and MCP configurations across multiple AI harnesses without breaking tool-specific file models or developer workflows?

---

## 2. Visual & UX Design Direction

- **Theme & Aesthetics:** Modern interactive terminal UI (TUI) with clean status tables, color-coded symlink diffs, and confirmation prompts built with Lip Gloss / Bubbletea.
- **Layout & Structure:**
  - `koharness init`: Interactive detection dashboard listing discovered AI harness clients, backup locations, and proposed symlinks.
  - `koharness create`: Harvesting dashboard to discover existing local capabilities (`~/.gemini`, `~/.claude.json`, `~/.codex`) and bootstrap a new repo.
  - `koharness sync`: Split view showing git status (clean vs dirty), remote rebase status, and capability symlink mapping.
  - `koharness doctor`: Interactive health inspector for broken symlinks and missing dependencies.
- **Key Interactions:** Dry-run preview before applying changes, interactive capability selection checklist, interactive approval gates before modifying client AI tool files.

---

## 3. Business & Product Scope

- **Target Persona:** Individual developers managing personal AI dotfiles and engineering teams standardizing shared AI workflows.
- **5-Second User Action:** Execute `koharness sync` to pull down team updates and link new prompts, skills, and MCP servers into local client AI harnesses.
- **MVP Scope:**
  - `koharness init`: Clones dotfiles repository to `~/.koharness/repo`, inspects client tools (Claude Code, Antigravity, Codex), backs up pre-existing configs to `~/.koharness/backups/`, and symlinks capabilities.
  - `koharness create`: Scans un-managed local capabilities, populates a fresh Git repository, backs up originals, and sets up symlinks.
  - `koharness sync`: Checks for uncommitted local changes, pulls with rebase, merges local MCP overrides, and updates symlinks across client harnesses.
  - `koharness doctor`: Audits active symlinks, missing environment variables, and script execution permissions.
  - `koharness lint`: Lints JSON syntax, file structure, and skill header schemas for CI/CD pipelines.
- **Not Doing (and Why):**
  - Proprietary cloud sync backend: Using standard Git repos keeps infrastructure zero-cost and self-hosted.
  - Automatic force-committing dirty user repos: Instructing users to create PRs prevents accidental code loss.

---

## 4. Technical Architecture & Data Strategy

### Directory Layout

```text
ai-workspace/
├── mcp/
│   ├── mcp.json            # Shared team MCP definitions with env expansion
│   ├── mcp.local.json      # Git-ignored local overrides & private paths
│   └── custom-servers/     # Custom Node/Python/Go MCP server implementations
├── prompts/                # Standardized System & Task Prompts (.md)
│   ├── code-review.md
│   └── refactoring.md
├── skills/                 # Multi-file portable tool packages
│   └── rails-migrations/
│       ├── SKILL.md        # Skill definition & frontmatter
│       ├── scripts/        # Executable Python/Ruby/Bash scripts
│       └── resources/      # Template files & assets
└── harnesses/              # Tool-specific adapter configs
    ├── claude-code/
    ├── antigravity/
    └── codex/
```

### Technology Stack, Testing & Release Strategy

- **Core Language:** Go (Golang) 1.22+ for cross-platform static binary distribution and zero runtime dependencies.
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra) for command routing, flag parsing, and shell completion.
- **Terminal UI (TUI):** [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) and [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) for interactive dashboards, status tables, and styled terminal diffs.
- **Git Integration:** [go-git/go-git/v5](https://github.com/go-git/go-git) for embedded Git clone, pull, rebase, and status checking.
- **JSON / YAML Schema Validation:** [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema) for validating MCP server JSON definitions and skill frontmatter.
- **Release Automation & Homebrew Tap:** [GoReleaser](https://goreleaser.com/) integrated with GitHub Actions (`.github/workflows/release.yml`) for multi-platform binary compilation and automated formula updates to custom Homebrew taps (`Formula/koharness.rb`).
- **Linting & Code Quality:**
  - **Go Codebase:** [golangci-lint](https://golangci-lint.run/) (aggregating `staticcheck`, `govet`, `gosec`, `errcheck`, `ineffassign`) for Go static code analysis.
  - **Repository Assets & Prompts:** Custom built-in `koharness lint` for schema validation (`mcp.json`), executable bit permissions (`chmod +x`), and skill frontmatter (`SKILL.md`), paired with `markdownlint-cli2` for prompt formatting.
- **Testing Framework:**
  - Standard Go `testing` package with `stretchr/testify` (assertions and mocks).
  - Isolated file system testing via `spf13/afero` in-memory mock file system.
  - End-to-end integration tests running CLI subcommands in disposable temporary directories (`t.TempDir()`).
- **CI / CD Automated Pipeline:** GitHub Actions matrix testing against macOS, Linux, and Windows runners with `golangci-lint-action`.

### Technical Workflow Commands

- `brew install koharness`
- `koharness init`
  - Clones targeted dotfiles repository to `~/.koharness/repo`.
  - Inspects client AI tools installed on system (Claude Code, Codex, Antigravity).
  - Backs up existing legacy configurations to `~/.koharness/backups/`.
  - Shows visual TUI findings and waits for explicit user confirmation.
  - Symlinks capabilities into client harness configuration directories.
- `koharness create`
  - Scans existing un-managed capability paths (`~/.gemini`, `~/.claude.json`, `~/.codex`).
  - Initializes a new Git repository at `~/.koharness/repo` (or current directory) with standard directory layout.
  - Presents interactive TUI checklist to select capabilities to commit versus store in `mcp.local.json`.
  - Backs up original local configs to `~/.koharness/backups/`, creates baseline Git commit, and replaces originals with symlinks.
- `koharness sync`
  - Checks if user dotfiles repo is dirty; alerts user and advises PR workflow if uncommitted changes exist.
  - Executes `git pull --rebase`.
  - Merges `mcp.json` with `mcp.local.json` and expands `${VAR}` environment variables.
  - Syncs updated capabilities to client AI tools.
- `koharness lint`
  - Validates JSON/YAML syntax, executable bits (`chmod +x`), and skill frontmatter for CI/CD checks.
- `koharness doctor`
  - Diagnoses symlink status, missing dependencies, and client harness health.

### Data Migration & Legacy Config Strategy

- **Backup on Init & Create:** `koharness init` and `koharness create` automatically snapshot existing files (`~/.claude.json`, Antigravity configs) to timestamped archives under `~/.koharness/backups/`.
- **Capability Harvesting on Create:** Extracts existing standalone skills, prompts, and MCP servers into a brand new Git repo when a team repository does not yet exist.
- **Layered Config Merging:** Preserves machine-specific credentials and local-only MCP servers by merging repository `mcp.json` defaults with local `mcp.local.json` overrides.
- **Environment Variable Expansion:** Expands `${DATABASE_URL}` dynamically to prevent committing secrets to repository remotes.

---

## 5. Key Assumptions & Open Questions

- [ ] Validate client harness config path changes across Claude Code, Antigravity, and Codex version updates.
- [ ] Implement wrapper script generators for client harnesses that do not support folder-based skill directories natively.
- [ ] Setup GitHub Actions workflow template using `koharness lint` for team repositories.
