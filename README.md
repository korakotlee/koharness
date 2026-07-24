# KoHarness

Centralize, version-control, and share AI capabilities across Claude Code, Google Antigravity, and OpenAI Codex / Copilot CLI.

`koharness` is a cross-harness CLI manager for prompts, skills, workflows, and Model Context Protocol (MCP) configurations.

---

## Key Features

- **Cross-Harness Sync:** Synchronize prompts, skills, and MCP configurations across Claude Code, Google Antigravity, and OpenAI Codex.
- **Interactive TUI:** Built with Lip Gloss and Bubbletea for intuitive detection dashboards, color-coded diff previews, and health inspection.
- **Dotfiles-Style Version Control:** Manage your team or personal AI capabilities in standard Git repositories with zero proprietary vendor lock-in.
- **Automated Backups:** Automatic snapshot and restoration of local configs (`~/.claude.json`, `~/.gemini`, `~/.codex`) before applying symlink modifications.
- **Layered MCP Management:** Support for shared team `mcp.json` merged dynamically with `mcp.local.json` overrides and environment variable expansion.
- **Schema & Quality Linting:** Built-in static validation for MCP JSON definitions, executable permissions, and skill frontmatter.

---

## Directory Structure

Standard layout for a `koharness` dotfiles repository:

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

---

## Command Usage

| Command | Description |
| :--- | :--- |
| `koharness init` | Interactively detect local AI harnesses, backup legacy configs, and symlink repository capabilities. |
| `koharness create` | Harvest unmanaged local capabilities, initialize a dotfiles repo, and setup symlinks. |
| `koharness sync` | Pull remote updates via rebase, merge local MCP overrides, expand env vars, and update symlinks. |
| `koharness doctor` | Audit active symlinks, verify executable permissions, and check missing environment variables. |
| `koharness lint` | Lint JSON/YAML schemas, executable script bits (`chmod +x`), and skill frontmatter in CI/CD. |

---

## Technology Stack

- **Core Language:** Go 1.22+
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra)
- **Terminal UI (TUI):** [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) & [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- **Git Engine:** [go-git/go-git/v5](https://github.com/go-git/go-git)
- **Filesystem Abstraction:** [spf13/afero](https://github.com/spf13/afero)
- **Schema Validation:** [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)

---

## Development & Playbooks

For guidelines on building features or reviewing code, see:
- [Feature Development Playbook](file:///Users/korakot/dev/koharness/PLAYBOOKS/feature-development.md)
- [Code Review Playbook](file:///Users/korakot/dev/koharness/PLAYBOOKS/code-review.md)
