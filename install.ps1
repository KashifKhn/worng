#Requires -Version 5.1
<#
.SYNOPSIS
    WORNG installer for Windows (PowerShell)

.DESCRIPTION
    Downloads and installs the WORNG esoteric programming language interpreter
    from GitHub Releases. Supports x64, arm64, and x86 (386).

.PARAMETER Version
    Install a specific version (e.g. "0.1.0"). Defaults to latest.

.PARAMETER InstallDir
    Directory to install the binary. Defaults to $Env:LOCALAPPDATA\worng\bin,
    falling back to $Env:USERPROFILE\.local\bin if LOCALAPPDATA is unset.

.PARAMETER NoModifyPath
    Skip adding the install directory to the user PATH.

.EXAMPLE
    irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1 | iex

.EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1))) -Version 0.1.0

.EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1))) -NoModifyPath
#>
[CmdletBinding()]
param(
    [string] $Version        = "",
    [string] $InstallDir     = "",
    [switch] $NoModifyPath
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$Repo       = "KashifKhn/worng"
$BinaryName = "worng"
$DocsUrl    = "https://github.com/KashifKhn/worng"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

function Write-Logo {
    Write-Host ""
    Write-Host "  ██╗    ██╗ ██████╗ ██████╗ ███╗   ██╗ ██████╗" -ForegroundColor Red
    Write-Host "  ██║    ██║██╔═══██╗██╔══██╗████╗  ██║██╔════╝" -ForegroundColor Red
    Write-Host "  ██║ █╗ ██║██║   ██║██████╔╝██╔██╗ ██║██║  ███╗" -ForegroundColor Red
    Write-Host "  ██║███╗██║██║   ██║██╔══██╗██║╚██╗██║██║   ██║" -ForegroundColor Red
    Write-Host "  ╚███╔███╔╝╚██████╔╝██║  ██║██║ ╚████║╚██████╔╝" -ForegroundColor Red
    Write-Host "   ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝" -ForegroundColor Red
    Write-Host ""
    Write-Host "        The esoteric programming language" -ForegroundColor DarkGray
    Write-Host "          if it looks right, it's wrong"  -ForegroundColor DarkGray
    Write-Host ""
}

function Write-Success { param([string]$Msg) Write-Host "  [+] $Msg" -ForegroundColor Green }
function Write-Fail    { param([string]$Msg) Write-Host "  [x] $Msg" -ForegroundColor Red }
function Write-Info    { param([string]$Msg) Write-Host "  --> $Msg" -ForegroundColor Cyan }
function Write-Warn    { param([string]$Msg) Write-Host "  [!] $Msg" -ForegroundColor Yellow }

function Get-ArchString {
    $arch = $Env:PROCESSOR_ARCHITECTURE
    # Under WOW64, the real machine arch is in PROCESSOR_ARCHITEW6432
    if ($Env:PROCESSOR_ARCHITEW6432) { $arch = $Env:PROCESSOR_ARCHITEW6432 }
    switch ($arch) {
        "AMD64"   { return "amd64" }
        "ARM64"   { return "arm64" }
        "x86"     { return "386"   }
        default   { return "amd64" }   # safe fallback for unknown Windows archs
    }
}

function Get-LatestVersion {
    $maxRetries = 3
    for ($i = 1; $i -le $maxRetries; $i++) {
        try {
            $response = Invoke-RestMethod `
                -Uri "https://api.github.com/repos/$Repo/releases/latest" `
                -Headers @{ "User-Agent" = "worng-installer" } `
                -ErrorAction Stop
            $tag = $response.tag_name -replace '^v', ''
            if ($tag) { return $tag }
        } catch {
            if ($i -lt $maxRetries) { Start-Sleep -Seconds 2 }
        }
    }
    return ""
}

function Get-InstalledVersion {
    $candidates = @()
    $inPath = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($inPath) { $candidates += $inPath.Source }
    if ($script:ResolvedInstallDir) {
        $candidate = Join-Path $script:ResolvedInstallDir "$BinaryName.exe"
        if (Test-Path $candidate) { $candidates += $candidate }
    }

    foreach ($bin in $candidates) {
        try {
            $output = & $bin version 2>&1
            if ($output -match 'v?(\d+\.\d+\.\d+)') {
                return $Matches[1]
            }
        } catch { }
    }
    return ""
}

function Add-ToUserPath {
    param([string]$Dir)

    $currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    $parts = $currentPath -split ";"
    if ($parts -contains $Dir) {
        return  # already present
    }
    $newPath = ($parts + $Dir) -join ";"
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    # Also update the current session
    $Env:PATH = "$Env:PATH;$Dir"
    Write-Success "Added to user PATH: $Dir"
}

function Invoke-Download {
    param([string]$Url, [string]$Dest)

    Write-Info "Downloading $BinaryName v$($script:ResolvedVersion)..."

    # Show a simple progress indicator
    $prevPref = $ProgressPreference
    $ProgressPreference = "SilentlyContinue"
    try {
        Invoke-WebRequest -Uri $Url -OutFile $Dest -UseBasicParsing -ErrorAction Stop
    } catch {
        $ProgressPreference = $prevPref
        throw
    }
    $ProgressPreference = $prevPref
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

Write-Logo

# Resolve version
if ($Version) {
    $script:ResolvedVersion = $Version -replace '^v', ''
    Write-Info "Installing version: v$($script:ResolvedVersion)"
} else {
    Write-Host "  Fetching latest version..." -NoNewline -ForegroundColor DarkGray
    $script:ResolvedVersion = Get-LatestVersion
    Write-Host "`r                                    `r" -NoNewline
    if (-not $script:ResolvedVersion) {
        Write-Fail "Could not determine latest version"
        Write-Info "Check releases: https://github.com/$Repo/releases"
        exit 1
    }
    Write-Success "Latest version: v$($script:ResolvedVersion)"
}

# Resolve install dir
if (-not $InstallDir) {
    $base = if ($Env:LOCALAPPDATA) { $Env:LOCALAPPDATA } else { "$Env:USERPROFILE\.local" }
    $InstallDir = Join-Path $base "worng\bin"
}
$script:ResolvedInstallDir = $InstallDir

# Check existing installation
$existingVersion = Get-InstalledVersion
if ($existingVersion) {
    if ($existingVersion -eq $script:ResolvedVersion) {
        Write-Info "Version v$existingVersion is already installed"
        exit 0
    } else {
        Write-Info "Upgrading from v$existingVersion to v$($script:ResolvedVersion)"
    }
}

# Determine platform
$arch = Get-ArchString
Write-Host ""
Write-Info "Platform: windows/$arch"
Write-Host ""

# Build download URL  — matches release.yml archive naming:
#   worng_<version>_windows_<arch>.zip
$archiveBase = "${BinaryName}_$($script:ResolvedVersion)_windows_${arch}"
$archive     = "$archiveBase.zip"
$downloadUrl = "https://github.com/$Repo/releases/download/v$($script:ResolvedVersion)/$archive"

# Download to temp dir
$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    $archivePath = Join-Path $tmpDir $archive
    try {
        Invoke-Download -Url $downloadUrl -Dest $archivePath
    } catch {
        Write-Fail "Failed to download $BinaryName v$($script:ResolvedVersion)"
        Write-Info "Check releases: https://github.com/$Repo/releases"
        exit 1
    }

    # Extract
    try {
        Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force
    } catch {
        Write-Fail "Failed to extract archive"
        exit 1
    }

    $exeName    = "$BinaryName.exe"
    $extractedBin = Join-Path $tmpDir $exeName

    if (-not (Test-Path $extractedBin)) {
        Write-Fail "Binary not found in archive: $exeName"
        exit 1
    }

    # Create install dir and copy binary
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $destBin = Join-Path $InstallDir $exeName
    Copy-Item -Path $extractedBin -Destination $destBin -Force
    Write-Success "Installed to $destBin"

} finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}

# Add to PATH
if (-not $NoModifyPath) {
    Add-ToUserPath -Dir $InstallDir
}

Write-Host ""
Write-Info "Run 'worng --help' to get started"
Write-Info "Docs: $DocsUrl"
Write-Host ""
