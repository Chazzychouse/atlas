# Usage (PowerShell): irm https://raw.githubusercontent.com/chazzychouse/atlas/main/install.ps1 | iex
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
