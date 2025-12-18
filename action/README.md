# HookRunner GitHub Action

Run HookRunner pre-commit hooks in your CI/CD pipeline.

## Usage

```yaml
name: Hooks
on: [push, pull_request]

jobs:
  hooks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run HookRunner
        uses: ashavijit/HookRunner@v1
        with:
          hook-type: pre-commit
          all-files: true
```

## Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `hook-type` | Hook to run (pre-commit, pre-push) | `pre-commit` |
| `all-files` | Run on all files | `false` |
| `fix` | Run in fix mode | `false` |
| `version` | HookRunner version | `latest` |

## Examples

### Basic Pre-commit

```yaml
- uses: ashavijit/HookRunner@v1
```

### All Files with Fix Mode

```yaml
- uses: ashavijit/HookRunner@v1
  with:
    all-files: true
    fix: true
```

### Specific Version

```yaml
- uses: ashavijit/HookRunner@v1
  with:
    version: v0.19.0
```

### Multiple Hook Types

```yaml
jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ashavijit/HookRunner@v1
        with:
          hook-type: pre-commit
          all-files: true

  pre-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ashavijit/HookRunner@v1
        with:
          hook-type: pre-push
```
