#!/bin/bash
set -e

REPO="ashavijit/hookrunner"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -sSL "https://raw.githubusercontent.com/$REPO/master/VERSION" 2>/dev/null || echo "v0.1.1")
  VERSION=$(echo "$VERSION" | tr -d '\n\r')
fi

BINARY="hookrunner-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY"

echo ""
echo "Installing hookrunner $VERSION..."
echo "  OS: $OS"
echo "  Arch: $ARCH"
echo ""

TMP_FILE=$(mktemp)
curl -sSL "$URL" -o "$TMP_FILE" || { echo "Download failed: $URL"; exit 1; }
chmod +x "$TMP_FILE"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "$INSTALL_DIR/hookrunner"
else
  sudo mv "$TMP_FILE" "$INSTALL_DIR/hookrunner"
fi

echo "Installed to $INSTALL_DIR/hookrunner"
hookrunner version 2>/dev/null || true
