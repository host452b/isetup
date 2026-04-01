#!/bin/sh
set -e

REPO="host452b/isetup"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect Arch
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
echo "Detecting latest version..."

# Method 1: GitHub API (may hit rate limit without auth)
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | head -1 | cut -d'"' -f4 || true)

# Method 2: Follow redirect from /releases/latest to get tag from URL
if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSI "https://github.com/${REPO}/releases/latest" 2>/dev/null | grep -i '^location:' | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true)
fi

# Method 3: Hardcoded fallback
if [ -z "$VERSION" ]; then
  VERSION="v0.3.0"
  echo "Could not auto-detect version, using fallback: ${VERSION}"
fi
VERSION_NUM=$(echo "$VERSION" | tr -d 'v')

URL="https://github.com/${REPO}/releases/download/${VERSION}/isetup_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
echo "Downloading isetup ${VERSION} for ${OS}/${ARCH}..."
echo "  ${URL}"

# Download and extract
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "${TMP}/isetup.tar.gz"
tar xzf "${TMP}/isetup.tar.gz" -C "$TMP"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/isetup" "${INSTALL_DIR}/isetup"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo mv "${TMP}/isetup" "${INSTALL_DIR}/isetup"
fi

chmod +x "${INSTALL_DIR}/isetup"
echo "Installed isetup ${VERSION} to ${INSTALL_DIR}/isetup"
echo ""
echo "Run: isetup install"
