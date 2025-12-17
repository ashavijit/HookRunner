# HookRunner

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/ashavijit/hookrunner.svg)](https://pkg.go.dev/github.com/ashavijit/hookrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/ashavijit/hookrunner)](https://goreportcard.com/report/github.com/ashavijit/hookrunner)
[![Build Status](https://github.com/ashavijit/hookrunner/actions/workflows/ci.yml/badge.svg)](https://github.com/ashavijit/hookrunner/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ashavijit/hookrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/ashavijit/hookrunner)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/ashavijit/hookrunner)](https://github.com/ashavijit/hookrunner/releases)
[![Downloads](https://img.shields.io/github/downloads/ashavijit/hookrunner/total)](https://github.com/ashavijit/hookrunner/releases)

A cross-platform pre-commit hook system with DAG-based execution, policy enforcement, and remote policy support.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Policy System](#policy-system)
- [DAG Execution Engine](#dag-execution-engine)
- [CLI Reference](#cli-reference)
- [CI/CD Integration](#cicd-integration)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

HookRunner is a single-binary tool that manages git hooks with a focus on:

- **Deterministic execution** using DAG-based hook ordering
- **Policy enforcement** for organizational standards
- **Cross-platform support** for Windows, macOS, and Linux
- **Language-agnostic** with presets for Go, Node.js, Python, Java, Ruby, and Rust

### Comparison with Other Tools

| Feature | pre-commit | Husky | Lefthook | HookRunner |
|---------|------------|-------|----------|------------|
| Single Binary | No | No | Yes | Yes |
| Policy Engine | No | No | No | Yes |
| Remote Policies | No | No | No | Yes |
| DAG Execution | No | No | No | Yes |
| Multi-language | Yes | No | Yes | Yes |
| Cross-platform | No | No | Yes | Yes |

---

## Features

### DAG-Based Hook Execution

Hooks are modeled as a directed acyclic graph, enabling:

- Parallel execution of independent hooks
- Explicit dependency ordering via the `after` field
- Deterministic execution order across all platforms

### Policy Engine

Enforce organizational rules at commit time:

- Maximum files per commit
- Forbidden directories and files
- Required files
- Commit message format validation
- File size limits
- Content pattern detection

### Remote Policy Support

Fetch policies from remote URLs with caching:

- HTTPS-only for security
- ETag-based conditional requests
- SHA256-based disk caching
- Local policy overrides

### Tool Management

Automatic tool downloading and caching:

- Platform-specific binary downloads
- SHA256 checksum verification
- Version-specific caching
- Fallback to system PATH

### Language Presets

Quick setup for common languages:

| Language | Included Tools |
|----------|----------------|
| Go | gofmt, govet, golangci-lint |
| Node.js | eslint, prettier, npm test |
| Python | black, flake8, mypy, pytest |
| Java | checkstyle, spotless, maven |
| Ruby | rubocop, rspec |
| Rust | cargo fmt, clippy, cargo test |

---

## Installation

### Using Go

```bash
go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
```

### From Source

```bash
git clone https://github.com/ashavijit/hookrunner.git
cd hookrunner
go build -o hookrunner ./cmd/hookrunner
```

### Quick Install Script (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/ashavijit/hookrunner/master/scripts/install.sh | bash
```

### Pre-built Binaries

Download from the [Releases](https://github.com/ashavijit/hookrunner/releases) page.

---

## Quick Start

### 1. Initialize Configuration

```bash
# Create hooks.yaml with a language preset
hookrunner init --lang go

# Or create a minimal config
hookrunner init
```

### 2. Install Git Hooks

```bash
hookrunner install
```

### 3. Commit Your Code

```bash
git add .
git commit -m "feat: add new feature"
# Hooks run automatically
```

### 4. Manual Hook Execution

```bash
# Run pre-commit hooks
hookrunner run pre-commit

# Run on all files (not just staged)
hookrunner run pre-commit --all-files

# Run with auto-fix enabled
hookrunner run pre-commit --fix
```

---

## Configuration

HookRunner reads configuration from `hooks.yaml`, `hooks.yml`, or `hooks.json` in your project root.

### Complete Configuration Example

```yaml
# Tool definitions (optional - uses system PATH if not specified)
tools:
  golangci-lint:
    version: 1.55.2
    install:
      windows: https://github.com/.../golangci-lint-1.55.2-windows-amd64.zip
      linux: https://github.com/.../golangci-lint-1.55.2-linux-amd64.tar.gz
      darwin: https://github.com/.../golangci-lint-1.55.2-darwin-amd64.tar.gz
    checksum: "sha256:abc123..."

# Policy configuration
policies:
  type: raw

  # Remote policies (fetched and cached)
  policies:
    - url: https://policies.example.com/security-baseline.yaml

  # Local policies (override remote)
  localPolicies:
    - name: commit-format
      version: local
      rules:
        max_files_changed: 20
        forbid_directories:
          - vendor/
          - node_modules/
        commit_message:
          regex: "^(feat|fix|chore|docs|refactor|test):"
          error: "Commit must follow conventional format"

# Hook definitions
hooks:
  pre-commit:
    - name: format
      tool: go
      args: ["fmt", "./..."]
      files: "\\.go$"

    - name: lint
      tool: golangci-lint
      args: ["run"]
      after: format
      timeout: 2m
      env:
        GOPROXY: direct

    - name: test
      tool: go
      args: ["test", "./..."]
      after: lint
      skip: CI

  commit-msg:
    - name: validate-message
      tool: hookrunner
      args: ["policy", "check-message"]
```

### Hook Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique identifier for the hook |
| `tool` | string | Command or tool name to execute |
| `args` | []string | Arguments passed to the tool |
| `fix_args` | []string | Arguments used when --fix flag is set |
| `files` | string | Regex pattern to filter files |
| `exclude` | string | Regex pattern to exclude files |
| `glob` | string | Glob pattern for file matching |
| `timeout` | string | Maximum execution time (e.g., "2m", "30s") |
| `after` | string | Name of hook that must complete first |
| `skip` | string | Environment variable that skips this hook if set |
| `env` | map | Environment variables for execution |
| `fail_fast` | bool | Stop on first failure (default: true) |

---

## Policy System

### Policy Rules Reference

| Rule | Type | Description |
|------|------|-------------|
| `max_files_changed` | int | Maximum number of files in a commit |
| `max_file_size_kb` | int | Maximum file size in kilobytes |
| `forbid_files` | []string | Regex patterns for forbidden files |
| `forbid_directories` | []string | Forbidden directory paths |
| `forbid_file_extensions` | []string | Forbidden file extensions |
| `required_files` | []string | Files that must be present |
| `forbid_file_content` | []object | Patterns to detect in file content |
| `commit_message.regex` | string | Regex for commit message validation |
| `commit_message.error` | string | Custom error message |
| `enforce_hooks` | []string | Hooks that cannot be skipped |
| `hook_time_budget_ms` | map | Maximum execution time per hook |
| `max_parallel_hooks` | int | Limit parallel hook execution |

### Forbid File Content Example

```yaml
forbid_file_content:
  - pattern: "password\\s*=\\s*['\"]"
    description: "Hardcoded password detected"
  - pattern: "TODO.*HACK"
    description: "HACK comment found"
```

### Remote Policy Format

Remote policies follow the same schema as local policies:

```yaml
name: security-baseline
version: 1.2.0
description: Organization security standards

rules:
  forbid_files:
    - "\\.env$"
    - "\\.pem$"
  forbid_directories:
    - secrets/
  max_file_size_kb: 500
  commit_message:
    regex: "^(feat|fix|chore|docs):"
    error: "Use conventional commit format"

metadata:
  team: platform
  owner: security@example.com
```

### Policy Commands

```bash
# List configured policies
hookrunner policy list

# Force refresh remote policies
hookrunner policy fetch

# Clear policy cache
hookrunner policy clear-cache
```

---

## DAG Execution Engine

### How It Works

1. Hooks without dependencies run in parallel (Level 1)
2. Hooks with `after` field wait for their dependency
3. Multiple hooks can depend on the same parent
4. Cycle detection prevents infinite loops

### Example Execution

```yaml
hooks:
  pre-commit:
    - name: format      # Level 1 - runs in parallel
    - name: lint        # Level 1 - runs in parallel
    - name: security    # Level 1 - runs in parallel
    - name: test
      after: lint       # Level 2 - waits for lint
    - name: integration
      after: test       # Level 3 - waits for test
```

Execution diagram:

```
Level 1:  [format]  [lint]  [security]   (parallel)
              |       |
Level 2:      +-------+
                  |
                [test]
                  |
Level 3:    [integration]
```

---

## CLI Reference

### Commands

| Command | Description |
|---------|-------------|
| `init` | Create configuration file |
| `init --lang <language>` | Create config with language preset |
| `install` | Install git hooks to .git/hooks |
| `uninstall` | Remove installed hooks |
| `run <hook>` | Execute a specific hook |
| `run-cmd <tool> [args]` | Run a tool directly |
| `list` | Display configured hooks |
| `doctor` | Diagnose installation and configuration |
| `presets` | List available language presets |
| `policy list` | Show configured policies |
| `policy fetch` | Refresh remote policies |
| `policy clear-cache` | Clear cached policies |
| `version` | Display version information |

### Run Flags

| Flag | Description |
|------|-------------|
| `--all-files` | Run on all tracked files, not just staged |
| `--verbose` | Show detailed execution output |
| `--fix` | Enable auto-fix mode (uses fix_args) |
| `--no-fail-fast` | Continue execution after failures |
| `--quiet` | Suppress output except errors |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SKIP` | Comma-separated hooks to skip |
| `HOOKRUNNER_VERBOSE` | Enable verbose output |
| `HOOKRUNNER_DRY_RUN` | Show what would run without executing |

Example:

```bash
SKIP=lint,test git commit -m "wip: work in progress"
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Hooks
on: [push, pull_request]

jobs:
  hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install HookRunner
        run: go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest

      - name: Run pre-commit hooks
        run: hookrunner run pre-commit --all-files
```

### GitLab CI

```yaml
hooks:
  image: golang:1.21
  script:
    - go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
    - hookrunner run pre-commit --all-files
```

### Using Pre-built Binary

```yaml
- name: Install HookRunner
  run: |
    curl -sSL https://github.com/ashavijit/hookrunner/releases/latest/download/hookrunner-linux-amd64 -o hookrunner
    chmod +x hookrunner
    ./hookrunner run pre-commit --all-files
```

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Run linter: `golangci-lint run`
6. Commit: `git commit -m "feat: add my feature"`
7. Push: `git push origin feature/my-feature`
8. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Made by [ashavijit](https://github.com/ashavijit)
