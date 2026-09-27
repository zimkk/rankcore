#!/bin/sh
# RankCore installer for macOS / Linux / WSL.
# Downloads the latest precompiled release from GitHub, verifies its
# SHA-256 checksum, installs it to a user-writable directory (no sudo),
# and registers the /rank skill with detected coding agents.
set -eu

REPO="zimkk/rankcore"
INSTALL_DIR="${RANKCORE_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${RANKCORE_VERSION:-latest}"

info() { printf '%s\n' "$*"; }
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux|darwin) ;;
    msys*|mingw*|cygwin*)
        fail "On Windows use PowerShell instead: irm https://raw.githubusercontent.com/${REPO}/main/install.ps1 | iex" ;;
    *)
        fail "Unsupported operating system: $OS" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) fail "Unsupported architecture: $ARCH" ;;
esac

ASSET="rankcore-${OS}-${ARCH}"
if [ "$VERSION" = "latest" ]; then
    BASE_URL="https://github.com/${REPO}/releases/latest/download"
else
    BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
fi

info "Installing RankCore (${OS}/${ARCH})..."

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

info "Downloading ${ASSET} from GitHub Releases..."
if ! curl -fsSL "${BASE_URL}/${ASSET}" -o "${TMP_DIR}/${ASSET}"; then
    fail "Download failed. A precompiled release for ${OS}/${ARCH} may not be published yet.
Build from source instead (requires Go 1.23+):
  go install github.com/${REPO}/cmd/rankcore@latest"
fi
curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt" \
    || fail "Could not download checksums.txt — release assets may be incomplete."

info "Verifying SHA-256 checksum..."
EXPECTED="$(awk -v name="$ASSET" '$2 == name {print $1}' "${TMP_DIR}/checksums.txt")"
[ -n "$EXPECTED" ] || fail "No checksum entry found for ${ASSET}."
if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL="$(sha256sum "${TMP_DIR}/${ASSET}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL="$(shasum -a 256 "${TMP_DIR}/${ASSET}" | awk '{print $1}')"
else
    fail "Neither sha256sum nor shasum is available; cannot verify the download."
fi
[ "$EXPECTED" = "$ACTUAL" ] || fail "Checksum mismatch for ${ASSET}. Aborting."

mkdir -p "$INSTALL_DIR"
mv "${TMP_DIR}/${ASSET}" "${INSTALL_DIR}/rankcore"
chmod 0755 "${INSTALL_DIR}/rankcore"
info "Installed rankcore to ${INSTALL_DIR}/rankcore"

case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *) info "NOTE: ${INSTALL_DIR} is not on your PATH. Add it with:
  export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
esac

info "Registering the /rank skill with detected coding agents..."
if ! "${INSTALL_DIR}/rankcore" setup; then
    info "Skipping automatic setup. You can run 'rankcore setup' manually later."
fi

info "Done. Open a web project in your coding agent and type: /rank"
