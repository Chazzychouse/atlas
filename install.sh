#!/bin/sh
# shellcheck disable=all
echo >/dev/null 2>&1 <#
# Polyglot installer: works as both POSIX sh and PowerShell
# Usage (sh):         curl -fsSL https://raw.githubusercontent.com/chazzychouse/atlas/main/install.sh | sh
# Usage (PowerShell): irm https://raw.githubusercontent.com/chazzychouse/atlas/main/install.sh | iex

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
exit 0
#>

# --- PowerShell installer ---
$ErrorActionPreference = 'Stop'

$Repo = "chazzychouse/atlas"
$Binary = "atlas"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:USERPROFILE ".atlas\bin" }

function Log($msg) { Write-Host "  $msg" }
function Err($msg) { throw "  Error: $msg" }

function Detect-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default  { Err "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Fetch-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name
}

function Verify-Checksum($FilePath, $ChecksumsPath, $ArchiveName) {
    $expected = (Get-Content $ChecksumsPath | Where-Object { $_ -match [regex]::Escape($ArchiveName) }) -replace '\s+.*$', ''
    if (-not $expected) {
        Err "checksum not found for $ArchiveName"
    }

    $actual = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
    if ($expected -ne $actual) {
        Err "checksum mismatch: expected $expected, got $actual"
    }
}

$arch = Detect-Arch
$version = Fetch-LatestVersion

if (-not $version) {
    Err "could not determine latest version"
}

$archive = "${Binary}_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$version/$archive"
$checksumsUrl = "https://github.com/$Repo/releases/download/$version/checksums.txt"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    Log "Downloading $Binary $version for windows/$arch..."
    Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmpDir $archive) -UseBasicParsing
    Invoke-WebRequest -Uri $checksumsUrl -OutFile (Join-Path $tmpDir "checksums.txt") -UseBasicParsing

    Log "Verifying checksum..."
    Verify-Checksum (Join-Path $tmpDir $archive) (Join-Path $tmpDir "checksums.txt") $archive

    Log "Extracting..."
    Expand-Archive -Path (Join-Path $tmpDir $archive) -DestinationPath $tmpDir -Force

    Log "Installing to $InstallDir..."
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    Move-Item -Path (Join-Path $tmpDir "$Binary.exe") -Destination (Join-Path $InstallDir "$Binary.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$userPath", "User")
        $env:Path = "$InstallDir;$env:Path"
        Log "Added $InstallDir to your PATH."
    }

    Log "Installed $Binary $version to $InstallDir\$Binary.exe"
}
finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
