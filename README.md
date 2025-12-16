# HookRunner

Cross-platform pre-commit hook system written in Go.

## Features

- Single binary, cross-platform (Windows, macOS, Linux)
- YAML/JSON configuration
- Parallel hook execution with dependency ordering
- Automatic tool download and caching
- SHA256 checksum verification
- Glob and regex file filtering
- Exclude patterns
- Skip/Only conditions
- Environment variables support
- Auto-fix mode
- Verbose/Quiet output
- Fail-fast toggle

## Installation

```bash
go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest
```

Or build from source:

```bash
git clone https://github.com/ashavijit/hookrunner.git
cd hookrunner
go build -o hookrunner ./cmd/hookrunner
```

## Quick Start

```bash
hookrunner init        # Create sample config
hookrunner install     # Install git hooks
git commit -m "test"   # Hooks run automatically
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Create sample hooks.yaml config |
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
    - name: gofmt
      tool: go
      args: ["fmt", "./..."]
      files: "\\.go$"

    - name: govet
      tool: go
      args: ["vet", "./..."]
      after: gofmt
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
      glob: "*.go"
      timeout: 2m
      after: format
      skip: CI
      only: LOCAL
      env:
        GOPROXY: direct
      pass_env: ["HOME", "PATH"]

  pre-push:
    - name: test
      tool: go
      args: ["test", "./..."]
      timeout: 5m
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
| `pass_env` | Forward specific env vars |

## CI Integration

```bash
./hookrunner run pre-commit --all-files
```

## License

MIT License
