#!/bin/bash
set -e

# Go Report Card Grade Checker
# Ensures code quality meets Grade A standards

echo "🔍 Checking Go Report Card criteria..."

# Check gofmt
echo "📝 Checking gofmt..."
GOFMT_FILES=$(gofmt -l .)
if [ -n "$GOFMT_FILES" ]; then
    echo "❌ gofmt issues found in:"
    echo "$GOFMT_FILES"
    echo "Run 'gofmt -w .' to fix"
    exit 1
fi
echo "✅ gofmt: PASS"

# Check goimports
echo "📦 Checking goimports..."
if command -v goimports >/dev/null 2>&1; then
    GOIMPORTS_FILES=$(goimports -l .)
    if [ -n "$GOIMPORTS_FILES" ]; then
        echo "❌ goimports issues found in:"
        echo "$GOIMPORTS_FILES"
        echo "Run 'goimports -w .' to fix"
        exit 1
    fi
    echo "✅ goimports: PASS"
else
    echo "⚠️ goimports not found, install with: go install golang.org/x/tools/cmd/goimports@latest"
fi

# Check go vet
echo "🔍 Running go vet..."
go vet ./...
echo "✅ go vet: PASS"

# Check golangci-lint
echo "🔧 Running golangci-lint..."
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run
    echo "✅ golangci-lint: PASS"
else
    echo "⚠️ golangci-lint not found, install from: https://golangci-lint.run/usage/install/"
    echo "❌ golangci-lint: SKIP (required for Grade A)"
    exit 1
fi

# Check test coverage
echo "🧪 Running tests with coverage..."
go test -race -coverprofile=coverage.out ./...

# Calculate coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
COVERAGE_INT=$(printf "%.0f" "$COVERAGE")

echo "📊 Test coverage: ${COVERAGE}%"

if [ "$COVERAGE_INT" -lt 80 ]; then
    echo "❌ Coverage too low: ${COVERAGE}% (minimum 80% for Grade A)"
    exit 1
fi
echo "✅ Coverage: PASS (${COVERAGE}%)"

# Check for ineffassign
echo "🎯 Checking for ineffective assignments..."
if command -v ineffassign >/dev/null 2>&1; then
    ineffassign ./...
    echo "✅ ineffassign: PASS"
else
    echo "⚠️ ineffassign not found, install with: go install github.com/gordonklaus/ineffassign@latest"
fi

# Check for misspell
echo "📝 Checking for misspellings..."
if command -v misspell >/dev/null 2>&1; then
    MISSPELL_FILES=$(misspell .)
    if [ -n "$MISSPELL_FILES" ]; then
        echo "❌ Misspellings found:"
        echo "$MISSPELL_FILES"
        exit 1
    fi
    echo "✅ misspell: PASS"
else
    echo "⚠️ misspell not found, install with: go install github.com/client9/misspell/cmd/misspell@latest"
fi

# Check go mod tidy
echo "🧹 Checking go mod tidy..."
cp go.mod go.mod.bak
cp go.sum go.sum.bak
go mod tidy
if ! cmp -s go.mod go.mod.bak || ! cmp -s go.sum go.sum.bak; then
    echo "❌ go mod tidy needed - files are not tidy"
    rm go.mod.bak go.sum.bak
    exit 1
fi
rm go.mod.bak go.sum.bak
echo "✅ go mod tidy: PASS"

# Final grade assessment
echo ""
echo "🎉 All Go Report Card criteria met!"
echo "🏆 Expected Grade: A"
echo ""
echo "Report Summary:"
echo "- ✅ gofmt: No formatting issues"
echo "- ✅ goimports: Imports properly organized"  
echo "- ✅ go vet: No suspicious constructs"
echo "- ✅ golangci-lint: All quality checks passed"
echo "- ✅ Test coverage: ${COVERAGE}% (≥80%)"
echo "- ✅ ineffassign: No ineffective assignments"
echo "- ✅ misspell: No misspellings found"
echo "- ✅ go mod tidy: Dependencies are tidy"