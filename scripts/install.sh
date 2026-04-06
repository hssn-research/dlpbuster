#!/usr/bin/env bash
# dlpbuster installer script
set -euo pipefail

REPO="hssn-research/dlpbuster"
BINARY="dlpbuster"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Could not determine latest release"
  exit 1
fi

EXT="tar.gz"
if [ "$OS" = "windows" ]; then EXT="zip"; fi
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}_${LATEST#v}_${OS}_${ARCH}.${EXT}"

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

echo "Downloading dlpbuster ${LATEST}..."
curl -fsSL "$URL" -o "${TMP}/archive.${EXT}"

if [ "$EXT" = "tar.gz" ]; then
  tar -C "$TMP" -xzf "${TMP}/archive.${EXT}"
else
  unzip -q "${TMP}/archive.${EXT}" -d "$TMP"
fi

install -m 755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
echo "dlpbuster installed to ${INSTALL_DIR}/${BINARY}"
echo "Run: dlpbuster --version"
