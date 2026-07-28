# KoHarness

Centralize, version-control, and share AI capabilities across Claude Code, Google Antigravity, and OpenAI Codex / Copilot CLI.

`koharness` is a cross-harness CLI manager for prompts, skills, workflows, and Model Context Protocol (MCP) configurations.

---

## The Problem: Fractured Team AI Tooling

In most engineering teams, AI setups are fragmented across developers:

- **Alice** crafts custom system prompts and bash scripts for Claude Code inside `~/.claude/`.
- **Bob** configures specialized multi-file skills and workflows for Google Antigravity.
- **Charlie** sets up custom Model Context Protocol (MCP) servers locally for Copilot CLI to inspect databases and logs.

When Alice builds a great debugging workflow or Charlie writes an MCP config for internal tools, there is no easy way to share it with the team. Capabilities stay trapped in individual dotfiles, private local setups, or loose slack snippets. When a new engineer joins, onboarding their AI setup means manually copy-pasting JSON files and fixing path mismatches.

Team memory and AI tooling remain siloed, leading to duplicated effort and inconsistent developer environments across the team.

## The Solution: A Shared AI Capability Repository

`koharness` solves this by turning team AI capabilities into a single, version-controlled dotfiles repository.

Instead of managing local configs in isolation:

1. **Centralize & Version-Control:** Store prompts, skills, workflows, and MCP configurations in a shared Git repository (`ai-workspace/` or dotfiles).
2. **Cross-Harness Sync:** Define capabilities once and `koharness` symlinks and adapts them dynamically across Claude Code, Google Antigravity, and OpenAI Codex.
3. **Instant Onboarding:** New team members run `koharness init` to clone and link team capabilities into their environment immediately.
4. **Layered Configs:** Shared team definitions (`mcp.json`) merge cleanly with local personal overrides (`mcp.local.json`), giving engineers flexibility without sacrificing team standardization.

With `koharness`, teams build a shared AI memory and tool stack that evolves together in Git.

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
| `koharness create [repo-path]` | Harvest unmanaged local capabilities, bootstrap a new dotfiles repo, back up original assets, and setup symlinks. |
| `koharness repo` | Navigate to, launch a subshell inside, or open your local dotfiles repository directory. |
| `koharness sync` | Pull remote updates via rebase, merge local MCP overrides, expand env vars, and update symlinks. |
| `koharness doctor` | Audit active symlinks, verify executable permissions, and check missing environment variables. |
| `koharness lint` | Lint JSON/YAML schemas, executable script bits (`chmod +x`), and skill frontmatter in CI/CD. |

### Creating a New Repository (`koharness create`)

To harvest your existing local skills, prompts, and MCP servers into a new dotfiles repository:

```bash
# Interactively harvest unmanaged capabilities into ~/.koharness/repo
koharness create

# Harvest into a custom repository path in non-interactive mode
koharness create ~/dev/my-ai-dotfiles --non-interactive
```

### Initializing an Existing Repository (`koharness init`)

To clone an existing team or personal dotfiles repository and link its capabilities into your local client harnesses (`~/.gemini`, `~/.claude.json`, `~/.codex`):

```bash
# Clone a git repository into ~/.koharness/repo and run the TUI setup dashboard
koharness init https://github.com/myteam/ai-dotfiles.git

# Clone into a custom directory path
koharness init https://github.com/myteam/ai-dotfiles.git ~/custom/ai-repo

# Overwrite/backup existing target directory if present
koharness init https://github.com/myteam/ai-dotfiles.git --force

# Run in non-interactive mode (ideal for automated developer environment setup)
koharness init https://github.com/myteam/ai-dotfiles.git --non-interactive
```

### Accessing Your Repository (`koharness repo`)

Navigate into or launch operations in your central dotfiles repository (`~/.koharness/repo` or custom path):

```bash
# Launch an interactive subshell ($SHELL) inside your repository
koharness repo

# Print raw absolute path to stdout (ideal for shell functions or command substitution)
cd "$(koharness repo -p)"

# Open the repository directory in your preferred code editor ($EDITOR or VS Code)
koharness repo -e code

# Generate shell integration helper snippet for zsh/bash/fish
koharness repo --shell-init
```



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
