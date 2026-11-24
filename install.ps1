# ccdrawu installer for Windows
# Usage: irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "vulcanshen/ccdrawu"
$installDir = "$env:LOCALAPPDATA\ccdrawu"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "Error: 32-bit systems are not supported." -ForegroundColor Red
    exit 1
}

# Get latest release tag: where the releases page's latest points, the
# redirect not followed, not GitHub's API, which answers 60 lookups an hour
# to everything on a network and then 403 (GitHub's latest is never a beta)
Write-Host "Fetching latest release..." -ForegroundColor Cyan
Add-Type -AssemblyName System.Net.Http
$handler = New-Object System.Net.Http.HttpClientHandler
$handler.AllowAutoRedirect = $false
$client = New-Object System.Net.Http.HttpClient($handler)
try {
    $location = $client.GetAsync("https://github.com/$repo/releases/latest").Result.Headers.Location
} catch {
    $location = $null
} finally {
    $client.Dispose()
}
if (-not $location -or "$location" -notmatch "/releases/tag/(v?([^/]+))$") {
    Write-Host "Error: Cannot find the latest release of $repo on GitHub." -ForegroundColor Red
    Write-Host "Check your connection, or download it from https://github.com/$repo/releases" -ForegroundColor Yellow
    exit 1
}
$tag = $Matches[1]
$version = $Matches[2]
Write-Host "Latest version: $version"

# Download
$fileName = "ccdrawu_${version}_windows_${arch}.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$tag/$fileName"
$tempFile = Join-Path $env:TEMP $fileName

Write-Host "Downloading $fileName..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing

# Extract
if (Test-Path $installDir) {
    try {
        Remove-Item $installDir -Recurse -Force -ErrorAction Stop
    } catch {
        Write-Host ""
        Write-Host "Error: Cannot update ccdrawu — the file is in use." -ForegroundColor Red
        Write-Host "Please close ccdrawu first, then run this installer again." -ForegroundColor Yellow
        Remove-Item $tempFile -ErrorAction SilentlyContinue
        exit 1
    }
}
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Expand-Archive -Path $tempFile -DestinationPath $installDir -Force
Remove-Item $tempFile

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to PATH." -ForegroundColor Green
    Write-Host "Please restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host "$installDir is already in PATH." -ForegroundColor Green
}

Write-Host ""
Write-Host "ccdrawu $version installed successfully!" -ForegroundColor Green
Write-Host "ccdrawu draws with Nerd Font glyphs: use a Nerd Font in your terminal profile." -ForegroundColor Yellow
Write-Host ""
Write-Host "Run 'ccdrawu install claude' to add the drawer to Claude Code," -ForegroundColor Cyan
Write-Host "then 'ccdrawu' to pick what it shows." -ForegroundColor Cyan
Write-Host "(In a new terminal: PATH changes take effect there.)" -ForegroundColor Cyan
