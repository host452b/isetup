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
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
if [ -z "$VERSION" ]; then
  echo "Failed to detect latest version"
  exit 1
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
