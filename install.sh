#!/bin/sh
# Usage: curl -fsSL https://raw.githubusercontent.com/chazzychouse/atlas/main/install.sh | sh
# PowerShell (Windows): irm https://raw.githubusercontent.com/chazzychouse/atlas/main/install.ps1 | iex

set -eu

REPO="chazzychouse/atlas"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="atlas"

main() {
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(fetch_latest_version)"

  if [ -z "$version" ]; then
    err "could not determine latest version"
  fi

  archive="${BINARY}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${archive}"
  checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  log "Downloading ${BINARY} ${version} for ${os}/${arch}..."
  download "$url" "${tmpdir}/${archive}"
  download "$checksums_url" "${tmpdir}/checksums.txt"

  log "Verifying checksum..."
  verify_checksum "${tmpdir}/${archive}" "${tmpdir}/checksums.txt" "$archive"

  log "Extracting..."
  tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"

  log "Installing to ${INSTALL_DIR}..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  fi

  chmod +x "${INSTALL_DIR}/${BINARY}"
  log "Installed ${BINARY} ${version} to ${INSTALL_DIR}/${BINARY}"
}

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

fetch_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
      grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" |
      grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//'
  else
    err "curl or wget is required"
  fi
}

download() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$2" "$1"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$2" "$1"
  else
    err "curl or wget is required"
  fi
}

verify_checksum() {
  archive_path="$1"
  checksums_file="$2"
  archive_name="$3"

  expected="$(grep "$archive_name" "$checksums_file" | awk '{print $1}')"
  if [ -z "$expected" ]; then
    err "checksum not found for ${archive_name}"
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$archive_path" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
  else
    log "Warning: no sha256 tool found, skipping checksum verification"
    return 0
  fi

  if [ "$expected" != "$actual" ]; then
    err "checksum mismatch: expected ${expected}, got ${actual}"
  fi
}

log() { printf '  %s\n' "$*"; }
err() { log "Error: $*" >&2; exit 1; }

main
