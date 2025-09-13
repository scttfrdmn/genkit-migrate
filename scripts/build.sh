#!/bin/bash
set -e

# Build script for genkit-migrate

VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS="-X github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate/cmd.version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate/cmd.commit=$COMMIT"
LDFLAGS="$LDFLAGS -X github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate/cmd.buildTime=$BUILD_TIME"

echo "Building genkit-migrate..."
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Build Time: $BUILD_TIME"

# Clean previous builds
rm -rf build/
mkdir -p build/

# Build for multiple platforms
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64"

for PLATFORM in $PLATFORMS; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    OUTPUT_NAME="genkit-migrate-$VERSION-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="$OUTPUT_NAME.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="$LDFLAGS" \
        -o "build/$OUTPUT_NAME" \
        ./cmd/genkit-migrate
done

echo "Build complete! Binaries are in the build/ directory."
ls -la build/