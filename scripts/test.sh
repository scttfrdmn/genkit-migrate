#!/bin/bash
set -e

echo "Running tests for genkit-migrate..."

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"

# Show coverage summary
go tool cover -func=coverage.out | tail -1

# Run linting if available
if command -v golangci-lint &> /dev/null; then
    echo "Running linter..."
    golangci-lint run
else
    echo "golangci-lint not found, skipping linting"
fi

echo "All tests passed!"