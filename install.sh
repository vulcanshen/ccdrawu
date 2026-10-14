#!/bin/sh
# ccdrawu installer for macOS / Linux (Windows: install.ps1).
# Usage: curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.sh | sh

set -e

REPO="vulcanshen/ccdrawu"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  *) echo "Error: unsupported OS: $OS. On Windows, run in PowerShell:"
     echo "  irm https://raw.githubusercontent.com/$REPO/main/install.ps1 | iex"
     exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
echo "Fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')
echo "Latest version: $VERSION"

# Install dir
if [ "$(id -u)" = "0" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
fi

FILENAME="ccdrawu_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/v${VERSION}/$FILENAME"

# Download
TMPDIR=$(mktemp -d)
echo "Downloading $FILENAME..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/$FILENAME"

# Extract
echo "Extracting..."
tar xzf "$TMPDIR/$FILENAME" -C "$TMPDIR"

# Install
mkdir -p "$INSTALL_DIR"
cp "$TMPDIR/ccdrawu" "$INSTALL_DIR/ccdrawu"
chmod +x "$INSTALL_DIR/ccdrawu"
rm -rf "$TMPDIR"

echo ""
echo "ccdrawu $VERSION installed to $INSTALL_DIR"

# Check if install dir is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo ""
    echo "WARNING: $INSTALL_DIR is not in your PATH. Add it by running:"
    echo ""
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.$(basename "$SHELL")rc && source ~/.$(basename "$SHELL")rc"
    ;;
esac

echo ""
echo "NOTE: ccdrawu draws with Nerd Font glyphs. Use a Nerd Font terminal"
echo "      profile (https://www.nerdfonts.com) or they render as boxes."
echo ""
echo "The drawer itself is a Claude Code plugin. In Claude Code, run:"
echo "  /plugin install ccdrawu --marketplace $REPO"
echo "Then run 'ccdrawu' to pick what it shows."
