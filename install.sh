#!/usr/bin/env bash

set -e

REPO="sagarmaheshwary/reqlog-ui"
INSTALL_PATH="/usr/local/bin/reqlog-ui"
REQUIRED="v0.7.1"

# Check reqlog binary
if ! command -v reqlog >/dev/null 2>&1; then
  echo "reqlog is required for reqlog-ui."
  echo
  echo "Install compatible version:"
  echo "curl -sSL https://raw.githubusercontent.com/sagarmaheshwary/reqlog/master/install.sh | bash -s $REQUIRED"
  exit 1
fi

REQLOG_VERSION=$(reqlog --version | awk '{print $3}')

echo "Detected reqlog version: $REQLOG_VERSION"
echo "reqlog-ui requires reqlog >= $REQUIRED"

if [ "$(printf '%s\n' "$REQUIRED" "$REQLOG_VERSION" | sort -V | head -n1)" != "$REQUIRED" ]; then
  echo
  echo "Incompatible reqlog version detected."
  echo "Please upgrade reqlog:"
  echo "curl -sSL https://raw.githubusercontent.com/sagarmaheshwary/reqlog/master/install.sh | bash -s $REQUIRED"
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