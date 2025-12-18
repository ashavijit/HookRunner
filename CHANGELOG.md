# Changelog

All notable changes to HookRunner will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- **Lua Policy Scripting** - Write custom policies using embedded Lua scripts
  - Built-in functions: `block()`, `pass()`, `read_file()`, `match()`
  - `check(file, content)` callback for per-file validation
  - Sample policies: no-todo, no-secrets, no-debug-js, no-print-python, max-file-size, require-copyright
- **GitHub Action** - Official action for CI/CD integration (`ashavijit/hookrunner-action`)
- Sample Lua policies in `samples/lua-policies/`

### Changed
- Updated gopher-lua dependency for Lua VM support

---

## [0.25.0] - 2025-12-17

### Added
- **Clean Room Mode** (`--clean-room`) - Run hooks in isolated temp directory with only staged files
  - Provides CI parity for local development
  - Ensures hooks run against exact staged content

### Fixed
- Added nolint directive for gosec G204 in `CreateCleanRoom` function

---

## [0.21.0] - 2025-12-15

### Added
- **Lefthook-style Features**:
  - `run` field for inline shell commands
  - `script` field for external script execution
  - `root` field for hook working directory
  - `tags` for hook categorization
  - `exclude_tags` for skipping tagged hooks
  - Local config override support (`.hooks/hooks.local.yaml`)

---

## [0.20.0] - 2025-12-14

### Fixed
- Install scripts now use GitHub API for dynamic version detection
- Improved cross-platform installation reliability

---

## [0.19.0] - 2025-12-13

### Added
- **Dry-run Mode** (`--dry-run`) - Preview hooks without execution
- **Validate Command** (`hookrunner validate`) - Check configuration syntax
- **Policy Enforcement System**:
  - `max_files_changed` - Limit files per commit
  - `forbid_directories` - Block commits to specific paths
  - `forbid_files` - Regex patterns for forbidden files
  - `forbid_file_extensions` - Block specific extensions
  - `required_files` - Ensure files are present
  - `max_file_size_kb` - File size limits
  - `forbid_file_content` - Pattern-based content blocking
  - `commit_message` - Conventional commit validation
- **Secret Detection** - Built-in patterns for:
  - AWS Access Keys
  - GitHub Personal Access Tokens
  - OpenAI API Keys
  - Slack Tokens
  - Private Keys
  - Hardcoded Passwords
- **Install Scripts**:
  - `install.sh` for Linux/macOS
  - `install.ps1` for Windows PowerShell

---

## [0.18.0] - 2025-12-12

### Fixed
- Added `cmd/hookrunner` directory to git (was previously ignored)
- Fixed errcheck linter directive on `SaveToDisk`

### Added
- Comprehensive README with badges and comparison table

---

## [0.17.0] - 2025-12-11

### Added
- **CI/CD Integration**:
  - GitHub Actions workflow
  - Auto-release on push to main
  - Multi-OS build support (Windows, macOS, Linux)
- golangci-lint configuration
- Release automation script

---

## [0.16.0] - 2025-12-10

### Added
- Comprehensive test coverage for all packages
- Unit tests for:
  - Configuration loading (YAML/JSON)
  - DAG graph building and cycle detection
  - Policy evaluation
  - Git operations
  - Tool management
  - Presets

---

## [0.15.0] - 2025-12-09

### Added
- **Policy Engine** - Organizational rule enforcement
- **DAG Execution Engine**:
  - Parallel execution of independent hooks
  - Dependency ordering via `after` field
  - Cycle detection
  - Deterministic execution order

---

## [0.14.0] - 2025-12-08

### Added
- **Multi-Language Presets**:
  | Language | Tools |
  |----------|-------|
  | Go | gofmt, govet, golangci-lint |
  | Node.js | eslint, prettier, npm test |
  | Python | black, flake8, mypy, pytest |
  | Java | checkstyle, spotless, maven |
  | Ruby | rubocop, rspec |
  | Rust | cargo fmt, clippy, cargo test |
- `hookrunner presets` command to list available presets

---

## [0.13.0] - 2025-12-07

### Added
- **Advanced CLI Features**:
  - `hookrunner init` - Create configuration file
  - `hookrunner init --lang <language>` - Initialize with language preset
  - `hookrunner uninstall` - Remove installed hooks
  - `hookrunner run-cmd <tool> [args]` - Run tools directly
  - `hookrunner version` - Show version information
  - `hookrunner doctor` - Diagnose installation
- **Run Flags**:
  - `--verbose` - Detailed output
  - `--fix` - Enable auto-fix mode (uses `fix_args`)
  - `--all-files` - Run on all tracked files
  - `--no-fail-fast` - Continue after failures
  - `--quiet` - Suppress output except errors
- **Skip Conditions**:
  - `skip` field with environment variable
  - `SKIP` environment variable for comma-separated hooks

---

## [0.12.0] - 2025-12-06

### Added
- **Remote Policy Support**:
  - Fetch policies from HTTPS URLs
  - ETag-based conditional requests
  - SHA256-based disk caching
  - Local policy overrides

---

## [0.11.0] - 2025-12-05

### Added
- **Tool Management**:
  - Automatic tool downloading
  - Platform-specific binary downloads
  - SHA256 checksum verification
  - Version-specific caching
  - Fallback to system PATH

---

## [0.10.0] - 2025-12-04

### Added
- Initial public release
- Cross-platform support (Windows, macOS, Linux)
- YAML and JSON configuration support
- Pre-commit and commit-msg hook types
- File filtering via regex patterns
- Timeout support for hooks
- Environment variable injection

---

## Initial Development

### [0.1.0] - 2025-12-01

- Project initialization
- Basic hook execution framework
- Git integration for staged files

---

[Unreleased]: https://github.com/ashavijit/hookrunner/compare/v0.25.0...HEAD
[0.25.0]: https://github.com/ashavijit/hookrunner/compare/v0.21.0...v0.25.0
[0.21.0]: https://github.com/ashavijit/hookrunner/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/ashavijit/hookrunner/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/ashavijit/hookrunner/compare/v0.18.0...v0.19.0
[0.18.0]: https://github.com/ashavijit/hookrunner/releases/tag/v0.18.0
