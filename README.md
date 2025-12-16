# HookRunner

Cross-platform pre-commit hook system written in Go.

## Supported Languages

| Language | Tools |
|----------|-------|
| Go | gofmt, govet, golangci-lint |
| Node.js | eslint, prettier, npm test |
| Python | black, flake8, mypy, pytest |
| Java | checkstyle, spotless, maven |
| Ruby | rubocop, rspec |
| Rust | cargo fmt, clippy, cargo test |

## Features

- Single binary, cross-platform (Windows, macOS, Linux)
- YAML/JSON configuration
- **DAG Execution Engine** - deterministic hook ordering with maximal parallelism
- **Policy Engine** - enforce org rules at commit time
- Language presets for quick setup
- Automatic tool download and caching
- Glob and regex file filtering
- Skip/Only conditions

## Installation

```bash
go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
```

## Quick Start

```bash
hookrunner init --lang go     # Create Go config
hookrunner install            # Install git hooks
git commit -m "feat: test"    # Hooks run automatically
```

## DAG Execution Engine

Hooks are modeled as a dependency graph for deterministic execution:

```yaml
hooks:
  pre-commit:
    - name: format
    - name: lint
      after: format
    - name: test
      after: lint
```

Execution flow:
```
format ──▶ lint ──▶ test
```

Parallel hooks run concurrently:
```yaml
- name: lint      # ─┐
- name: security  # ─┼──▶ runs in parallel
- name: format    # ─┘
```

## Policy Engine

Enforce organizational rules at commit time:

```yaml
policies:
  max_files_changed: 20
  forbid_directories: ["vendor/", "generated/"]
  forbid_files: ["\\.env$"]
  commit_message:
    regex: "^(feat|fix|chore|docs|refactor|test):"
    min_length: 10
    max_length: 72
```

Policy violations block the commit:
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

## Flags

```bash
hookrunner run pre-commit --all-files     # Run on all files
hookrunner run pre-commit --verbose       # Detailed output
hookrunner run pre-commit --fix           # Auto-fix mode
hookrunner run pre-commit --no-fail-fast  # Continue on failure
SKIP=gofmt git commit                     # Skip specific hooks
```

## Full Configuration

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

## Test Coverage

```bash
go test ./... -v
# 27 tests passing
# - config: 8 tests
# - dag: 8 tests
# - policy: 11 tests
```

## License

MIT License
