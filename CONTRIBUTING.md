# Contributing to HookRunner

Thank you for your interest in contributing to HookRunner!

## Development Setup

```bash
git clone https://github.com/ashavijit/hookrunner.git
cd hookrunner
go mod download
go build ./...
```

## Running Tests

```bash
go test ./... -v
go test ./... -cover
```

## Code Style

- Run `go fmt` before committing
- Run `go vet` to check for issues
- Follow standard Go conventions

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Commit Message Format

We follow conventional commits:

```
feat: add new feature
fix: fix a bug
docs: update documentation
test: add tests
refactor: refactor code
chore: maintenance tasks
```

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
