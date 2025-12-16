# HookRunner

Cross-platform pre-commit hook system written in Go.

## Features

- Single binary, cross-platform (Windows, macOS, Linux)
- YAML/JSON configuration
- Parallel hook execution with dependency ordering
- Automatic tool download and caching
- SHA256 checksum verification
- Smart file filtering for staged files
- Timeout support

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

## Usage

### Initialize Configuration

Create a `hooks.yaml` in your repository root:

```yaml
tools:
  golangci-lint:
    version: 1.55.2
    install:
      windows: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-windows-amd64.zip
      linux: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-linux-amd64.tar.gz
      darwin: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-darwin-amd64.tar.gz

hooks:
  pre-commit:
    - name: gofmt
      tool: go
      args: ["fmt", "./..."]
      files: "\\.go$"

    - name: govet
      tool: go
      args: ["vet", "./..."]
      files: "\\.go$"
      after: gofmt
```

### Commands

```bash
hookrunner install              # Install git hooks
hookrunner run pre-commit       # Run pre-commit hooks on staged files
hookrunner run pre-commit --all-files  # Run on all files
hookrunner list                 # List configured hooks
hookrunner doctor               # Diagnose setup
```

### CI Integration

```bash
./hookrunner run pre-commit --all-files
```

## Configuration Reference

### Tools

```yaml
tools:
  tool-name:
    version: "1.0.0"
    install:
      windows: https://...
      linux: https://...
      darwin: https://...
    checksum: "sha256hash"  # optional
```

### Hooks

```yaml
hooks:
  pre-commit:  # or pre-push, commit-msg
    - name: hook-name
      tool: tool-name  # name from tools section or system command
      args: ["arg1", "arg2"]
      files: "\\.go$"  # regex to filter files
      timeout: 2m      # optional timeout
      after: other-hook  # dependency ordering
```

## License

MIT License
