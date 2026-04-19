#!/usr/bin/env bash

set -euo pipefail

REPO="ThinhDangDev/go-template"
BINARY_NAME="go-template"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$arch" in
  x86_64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

case "$os" in
  linux|darwin) ;;
  *)
    echo "unsupported operating system: $os" >&2
    exit 1
    ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

if ! command -v tar >/dev/null 2>&1; then
  echo "tar is required" >&2
  exit 1
fi

version="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | cut -d '"' -f 4)"
if [[ -z "$version" ]]; then
  echo "failed to determine latest version" >&2
  exit 1
fi

archive="${BINARY_NAME}_${version#v}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$version/$archive"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

curl -fsSL "$url" -o "$tmpdir/$archive"
tar -xzf "$tmpdir/$archive" -C "$tmpdir"

if [[ ! -w "$INSTALL_DIR" ]]; then
  echo "install dir $INSTALL_DIR is not writable; set INSTALL_DIR to a writable path or rerun with sudo" >&2
  exit 1
fi

install -m 0755 "$tmpdir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
echo "installed $BINARY_NAME $version to $INSTALL_DIR/$BINARY_NAME"
