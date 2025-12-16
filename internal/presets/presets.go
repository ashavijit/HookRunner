package presets

type Preset struct {
	Name        string
	Description string
	Config      string
}

var Languages = map[string]Preset{
	"go": {
		Name:        "Go",
		Description: "Go language hooks (gofmt, govet, golangci-lint)",
		Config: `hooks:
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

  pre-push:
    - name: test
      tool: go
      args: ["test", "-race", "./..."]
      timeout: 5m
`,
	},

	"nodejs": {
		Name:        "Node.js",
		Description: "Node.js hooks (eslint, prettier)",
		Config: `hooks:
  pre-commit:
    - name: eslint
      tool: npx
      args: ["eslint", "--fix", "."]
      files: "\\.(js|jsx|ts|tsx)$"
      fix_args: ["eslint", "--fix", "."]

    - name: prettier
      tool: npx
      args: ["prettier", "--check", "."]
      files: "\\.(js|jsx|ts|tsx|json|css|md)$"
      fix_args: ["prettier", "--write", "."]
      after: eslint

  pre-push:
    - name: test
      tool: npm
      args: ["test"]
      timeout: 5m
`,
	},

	"python": {
		Name:        "Python",
		Description: "Python hooks (black, flake8, mypy, pytest)",
		Config: `hooks:
  pre-commit:
    - name: black
      tool: black
      args: ["--check", "."]
      files: "\\.py$"
      fix_args: ["."]

    - name: flake8
      tool: flake8
      args: ["."]
      files: "\\.py$"
      after: black

    - name: mypy
      tool: mypy
      args: ["."]
      files: "\\.py$"
      after: flake8

  pre-push:
    - name: pytest
      tool: pytest
      args: ["-v"]
      timeout: 5m
`,
	},

	"java": {
		Name:        "Java",
		Description: "Java hooks (checkstyle, spotless, maven test)",
		Config: `hooks:
  pre-commit:
    - name: checkstyle
      tool: mvn
      args: ["checkstyle:check"]
      files: "\\.java$"

    - name: spotless
      tool: mvn
      args: ["spotless:check"]
      files: "\\.java$"
      fix_args: ["spotless:apply"]

  pre-push:
    - name: test
      tool: mvn
      args: ["test"]
      timeout: 10m
`,
	},

	"ruby": {
		Name:        "Ruby",
		Description: "Ruby hooks (rubocop, rspec)",
		Config: `hooks:
  pre-commit:
    - name: rubocop
      tool: rubocop
      args: ["--parallel"]
      files: "\\.rb$"
      fix_args: ["--autocorrect"]

  pre-push:
    - name: rspec
      tool: rspec
      args: ["--format", "progress"]
      timeout: 5m
`,
	},

	"rust": {
		Name:        "Rust",
		Description: "Rust hooks (cargo fmt, clippy, test)",
		Config: `hooks:
  pre-commit:
    - name: cargo-fmt
      tool: cargo
      args: ["fmt", "--", "--check"]
      files: "\\.rs$"
      fix_args: ["fmt"]

    - name: clippy
      tool: cargo
      args: ["clippy", "--", "-D", "warnings"]
      files: "\\.rs$"
      after: cargo-fmt

  pre-push:
    - name: test
      tool: cargo
      args: ["test"]
      timeout: 5m
`,
	},
}

func List() []Preset {
	presets := make([]Preset, 0, len(Languages))
	for _, p := range Languages {
		presets = append(presets, p)
	}
	return presets
}

func Get(name string) (Preset, bool) {
	p, ok := Languages[name]
	return p, ok
}

func AvailableLanguages() []string {
	return []string{"go", "nodejs", "python", "java", "ruby", "rust"}
}
