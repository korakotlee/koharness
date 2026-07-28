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
