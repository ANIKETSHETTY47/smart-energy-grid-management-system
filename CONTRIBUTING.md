# Contributing to Smart Energy Grid Management System

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone <your-fork-url>`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes: `git commit -m "Add feature: description"`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Create a Pull Request

## Development Setup

See the [README.md](README.md) for detailed setup instructions.

## Code Standards

### Go Style Guide

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Run `go vet` before committing
- Ensure all tests pass: `go test ./...`
- Write meaningful commit messages

### Commit Messages

Format: `<type>: <subject>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

Example: `feat: add real-time sensor monitoring dashboard`

### Testing

- Write unit tests for new functions
- Ensure test coverage remains above 70%
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example:
```go
func TestSensorValidation(t *testing.T) {
    tests := []struct {
        name    string
        sensor  Sensor
        wantErr bool
    }{
        {"valid sensor", Sensor{ID: "1", Name: "Test"}, false},
        {"empty ID", Sensor{Name: "Test"}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateSensor(tt.sensor)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateSensor() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Pull Request Process

1. **Update Documentation**: Update README.md if adding new features
2. **Add Tests**: Ensure new code has adequate test coverage
3. **Pass CI**: All GitHub Actions checks must pass
4. **Code Review**: Address reviewer comments promptly
5. **Squash Commits**: Squash commits before merging if requested

## Project Structure

When adding new code, follow the existing structure:

- `cmd/` - Application entry points
- `internal/cloud/` - AWS service integrations
- `internal/config/` - Configuration loading
- `internal/database/` - Database operations
- `internal/domain/` - Domain models and interfaces
- `internal/http/` - HTTP handlers and routing
- `internal/repository/` - Data access layer
- `internal/service/` - Business logic
- `scripts/` - SQL and utility scripts
- `web/` - Frontend applications

## Reporting Issues

When reporting issues, please include:

1. **Description**: Clear description of the issue
2. **Steps to Reproduce**: Detailed steps
3. **Expected Behavior**: What should happen
4. **Actual Behavior**: What actually happens
5. **Environment**: OS, Go version, etc.
6. **Logs**: Relevant error logs or stack traces

## Feature Requests

For feature requests, please:

1. Check existing issues first
2. Clearly describe the feature
3. Explain the use case
4. Provide examples if possible
5. Consider implementation complexity

## Questions?

Feel free to open a discussion or reach out to the maintainers.

Thank you for contributing! 🎉
