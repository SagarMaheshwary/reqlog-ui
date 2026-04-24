#!/usr/bin/env bash

set -e

REPO="sagarmaheshwary/reqlog-ui"
INSTALL_PATH="/usr/local/bin/reqlog-ui"

# Check reqlog binary
if ! command -v reqlog >/dev/null 2>&1; then
  echo "reqlog is required for reqlog-ui."
  echo "Install it using:"
  echo "curl -sSL https://raw.githubusercontent.com/sagarmaheshwary/reqlog/main/install.sh | bash"
  exit 1
fi

echo "Installing reqlog-ui..."

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux) PLATFORM="linux" ;;
  Darwin) PLATFORM="darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Normalize architecture
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

BINARY="reqlog-ui-${PLATFORM}-${ARCH}"
TAR_FILE="${BINARY}.tar.gz"

LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)

URL="https://github.com/$REPO/releases/download/$LATEST/${TAR_FILE}"

echo "Downloading $BINARY..."

curl -L "$URL" -o "$TAR_FILE"
tar -xzf "$TAR_FILE"

chmod +x "$BINARY"
sudo mv "$BINARY" "$INSTALL_PATH"

rm "$TAR_FILE"

echo "Installed reqlog-ui at $INSTALL_PATH"