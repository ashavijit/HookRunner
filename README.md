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
- Language presets for quick setup
- Parallel hook execution with dependency ordering
- Automatic tool download and caching
- SHA256 checksum verification
- Glob and regex file filtering
- Skip/Only conditions
- Environment variables support
- Auto-fix mode

## Installation

```bash
go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
```

## Quick Start

```bash
hookrunner init --lang go     # Create Go config
hookrunner init --lang nodejs # Create Node.js config
hookrunner init --lang python # Create Python config
hookrunner install            # Install git hooks
git commit -m "test"          # Hooks run automatically
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
hookrunner run pre-commit --quiet         # Minimal output
hookrunner run pre-commit --fix           # Auto-fix mode
hookrunner run pre-commit --no-fail-fast  # Continue on failure
SKIP=gofmt git commit                     # Skip specific hooks
```

## Configuration

### Basic Example

```yaml
hooks:
  pre-commit:
    - name: format
      tool: go
      args: ["fmt", "./..."]
      files: "\\.go$"
```

### Full Example

```yaml
tools:
  golangci-lint:
    version: 1.55.2
    install:
      windows: https://...zip
      linux: https://...tar.gz
      darwin: https://...tar.gz
    checksum: "sha256hash"

hooks:
  pre-commit:
    - name: lint
      tool: golangci-lint
      args: ["run"]
      fix_args: ["run", "--fix"]
      files: "\\.go$"
      exclude: "_test\\.go$"
      timeout: 2m
      after: format
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
| `glob` | Glob pattern for files |
| `exclude` | Regex to exclude files |
| `timeout` | Execution timeout |
| `after` | Dependency on another hook |
| `skip` | Skip if env var is set |
| `only` | Run only if env var is set |
| `env` | Environment variables |

## CI Integration

```bash
./hookrunner run pre-commit --all-files
```

## License

MIT License
