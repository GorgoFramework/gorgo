# Contributing to Gorgo Framework

We love your input! We want to make contributing to Gorgo Framework as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Pull Requests

Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `develop`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Git

### Setting up the development environment

1. **Clone the repository**
   ```bash
   git clone https://github.com/GorgoFramework/gorgo.git
   cd gorgo
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests to ensure everything works**
   ```bash
   go test ./...
   ```

4. **Run examples**
   ```bash
   go run examples/basic_example/main.go
   go run examples/advanced_plugins_example/main.go
   ```

## Coding Standards

### Go Style Guide

We follow the standard Go style guidelines:

- Use `gofmt` to format your code
- Use `golint` and `go vet` to catch common mistakes
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Write clear and concise comments

### Code Formatting

Before submitting a pull request, please ensure your code is properly formatted:

```bash
# Format code
go fmt ./...

# Run linting
golint ./...

# Run vet
go vet ./...
```

### Naming Conventions

- **Packages**: Use lowercase, single words when possible
- **Functions**: Use camelCase, start with uppercase for exported functions
- **Variables**: Use camelCase, start with lowercase
- **Constants**: Use PascalCase for exported constants, camelCase for unexported
- **Files**: Use lowercase with underscores for separation (e.g., `plugin_manager.go`)

### Comments

- All exported functions, types, and variables must have comments
- Comments should start with the name of the function/type/variable
- Use complete sentences in comments
- Keep comments up to date with code changes

Example:
```go
// NewApplication creates a new Gorgo application instance with default configuration.
// It initializes the plugin manager, middleware chain, and event bus.
func NewApplication() *Application {
    // implementation
}
```

## Testing

### Writing Tests

- Write tests for all new functionality
- Use table-driven tests when appropriate
- Test both happy paths and error cases
- Use meaningful test names that describe what is being tested

Example test structure:
```go
func TestContextQuery(t *testing.T) {
    tests := []struct {
        name     string
        query    string
        key      string
        expected string
    }{
        {
            name:     "existing key",
            query:    "key=value&foo=bar",
            key:      "key",
            expected: "value",
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestContextQuery ./pkg/gorgo
```

### Benchmarks

When adding performance-critical code, include benchmarks:

```go
func BenchmarkContextQuery(b *testing.B) {
    ctx := setupTestContext()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        ctx.Query("key")
    }
}
```

## Plugin Development

### Creating New Plugins

When contributing new plugins:

1. **Follow the plugin interface**
   ```go
   type Plugin interface {
       GetMetadata() PluginMetadata
       Initialize(app Application) error
       GetState() PluginState
   }
   ```

2. **Implement required interfaces based on functionality**:
   - `ConfigurablePlugin` for configuration support
   - `ServiceProvider` for providing services
   - `EventSubscriber` for event handling
   - `LifecycleHooks` for lifecycle management
   - `HotReloadable` for hot reload support
   - `MiddlewareProvider` for middleware

3. **Plugin structure**:
   ```
   plugins/
   ├── your_plugin/
   │   ├── plugin.go
   │   ├── config.go
   │   ├── README.md
   │   └── examples/
   │       └── example.go
   ```

4. **Plugin documentation**:
   - Include a README.md with usage examples
   - Document all configuration options
   - Provide working examples

### Plugin Guidelines

- Keep plugins focused on a single responsibility
- Ensure plugins are thread-safe
- Handle errors gracefully
- Provide meaningful error messages
- Support graceful shutdown
- Include comprehensive tests

## Documentation

### Code Documentation

- Document all exported functions, types, and variables
- Include usage examples in comments where helpful
- Keep documentation up to date with code changes

### README Updates

When adding new features:

- Update the main README.md if necessary
- Add examples to demonstrate new functionality
- Update feature lists and compatibility information

### API Documentation

- Use godoc comments for API documentation
- Include code examples in documentation
- Document error conditions and return values

## Commit Messages

Use clear and meaningful commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:
```
Add SQL plugin with connection pooling

- Implement database connection management
- Add transaction middleware support
- Include connection health checks
- Add comprehensive test coverage

Closes #123
```

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

- **Description**: Clear description of the bug
- **Steps to reproduce**: Detailed steps to reproduce the issue
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: Go version, OS, Gorgo version
- **Code sample**: Minimal code that reproduces the issue

### Feature Requests

When requesting features:

- **Description**: Clear description of the proposed feature
- **Use case**: Why this feature would be useful
- **Examples**: How the feature would be used
- **Alternatives**: Any alternative solutions considered

## Code Review Process

1. **All submissions require review**: All pull requests must be reviewed by at least one maintainer
2. **Automated checks**: PRs must pass all automated tests and linting
3. **Documentation**: New features must include documentation
4. **Breaking changes**: Must be clearly marked and justified

## Release Process

- We use semantic versioning (SemVer)
- Breaking changes increment the major version
- New features increment the minor version
- Bug fixes increment the patch version

## Community

- Be respectful and inclusive
- Help others learn and grow
- Share knowledge and best practices
- Collaborate constructively

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

## Getting Help

If you need help:

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Reach out to maintainers

---

Thank you for contributing to Gorgo Framework! 🚀 