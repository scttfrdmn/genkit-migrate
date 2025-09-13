#!/bin/bash
set -e

# Install script for genkit-migrate

INSTALL_DIR=${INSTALL_DIR:-"/usr/local/bin"}
REPO="genkit-migrate/genkit-migrate"
BINARY_NAME="genkit-migrate"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case $OS in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

echo "Installing genkit-migrate for $OS/$ARCH..."

# Get the latest release
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to get latest release"
    exit 1
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/genkit-migrate-$LATEST_RELEASE-$OS-$ARCH"

if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="$DOWNLOAD_URL.exe"
    BINARY_NAME="$BINARY_NAME.exe"
fi

echo "Downloading $DOWNLOAD_URL..."

# Download the binary
curl -L -o "/tmp/$BINARY_NAME" "$DOWNLOAD_URL"

# Make it executable
chmod +x "/tmp/$BINARY_NAME"

# Move to install directory
sudo mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo "genkit-migrate installed successfully to $INSTALL_DIR/$BINARY_NAME"
echo "Run 'genkit-migrate --help' to get started!"