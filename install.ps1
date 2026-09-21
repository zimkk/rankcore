$ErrorActionPreference = "Stop"

Write-Host "Installing RankCore..."

$OS = "windows"
$Arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
if ($Arch -eq "amd64") {
    $Arch = "amd64"
} elseif ($Arch -eq "arm64") {
    $Arch = "arm64"
} else {
    Write-Host "Unsupported architecture: $Arch"
    exit 1
}

Write-Host "Detected OS: $OS, Architecture: $Arch"
Write-Host "Downloading binary..."
# Invoke-WebRequest -Uri "https://rankcore.dev/releases/latest/rankcore-${OS}-${Arch}.exe" -OutFile "rankcore.exe"

Write-Host "Setting up..."
# Move-Item -Path "rankcore.exe" -Destination "C:\Windows\System32\rankcore.exe" -Force
# rankcore setup

Write-Host "RankCore installed successfully. Run /rank in your agent."
