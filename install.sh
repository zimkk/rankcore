#!/bin/sh
set -e

echo "Installing RankCore..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

echo "Detected OS: $OS, Architecture: $ARCH"
echo "Downloading binary..."
# curl -fsSL https://rankcore.dev/releases/latest/rankcore-${OS}-${ARCH} -o rankcore

echo "Setting up..."
# chmod +x rankcore
# mv rankcore /usr/local/bin/rankcore
# rankcore setup

echo "RankCore installed successfully. Run /rank in your agent."
