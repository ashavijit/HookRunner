# Product Requirements Document (PRD)

## Product Name

**HookRunner** (working name)

---

## 1. Problem Statement

Modern development teams rely on pre-commit hooks to enforce code quality, security, and consistency. However:

* Native git hooks are **not cross-platform friendly** (Windows vs Unix shells)
* Tooling often requires **global installations**
* Behavior differs between **local machines and CI**
* Hooks become slow, flaky, or hard to maintain

Existing tools partially solve this, but teams still struggle with:

* Reproducibility
* Performance
* Ease of onboarding

---

## 2. Goals & Non-Goals

### Goals

* Cross-platform (Windows, macOS, Linux)
* Single self-contained binary (written in Go)
* Declarative configuration using **YAML or JSON**
* Zero global tool dependency
* Same behavior locally and in CI
* Fast execution with smart file filtering

### Non-Goals

* Replace CI systems
* Provide IDE integrations (out of scope for v1)
* Execute untrusted remote scripts

---

## 3. Target Users

* Backend engineers (Go, Python, Java, Node)
* Platform / DevOps teams
* Monorepo maintainers
* Open-source maintainers

---

## 4. High-Level Architecture

```
Repo
 ├─ .hooks/
 │   ├─ hookrunner(.exe)
 │   ├─ hooks.yaml
 │   └─ cache/
 ├─ .git/hooks/pre-commit
 └─ ci/
```

---

## 5. Configuration Design (YAML / JSON)

### 5.1 Config File

* Supported formats: `hooks.yaml`, `hooks.yml`, `hooks.json`
* Auto-detected at runtime

### 5.2 Example (YAML)

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

    - name: golangci-lint
      tool: golangci-lint
      args: ["run"]
      timeout: 2m

    - name: secrets
      tool: gitleaks
      args: ["protect", "--staged"]
```

### 5.3 Example (JSON)

```json
{
  "hooks": {
    "pre-commit": [
      {
        "name": "gofmt",
        "tool": "go",
        "args": ["fmt", "./..."],
        "files": "\\.go$"
      }
    ]
  }
}
```

---

## 6. Functional Requirements

### 6.1 Hook Execution

* Support git hooks: `pre-commit`, `pre-push`, `commit-msg`
* Execute hooks based on config
* Exit non-zero on failure

### 6.2 Tool Management

* Download tools per OS & architecture
* Cache tools locally under `.hooks/cache/`
* Verify checksum (SHA256)

### 6.3 File Filtering

* Detect staged files using `git diff --cached`
* Run hooks only on matching files

### 6.4 Parallelism

* Hooks run in parallel by default
* Allow dependency ordering via `after`

---

## 7. Non-Functional Requirements

### Performance

* Hook startup < 100ms
* Parallel execution where possible

### Reliability

* Deterministic behavior across platforms
* Clear error messages

### Security

* No shell execution
* Only trusted binaries
* Version-pinned tools

---

## 8. CLI Interface

```bash
hookrunner install
hookrunner run pre-commit
hookrunner run pre-commit --all-files
hookrunner list
hookrunner doctor
```

---

## 9. CI Integration

Same runner used in CI:

```bash
.hooks/hookrunner run pre-commit --all-files
```

Ensures no duplication between local hooks and CI pipelines.

---

## 10. UX Requirements

* Colored output
* Execution time per hook
* Clear failure summary

Example:

```
✓ gofmt (120ms)
✗ golangci-lint (2.3s)
  → errcheck: unchecked error
```

---

## 11. Open Questions

* Should config support remote includes?
* How strict should checksum validation be?
* Should hooks support auto-fix by default?

---

## 12. Success Metrics

* < 5 min onboarding time
* 90% hook execution success rate
* Reduced CI failures due to linting

---

## 13. Future Enhancements

* Plugin SDK
* IDE integrations
* Hook performance profiling
* Remote cache support

---

## 14. Interview One-Liner

> "We designed a cross-platform hook system in Go using a declarative YAML/JSON config, shipping a single binary that manages tool installation, execution, and CI reuse to guarantee reproducibility and speed."
