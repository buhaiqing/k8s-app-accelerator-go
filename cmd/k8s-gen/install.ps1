# install.ps1 — one-shot install / update for the k8s-gen CLI on Windows
#
# Usage (run in PowerShell as a normal user):
#   irm https://raw.githubusercontent.com/buhaiqing/k8s-app-accelerator-go/main/cmd/k8s-gen/install.ps1 | iex
#   # install a specific version:
#   $Version = "1.0.0"; irm ... | iex
#
$ErrorActionPreference = "Stop"
$Repo = "buhaiqing/k8s-app-accelerator-go"
$Binary = "k8s-gen"

# Detect OS / arch
$OS = "windows"
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") { $Arch = "amd64" }
elseif ($Arch -eq "ARM64") { $Arch = "arm64" }
else { throw "unsupported architecture: $Arch" }

# Install dir
if (-not $env:INSTALL_DIR) {
    $InstallDir = "$env:LOCALAPPDATA\Programs\k8s-gen"
} else {
    $InstallDir = $env:INSTALL_DIR
}
$VersionFile = Join-Path $InstallDir "$Binary.version"

# Resolve version
if ($Version) {
    $Version = $Version.TrimStart("v")
    $Tag = "v$Version"
} else {
    $Response = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    $Tag = $Response.tag_name
    $Version = $Tag.TrimStart("v")
}

# Skip if up to date
if (Test-Path $VersionFile) {
    $Current = Get-Content $VersionFile -Raw -ErrorAction SilentlyContinue
    if ($Current -eq $Version) {
        Write-Host "✅ $Binary $Version is already installed (nothing to do)"
        exit 0
    }
}

# Download
$AssetName = "${Binary}_${OS}-${Arch}.tar.gz"
$Url = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
$Tmp = [System.IO.Path]::GetTempPath()
$DownloadFile = Join-Path $Tmp $AssetName
Write-Host "downloading $Url"
try {
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest $Url -OutFile $DownloadFile -UseBasicParsing -TimeoutSec 60
} catch {
    $AssetName = "${Binary}_${OS}-${Arch}.zip"
    $Url = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
    Invoke-WebRequest $Url -OutFile $DownloadFile -UseBasicParsing -TimeoutSec 60
}

# Extract
$ExtractDir = Join-Path $Tmp "k8s-gen-extract"
New-Item -ItemType Directory -Force -Path $ExtractDir | Out-Null
if ($AssetName -match "\.zip$") {
    Expand-Archive -Path $DownloadFile -DestinationPath $ExtractDir -Force
} else {
    tar -xzf $DownloadFile -C $ExtractDir
}

$Src = Get-ChildItem $ExtractDir -File -Recurse | Where-Object { $_.Name -eq "$Binary.exe" -or $_.Name -eq $Binary } | Select-Object -First 1
if (-not $Src) { throw "binary not found in archive" }

# Install
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item $Src.FullName (Join-Path $InstallDir "$Binary.exe") -Force
$Version | Set-Content $VersionFile -NoNewline

Write-Host "✅ $Binary $Version installed at $InstallDir\$Binary.exe"
& (Join-Path $InstallDir "$Binary.exe") --help | Select-Object -First 5
