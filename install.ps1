# RankCore installer for Windows.
# Downloads the latest precompiled release from GitHub, verifies its
# SHA-256 checksum, installs it to a user-writable directory (no admin),
# and registers the /rank skill with detected coding agents.
$ErrorActionPreference = "Stop"

$Repo = "zimkk/rankcore"
$InstallDir = if ($env:RANKCORE_INSTALL_DIR) { $env:RANKCORE_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\RankCore" }
$Version = if ($env:RANKCORE_VERSION) { $env:RANKCORE_VERSION } else { "latest" }

function Fail($Message) {
    Write-Host "error: $Message" -ForegroundColor Red
    exit 1
}

$Arch = switch ($env:PROCESSOR_ARCHITECTURE.ToUpper()) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { Fail "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$Asset = "rankcore-windows-$Arch.exe"
if ($Version -eq "latest") {
    $BaseUrl = "https://github.com/$Repo/releases/latest/download"
} else {
    $BaseUrl = "https://github.com/$Repo/releases/download/$Version"
}

Write-Host "Installing RankCore (windows/$Arch)..."

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("rankcore-install-" + [System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    $AssetPath = Join-Path $TmpDir $Asset
    $ChecksumsPath = Join-Path $TmpDir "checksums.txt"

    Write-Host "Downloading $Asset from GitHub Releases..."
    try {
        Invoke-WebRequest -Uri "$BaseUrl/$Asset" -OutFile $AssetPath -UseBasicParsing
    } catch {
        Fail "Download failed. A precompiled release for windows/$Arch may not be published yet.`nBuild from source instead (requires Go 1.23+):`n  go install github.com/$Repo/cmd/rankcore@latest"
    }
    try {
        Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath -UseBasicParsing
    } catch {
        Fail "Could not download checksums.txt - release assets may be incomplete."
    }

    Write-Host "Verifying SHA-256 checksum..."
    $Expected = (Get-Content $ChecksumsPath | Where-Object { $_ -match "\s$([regex]::Escape($Asset))$" } | ForEach-Object { ($_ -split '\s+')[0] }) | Select-Object -First 1
    if (-not $Expected) { Fail "No checksum entry found for $Asset." }
    $Actual = (Get-FileHash -Path $AssetPath -Algorithm SHA256).Hash.ToLower()
    if ($Expected.ToLower() -ne $Actual) { Fail "Checksum mismatch for $Asset. Aborting." }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $Dest = Join-Path $InstallDir "rankcore.exe"
    Move-Item -Path $AssetPath -Destination $Dest -Force
    Write-Host "Installed rankcore to $Dest"

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not ($UserPath -split ";" | Where-Object { $_ -eq $InstallDir })) {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH (restart your terminal to pick it up everywhere)."
    }

    Write-Host "Registering the /rank skill with detected coding agents..."
    try {
        & $Dest setup
    } catch {
        Write-Host "Skipping automatic setup. You can run 'rankcore setup' manually later."
    }

    Write-Host "Done. Open a web project in your coding agent and type: /rank"
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
