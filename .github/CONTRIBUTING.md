# Contributing to Knot

Thank you for your interest in contributing to Knot! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct (see CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

1. **Check existing issues** first to avoid duplicates
2. **Use the bug report template** when creating new issues
3. **Provide detailed information** including:
   - Knot version (`knot --version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Command output and logs

### Suggesting Features

1. **Check existing feature requests** to avoid duplicates
2. **Use the feature request template**
3. **Describe your use case** clearly
4. **Provide example commands** showing how the feature would work

### Contributing Code

#### Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of CLI development

#### Development Setup

1. **Fork the repository**
   ```bash
   gh repo fork denkhaus/knot --clone
   cd knot
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   go build -o knot cmd/knot/main.go
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

#### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes**
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all tests pass

3. **Test your changes**
   ```bash
   # Run all tests
   go test ./...
   
   # Test specific functionality
   go test ./internal/commands/task/
   
   # Run integration tests
   go test ./internal/ -tags=integration
   
   # Test the CLI manually
   ./knot --help
   ./knot project create --name "Test Project"
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   # or
   git commit -m "fix: resolve issue with task deletion"
   ```

   Use conventional commit format:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test additions/changes
   - `refactor:` for code refactoring
   - `chore:` for maintenance tasks

5. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   gh pr create --title "Add new feature" --body "Description of changes"
   ```

#### Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Use the existing error handling patterns

#### Testing

- Write unit tests for new functions
- Add integration tests for new commands
- Test edge cases and error conditions
- Ensure tests are deterministic and fast

#### Documentation

- Update README.md for new features
- Add command examples
- Update help text and usage information
- Document configuration options

### Project Structure

```
knot/
├── cmd/knot/           # Main CLI entry point
├── internal/
│   ├── app/           # Application setup and configuration
│   ├── commands/      # CLI command implementations
│   │   ├── project/   # Project management commands
│   │   ├── task/      # Task management commands
│   │   ├── dependency/# Dependency management commands
│   │   └── ...
│   ├── manager/       # Business logic layer
│   ├── repository/    # Data access layer
│   ├── types/         # Type definitions
│   ├── validation/    # Input validation
│   └── ...
├── .github/           # GitHub workflows and templates
└── docs/              # Additional documentation
```

### Adding New Commands

1. **Create command file** in appropriate package under `internal/commands/`
2. **Implement the command** following existing patterns
3. **Add to parent command** in the relevant commands.go file
4. **Write tests** for the new command
5. **Update documentation** with examples

Example command structure:
```go
func NewExampleCommand(appCtx *shared.AppContext) *cli.Command {
    return &cli.Command{
        Name:   "example",
        Usage:  "Example command description",
        Action: exampleAction(appCtx),
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:     "required-flag",
                Usage:    "Description of required flag",
                Required: true,
            },
        },
    }
}

func exampleAction(appCtx *shared.AppContext) cli.ActionFunc {
    return func(c *cli.Context) error {
        // Implementation here
        return nil
    }
}
```

### Adding New Templates

1. **Create template YAML** in `internal/templates/`
2. **Add to embedded files** in `embedded.go`
3. **Test template application**
4. **Update documentation**

### Release Process

Releases are automated through GitHub Actions:

1. **Create a tag** following semantic versioning
   ```bash
   git tag v1.1.0
   git push origin v1.1.0
   ```

2. **GitHub Actions will**:
   - Run all tests
   - Build binaries for multiple platforms
   - Create GitHub release
   - Publish to package managers

### Getting Help

- **GitHub Discussions**: Ask questions and discuss ideas
- **Issues**: Report bugs and request features
- **Discord/Slack**: Real-time chat (if available)

### Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- README acknowledgments

## Development Tips

### Useful Commands

```bash
# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o knot-linux ./cmd/knot
GOOS=windows GOARCH=amd64 go build -o knot-windows.exe ./cmd/knot
GOOS=darwin GOARCH=amd64 go build -o knot-darwin ./cmd/knot

# Run linter
golangci-lint run

# Generate mocks (if using)
go generate ./...
```

### Debugging

- Use `KNOT_LOG_LEVEL=debug` for verbose logging
- Add temporary debug prints during development
- Use Go's built-in debugger or IDE debugging tools

### Performance

- Profile memory usage with `go test -memprofile`
- Profile CPU usage with `go test -cpuprofile`
- Benchmark critical paths with `go test -bench`

Thank you for contributing to Knot! 🎉