# HookRunner Manual Test Results
Generated: 2025-12-18 00:12:31

## Build
```
go build -o hookrunner.exe ./cmd/hookrunner
OK
```

## 1. Version
```
hookrunner 0.19.0 (dev) built unknown
```

## 2. Help
```
A cross-platform pre-commit hook system with YAML/JSON configuration
Supports: Go, Node.js, Python, Java, Ruby, Rust

Usage:
  hookrunner [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  doctor      Diagnose setup
  help        Help about any command
  init        Create config file (use --lang for language preset)
  install     Install git hooks
  list        List configured hooks
  policy      Manage policies
  presets     List available language presets
  run         Run specified hook
  run-cmd     Run a tool directly
  uninstall   Remove installed git hooks
  validate    Validate configuration file
  version     Show version information

Flags:
  -h, --help   help for hookrunner
```

## 3. Doctor
```
HookRunner Doctor
=================

[OK] Git repository detected
[OK] Config file: C:\Users\AVIJIT\Desktop\HookRunner\hooks.yaml
[OK] Hooks configured: 3
[OK] Tools defined: 1
[OK] Cache directory exists

Version: 0.19.0
Supported: go, nodejs, python, java, ruby, rust
```

## 4. Presets
```
Available Language Presets:
===========================

  go         Go language hooks (gofmt, govet, golangci-lint)
  nodejs     Node.js hooks (eslint, prettier)
  python     Python hooks (black, flake8, mypy, pytest)
  java       Java hooks (checkstyle, spotless, maven test)
  ruby       Ruby hooks (rubocop, rspec)
  rust       Rust hooks (cargo fmt, clippy, test)

Usage: hookrunner init --lang <language>
```

## 5. List Hooks
```
Config: C:\Users\AVIJIT\Desktop\HookRunner\hooks.yaml

pre-commit:
  - gofmt (tool: go)
  - govet (tool: go) (after: gofmt)

pre-push:
  - test (tool: go)

Tools:
  - golangci-lint v1.55.2
```

## 6. Validate Config
```
Validating configuration...

[OK] Config file: C:\Users\AVIJIT\Desktop\HookRunner\hooks.yaml
[OK] Hooks configured: 3
[OK] pre-commit DAG is valid
[OK] pre-push DAG is valid
[OK] Tool 'go' found

[OK] Configuration is valid
```

## 7. Policy List
```
Policy Configuration:
=====================

Type: raw

Remote Policies:
  - https://hookrunner.avijitsen.site/policies/sensitive-secrets.yaml

Local Policies:
  - commit-format

Loaded Remote Policies:
  ✓ secret-scan@1
```

## 8. Dry Run
```
hookrunner run pre-commit --dry-run --all-files

Dry-run mode: showing hooks that would execute

Level 1:
  ▶ gofmt
      tool: go
      args: [fmt ./...]

Level 2:
  ▶ govet
      tool: go
      args: [vet ./...]
```

## 9. Run Pre-commit
```
hookrunner run pre-commit --all-files

[PASS] gofmt (224ms)
[PASS] govet (646ms)

Ran 2 hooks in 870ms (2 passed, 0 failed)
[PASS] policies
```

## 10. Run Help
```
Run specified hook

Usage:
  hookrunner run [hook-type] [flags]

Flags:
      --all-files      Run on all files
      --dry-run        Show what would run without executing
      --fix            Run in fix mode
  -h, --help           help for run
      --no-color       Disable colored output
      --no-fail-fast   Continue on failure
  -q, --quiet          Quiet output
      --verbose        Verbose output
```

## 11. Go Tests
```
ok   github.com/ashavijit/hookrunner/internal/config   0.358s
ok   github.com/ashavijit/hookrunner/internal/dag      0.311s
ok   github.com/ashavijit/hookrunner/internal/executor 0.566s
ok   github.com/ashavijit/hookrunner/internal/git      0.767s
ok   github.com/ashavijit/hookrunner/internal/policy   0.420s
ok   github.com/ashavijit/hookrunner/internal/presets  0.277s
ok   github.com/ashavijit/hookrunner/internal/tool     0.556s
ok   github.com/ashavijit/hookrunner/internal/version  0.286s

All 8 packages PASS
```

## 12. Secret Detection Test
```
echo "AKIAIOSFODNN7EXAMPLE" > test_secrets.txt
git add test_secrets.txt
hookrunner run pre-commit

[FAIL] policies
  ✗ [secret_detected] AWS Access Key detected in test_secrets.txt - remove before committing
  ✗ [secret_detected] Hardcoded Password detected in test_secrets.txt - remove before committing
  ✗ [secret_detected] GitHub Personal Access Token detected in test_secrets.txt - remove before committing

Exit code: 1 (blocked)
```

## 13. Policy Tests - All Supported Rules

### 13.1 regex_block (Remote Policy)
Blocks secrets like AWS keys, GitHub tokens, passwords, private keys.
```yaml
rules:
  regex_block:
    - "AKIA[0-9A-Z]{16}"
    - "-----BEGIN PRIVATE KEY-----"
    - "(?i)password="
    - "ghp_[A-Za-z0-9_]{36}"
```
Test: Detected AWS keys, passwords, GitHub tokens in staged files.

### 13.2 forbid_directories (Local Policy)
Blocks commits containing vendor/ or node_modules/ directories.
```yaml
rules:
  forbid_directories:
    - vendor/
    - node_modules/
```
Test: Files in vendor/ or node_modules/ trigger violation.

### 13.3 forbid_file_extensions (Local Policy)
Blocks commits with .exe or .dll files.
```yaml
rules:
  forbid_file_extensions:
    - .exe
    - .dll
```
Test: Binary files with blocked extensions trigger violation.

### 13.4 max_files_changed (Local Policy)
Limits commits to 50 files maximum.
```yaml
rules:
  max_files_changed: 50
```
Test: Commits with >50 files are blocked.

### 13.5 commit_message (Local Policy)
Enforces conventional commit format.
```yaml
rules:
  commit_message:
    regex: "^(feat|fix|chore|docs|refactor|test|build|ci):"
    error: "Commit must follow conventional format"
```
Test: Non-conventional commits are blocked.

### 13.6 forbid_file_content
Blocks specific patterns in file content.
```yaml
rules:
  forbid_file_content:
    - pattern: "TODO"
      description: "Remove TODOs before commit"
```

### 13.7 max_file_size_kb
Limits file size in commits.
```yaml
rules:
  max_file_size_kb: 1024
```

### 13.8 required_files
Ensures specific files exist.
```yaml
rules:
  required_files:
    - README.md
    - LICENSE
```

### 13.9 enforce_hooks
Requires specific hooks to run.
```yaml
rules:
  enforce_hooks:
    - gofmt
    - govet
```

### 13.10 hook_time_budget_ms
Limits hook execution time.
```yaml
rules:
  hook_time_budget_ms:
    gofmt: 5000
    govet: 10000
```

### 13.11 max_parallel_hooks
Limits parallel hook execution.
```yaml
rules:
  max_parallel_hooks: 4
```

## Policy Sources

| Type | Source |
|------|--------|
| Remote | https://hookrunner.avijitsen.site/policies/sensitive-secrets.yaml |
| Local | Defined in hooks.yaml under policies.localPolicies |

## Secret Patterns Detected

| Pattern | Description |
|---------|-------------|
| `AKIA[0-9A-Z]{16}` | AWS Access Key |
| `-----BEGIN PRIVATE KEY-----` | Private Key |
| `(?i)password=` | Hardcoded Password |
| `ghp_[A-Za-z0-9_]{36}` | GitHub Personal Access Token |
| `gho_[A-Za-z0-9_]{36}` | GitHub OAuth Token |
| `sk-[A-Za-z0-9]{48}` | OpenAI API Key |
| `xox[baprs]-[A-Za-z0-9-]+` | Slack Token |
