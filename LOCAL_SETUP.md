# Local Development Setup

Get started with HookRunner development in minutes.

## Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Git** - [Download](https://git-scm.com/downloads)
- **golangci-lint** (optional) - `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Quick Start

```bash
# Clone the repository
git clone https://github.com/ashavijit/hookrunner.git
cd hookrunner

# Build and install hooks
go build -o hookrunner ./cmd/hookrunner
./hookrunner install
```

---

## ⚡ Using FluxFile (Recommended)

We use [**FluxFile**](https://github.com/ashavijit/fluxfile) - a modern task runner that's cleaner than Make.

### Install FluxFile

```bash
# macOS / Linux (Homebrew)
brew install ashavijit/tap/flux

# Windows (Scoop)
scoop bucket add flux https://github.com/ashavijit/fluxfile
scoop install flux

# Or quick install script
curl -fsSL https://raw.githubusercontent.com/ashavijit/fluxfile/main/scripts/install.sh | sh
```

### Common Tasks

```bash
flux -l              # List all available tasks

# Building
flux -t build        # Build optimized binary (31% smaller)
flux -t build-dev    # Fast build with debug info
flux -t install      # Install to GOPATH/bin

# Testing
flux -t test         # Run all tests with coverage
flux -t test-race    # Run with race detector
flux -t lint         # Run golangci-lint

# Development
flux -w -t dev       # Watch mode - auto-rebuild on changes
flux -t hooks        # Install git hooks
flux -t doctor       # Check your dev environment

# Release
flux -t release-all  # Cross-compile for all platforms
flux -t size         # Compare binary sizes
flux -t clean        # Remove build artifacts
```

### Why FluxFile over Make?

| Feature | Makefile | FluxFile |
|---------|----------|----------|
| Syntax | Tab-sensitive, cryptic | Clean, YAML-like |
| Watch mode | External tools needed | Built-in (`-w` flag) |
| Caching | Manual | Automatic with inputs/outputs |
| Dependencies | Manual ordering | Automatic with `deps:` |
| Parallel | Complex syntax | Simple `parallel: true` |
| Cross-platform | Shell-specific | Works everywhere |

---

## Alternative: Using Make

If you prefer Make:

```bash
make build       # Build optimized binary
make build-dev   # Development build
make test        # Run tests
make lint        # Run linter
make install     # Install to GOPATH/bin
make clean       # Remove artifacts
```

---

## Alternative: Using PowerShell (Windows)

```powershell
.\build.ps1 build      # Build optimized binary
.\build.ps1 build-dev  # Development build
.\build.ps1 test       # Run tests
.\build.ps1 size       # Compare binary sizes
.\build.ps1 help       # Show all commands
```

---

## Project Structure

```
hookrunner/
├── cmd/hookrunner/     # CLI entry point
├── internal/
│   ├── cli/            # Command implementations
│   ├── config/         # Configuration parsing
│   ├── dag/            # DAG execution engine
│   ├── executor/       # Hook execution
│   ├── git/            # Git operations
│   ├── lua/            # Lua policy runner
│   ├── policy/         # Policy engine
│   ├── presets/        # Language presets
│   ├── tool/           # Tool management
│   └── version/        # Version info
├── samples/            # Sample configurations
│   └── lua-policies/   # Example Lua policies
├── scripts/            # Install scripts
├── FluxFile            # Task definitions (recommended)
├── Makefile            # Make tasks (alternative)
├── build.ps1           # PowerShell build (Windows)
└── hooks.yaml          # Example configuration
```

---

## Running Tests

```bash
# All tests
flux -t test
# or
go test ./... -v

# With coverage
flux -t test-coverage
# or
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# With race detector
flux -t test-race
# or
go test ./... -race
```

---

## Code Quality

```bash
# Format code
go fmt ./...

# Run vet
go vet ./...

# Run linter (requires golangci-lint)
flux -t lint
# or
golangci-lint run
```

---

## Building Releases

```bash
# Build for current platform (optimized, 31% smaller)
flux -t build

# Cross-compile all platforms
flux -t release-all

# Binaries are in dist/
ls dist/
```

---

## Git Hooks

HookRunner uses itself for pre-commit hooks:

```bash
# Install hooks
./hookrunner install

# Hooks run automatically on commit
git commit -m "feat: my change"

# Skip hooks temporarily
SKIP=lint,test git commit -m "wip: work in progress"
```

---

## Need Help?

- 📖 [README](README.md) - Full documentation
- 🐛 [Issues](https://github.com/ashavijit/hookrunner/issues) - Report bugs
- 💡 [FluxFile Docs](https://github.com/ashavijit/fluxfile) - Task runner documentation
