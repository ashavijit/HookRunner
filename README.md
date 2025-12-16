<div align="center">

# HookRunner

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ashavijit/hookrunner)](https://goreportcard.com/report/github.com/ashavijit/hookrunner)
[![Coverage](https://img.shields.io/badge/coverage-75%25-brightgreen)](https://github.com/ashavijit/hookrunner)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/ashavijit/hookrunner/pulls)

**Cross-platform pre-commit hook system with DAG execution and policy engine**

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#configuration) • [Contributing](#contributing)

</div>

---

## Why HookRunner?

> *"We model hooks as a DAG to ensure deterministic execution with maximal parallelism."*

| Feature | pre-commit | Husky | Lefthook | **HookRunner** |
|---------|------------|-------|----------|----------------|
| Single Binary | ❌ | ❌ | ✅ | ✅ |
| Policy Engine | ❌ | ❌ | ❌ | ✅ |
| DAG Execution | ❌ | ❌ | ❌ | ✅ |
| Multi-language | ✅ | ❌ | ✅ | ✅ |
| Cross-platform | ❌ | ❌ | ✅ | ✅ |

## Supported Languages

| Language | Tools |
|----------|-------|
| 🐹 Go | gofmt, govet, golangci-lint |
| 🟢 Node.js | eslint, prettier, npm test |
| 🐍 Python | black, flake8, mypy, pytest |
| ☕ Java | checkstyle, spotless, maven |
| 💎 Ruby | rubocop, rspec |
| 🦀 Rust | cargo fmt, clippy, cargo test |

## Features

- **DAG Execution Engine** - Deterministic hook ordering with maximal parallelism
- **Policy Engine** - Enforce org rules (max files, forbidden dirs, commit message format)
- **Single Binary** - Cross-platform (Windows, macOS, Linux)
- **YAML/JSON Config** - Simple, declarative configuration
- **Language Presets** - Quick setup for Go, Node.js, Python, Java, Ruby, Rust
- **Auto-fix Mode** - Automatic code formatting
- **Tool Management** - Auto-download and cache tools

## Installation

### From Source

```bash
git clone https://github.com/ashavijit/hookrunner.git
cd hookrunner
go build -o hookrunner ./cmd/hookrunner
```

### Go Install

```bash
go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
```

## Quick Start

```bash
# Initialize with language preset
hookrunner init --lang go

# Install git hooks
hookrunner install

# Commit triggers hooks automatically
git commit -m "feat: add new feature"
```

## DAG Execution Engine

Hooks are modeled as a dependency graph for deterministic execution:

```yaml
hooks:
  pre-commit:
    - name: format     # ──┐
    - name: lint       #   ├──▶ runs in parallel
    - name: security   # ──┘
    - name: test
      after: lint      # runs after lint completes
```

```
┌────────┐   ┌────────┐   ┌──────────┐
│ format │   │  lint  │   │ security │   Level 1 (parallel)
└────────┘   └───┬────┘   └──────────┘
                 │
                 ▼
            ┌────────┐
            │  test  │                    Level 2
            └────────┘
```

## Policy Engine

Enforce organizational rules at commit time:

```yaml
policies:
  max_files_changed: 20
  forbid_directories: ["vendor/", "generated/"]
  commit_message:
    regex: "^(feat|fix|chore|docs|refactor|test):"
    min_length: 10
    max_length: 72
```

**Output on violation:**
```
[FAIL] policies
  - [max_files_changed] too many files changed: 25 (max: 20)
  - [commit_message.regex] commit message does not match pattern
```

## Commands

| Command | Description |
|---------|-------------|
| `init --lang <lang>` | Create config with language preset |
| `presets` | List available language presets |
| `install` | Install git hooks |
| `uninstall` | Remove installed hooks |
| `run <hook>` | Run specified hook |
| `run-cmd <tool>` | Run a tool directly |
| `list` | List configured hooks |
| `doctor` | Diagnose setup |
| `version` | Show version |

## CLI Flags

```bash
hookrunner run pre-commit --all-files     # Run on all files
hookrunner run pre-commit --verbose       # Detailed output
hookrunner run pre-commit --fix           # Auto-fix mode
hookrunner run pre-commit --no-fail-fast  # Continue on failure
SKIP=gofmt git commit                     # Skip specific hooks
```

## Configuration

### Full Example

```yaml
tools:
  golangci-lint:
    version: 1.55.2
    install:
      windows: https://...zip
      linux: https://...tar.gz
    checksum: "sha256..."

policies:
  max_files_changed: 20
  forbid_directories: ["vendor/"]
  commit_message:
    regex: "^(feat|fix|chore):"

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
```

### Hook Fields

| Field | Description |
|-------|-------------|
| `name` | Hook identifier |
| `tool` | Command or tool name |
| `args` | Arguments to pass |
| `fix_args` | Arguments for --fix mode |
| `files` | Regex pattern to match files |
| `exclude` | Regex to exclude files |
| `timeout` | Execution timeout |
| `after` | Dependency on another hook |
| `skip` | Skip if env var is set |
| `env` | Environment variables |

## CI Integration

Since HookRunner is a single binary, just build and run in CI:

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

      - name: Build HookRunner
        run: go build -o hookrunner ./cmd/hookrunner

      - name: Run pre-commit hooks
        run: ./hookrunner run pre-commit --all-files
```

### GitLab CI

```yaml
hooks:
  image: golang:1.21
  script:
    - go build -o hookrunner ./cmd/hookrunner
    - ./hookrunner run pre-commit --all-files
```

### Or use go run directly

```yaml
# No binary needed
- run: go run ./cmd/hookrunner run pre-commit --all-files
```

## Test Coverage

```
Package     Coverage
─────────────────────
dag         100.0%
presets     100.0%
version     100.0%
config       85.7%
git          80.4%
policy       69.8%
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

Made with ❤️ by [ashavijit](https://github.com/ashavijit)

</div>
