#!/bin/bash
set -e

# Release script for genkit-migrate

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid version format. Use vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

echo "Preparing release $VERSION..."

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Warning: You are not on the main branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: There are uncommitted changes"
    exit 1
fi

# Run tests
echo "Running tests..."
./scripts/test.sh

# Build binaries
echo "Building binaries..."
VERSION=$VERSION ./scripts/build.sh

# Create changelog entry
CHANGELOG_FILE="CHANGELOG.md"
if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "# Changelog" > "$CHANGELOG_FILE"
fi

# Update version in root.go
sed -i.bak "s/const version = .*/const version = \"$VERSION\"/" cmd/genkit-migrate/cmd/root.go
rm cmd/genkit-migrate/cmd/root.go.bak

# Commit version update
git add cmd/genkit-migrate/cmd/root.go
git commit -m "Bump version to $VERSION"

# Create tag
git tag -a "$VERSION" -m "Release $VERSION"

echo "Release $VERSION prepared!"
echo "Next steps:"
echo "1. Push the changes: git push origin main"
echo "2. Push the tag: git push origin $VERSION"
echo "3. Create a GitHub release with the binaries in build/"
echo "4. Upload the binaries to the GitHub release"