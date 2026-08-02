# KoHarness CLI Usage & Testing Guide

This document explains how to build, test, and run the `koharness` CLI application.

## Prerequisites

- **Go**: 1.24.0 or newer (or via `mise` / `brew`).

## 1. Building the Binary

Build the executable locally in the project root:

```bash
go build -o koharness .
```

To test custom release version injection via `ldflags`:

```bash
go build -ldflags "-X github.com/korakotlee/koharness/pkg/version.Version=v1.0.0 -X github.com/korakotlee/koharness/pkg/version.Commit=a1b2c3d" -o koharness .
```

## 2. Running Automated Tests

Run all unit tests across packages:

```bash
go test -v ./...
```

Run tests for specific packages:

```bash
# Test version resolution logic
go test -v ./pkg/version/...

# Test terminal UI styles and banner rendering
go test -v ./pkg/tui/...

# Test Cobra root command flags
go test -v ./cmd/...

# Test atomic backup & symlink engine
go test -v ./pkg/symlink/...

# Test MCP configuration merger, env expander, and validator
go test -v ./pkg/mcp/...
```

## 3. Testing CLI Commands & Flags

### Version Metadata (`--version`, `-V`)
Displays the 5-line block-art `KH` header banner along with the resolved version, commit ID, build timestamp, and git user email.

```bash
./koharness --version
./koharness -V
```

### Help Menu (`--help`, `-h`)
Renders the colorized help interface.

```bash
./koharness --help
./koharness -h
```

### Persistent Flags
Test passing persistent flags to the root command:

```bash
# Enable verbose debug logging
./koharness --verbose

# Specify a custom config file path
./koharness --config ./config.yaml
```

## 4. Testing Terminal UI Behaviors

### Dynamic Git User Email Sourcing
The metadata panel resolves `git config user.email` at runtime.
Test overriding your local git email to verify dynamic rendering:

```bash
git config user.email "dev@example.com"
./koharness --version
```

### Unstyled Plain Output (`NO_COLOR`)
Test rendering in environments without ANSI colors:

```bash
NO_COLOR=1 ./koharness --version
```

## 5. Proton Pass CLI Integration & MCP Credential Injection

KoHarness integrates with Proton Pass CLI (`pass-cli`) to resolve external secret reference URIs during `koharness sync` operations. This allows dotfiles repositories to store template configurations using secret URIs rather than unencrypted API keys.

### Installing and Setting Up Proton Pass CLI

1. **Install `pass-cli`**:
   - macOS / Homebrew:
     ```bash
     brew install protonpass-cli
     ```
   - Binary Download: Download the latest binary from the official Proton Pass releases repository and ensure `pass-cli` is added to your system `$PATH`.

2. **Authenticate with Proton Pass**:
   - Log in to your Proton account using the CLI:
     ```bash
     pass-cli login
     ```
   - Verify authentication status:
     ```bash
     pass-cli status
     ```

### Secret URI Syntax

Secret URIs follow the format:

```
pass://<vault>/<item>/<field>
```

- `<vault>`: Name of your Proton Pass vault (e.g., `Development`).
- `<item>`: Name of the stored login or secret item (e.g., `Anthropic`).
- `<field>`: Field name containing the secret (e.g., `api_key` or `password`).

### Example: Setting Up an MCP Server with Proton Pass and KoHarness

1. **Create the Secret in Proton Pass**:
   In your Proton Pass vault `Development`, create a secret item `Anthropic` with a custom field `api_key` set to your Anthropic API key (`sk-ant-api03-...`).

2. **Define Template Configuration in your KoHarness Repository**:
   In your dotfiles repository at `mcp/anthropic.json`:

   ```json
   {
     "mcpServers": {
       "anthropic": {
         "command": "npx",
         "args": ["-y", "@modelcontextprotocol/server-anthropic"],
         "env": {
           "ANTHROPIC_API_KEY": "pass://Development/Anthropic/api_key"
         }
       }
     }
   }
   ```

3. **Run `koharness sync`**:
   ```bash
   koharness sync
   ```
   KoHarness automatically detects `pass-cli`, resolves `pass://Development/Anthropic/api_key`, and injects the actual API key string into your local client harness configurations (e.g., `~/.gemini/antigravity-ide/mcp/anthropic.json`).

### Fallback Behavior

If `pass-cli` is not installed on the system, logged out, or if a specific secret key is missing:
- KoHarness outputs a non-fatal `[CREDENTIAL WARNING]` pill message.
- Sync falls back to existing local environment variables or `mcp.local.json` entries.
- The `koharness sync` process completes smoothly without failing.

