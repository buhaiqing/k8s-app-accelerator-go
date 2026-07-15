#!/usr/bin/env bash
#
# install.sh — one-shot install / update for the k8s-gen CLI.
#
# Cross-platform: works on macOS, Linux and Windows (Git Bash / WSL / MSYS).
# Automatically detects OS + architecture, picks the matching GitHub Release
# asset, and installs with zero manual steps. Re-running updates in place.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/buhaiqing/k8s-app-accelerator-go/main/cmd/k8s-gen/install.sh | bash
#   # or after cloning the repo:
#   bash cmd/k8s-gen/install.sh
#   # install a specific version:
#   VERSION=1.0.0 curl -fsSL ... | bash
#
# Env overrides:
#   VERSION       pin a version (default: latest release)
#   INSTALL_DIR   install path (default: ~/.local/bin on *nix, %LOCALAPPDATA%\Programs\k8s-gen on Windows)
#
set -euo pipefail

REPO="buhaiqing/k8s-app-accelerator-go"
BINARY="k8s-gen"

err() { echo "error: $*" >&2; exit 1; }

# --- detect OS / arch -------------------------------------------------------
RAW_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *) err "unsupported architecture: $ARCH" ;;
esac
case "$RAW_OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  mingw32* | mingw64* | msys* | cygwin* | uwin*) OS="windows" ;;
  *) err "unsupported OS: $RAW_OS (only linux/darwin/windows supported)" ;;
esac
BIN_NAME="$BINARY"
[ "$OS" = "windows" ] && BIN_NAME="${BINARY}.exe"

# --- default install dir (user-writable, no sudo) ---------------------------
if [ -z "${INSTALL_DIR:-}" ]; then
  if [ "$OS" = "windows" ]; then
    INSTALL_DIR="${LOCALAPPDATA:-$HOME/AppData/Local}/Programs/k8s-gen"
  else
    INSTALL_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
  fi
fi
VERSION_FILE="${INSTALL_DIR}/${BINARY}.version"

# --- resolve version --------------------------------------------------------
command -v curl >/dev/null 2>&1 || err "curl is required"

if [ -n "${VERSION:-}" ]; then
  VERSION="${VERSION#v}"
  LATEST="v${VERSION}"
else
  echo "resolving latest $BINARY release for ${OS}/${ARCH} ..."
  LATEST="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -n "$LATEST" ] || err "could not determine latest release tag"
  VERSION="${LATEST#v}"
fi

# --- skip if already up to date ---------------------------------------------
if [ -f "$VERSION_FILE" ]; then
  CURRENT="$(cat "$VERSION_FILE" 2>/dev/null || true)"
  if [ "$CURRENT" = "$VERSION" ]; then
    echo "✅ $BIN_NAME $VERSION is already installed at $INSTALL_DIR/$BIN_NAME (nothing to do)"
    "$INSTALL_DIR/$BIN_NAME" --help >/dev/null 2>&1 || true
    exit 0
  fi
fi

# --- find asset for this platform ------------------------------------------
ASSET="${BINARY}_${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "downloading $URL"
if ! curl -fsSL "$URL" -o "$TMP/$ASSET"; then
  ASSET="${BINARY}_${OS}-${ARCH}.zip"
  URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
  curl -fsSL "$URL" -o "$TMP/$ASSET" || err "download failed: $URL"
fi

# --- extract ----------------------------------------------------------------
if [[ "$ASSET" == *.zip ]]; then
  if command -v unzip >/dev/null 2>&1; then
    unzip -o "$TMP/$ASSET" -d "$TMP" >/dev/null
  else
    err "need unzip to extract the .zip asset"
  fi
else
  tar -xzf "$TMP/$ASSET" -C "$TMP"
fi

SRC="$(find "$TMP" -type f \( -name "$BINARY" -o -name "${BINARY}.exe" \) | head -1)"
[ -n "$SRC" ] || err "binary $BIN_NAME not found in archive"

# --- install ----------------------------------------------------------------
mkdir -p "$INSTALL_DIR"
echo "installing to $INSTALL_DIR/$BIN_NAME"
if [ "$OS" = "windows" ]; then
  install -m 0755 "$SRC" "$INSTALL_DIR/$BIN_NAME" 2>/dev/null \
    || cp "$SRC" "$INSTALL_DIR/$BIN_NAME" \
    || err "$INSTALL_DIR is not writable; run from an elevated shell or set INSTALL_DIR to a writable path"
elif [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$SRC" "$INSTALL_DIR/$BIN_NAME"
else
  sudo install -m 0755 "$SRC" "$INSTALL_DIR/$BIN_NAME"
fi
echo "$VERSION" > "$VERSION_FILE"

echo "✅ $BIN_NAME $VERSION installed at $INSTALL_DIR/$BIN_NAME"
"$INSTALL_DIR/$BIN_NAME" --help | head -5
