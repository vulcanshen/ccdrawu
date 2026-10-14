# ccdrawu uninstaller for Windows
# Usage: irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$installDir = "$env:LOCALAPPDATA\ccdrawu"

if (-not (Test-Path "$installDir\ccdrawu.exe")) {
    Write-Host "ccdrawu not found in $installDir" -ForegroundColor Red
    exit 1
}

Remove-Item $installDir -Recurse -Force
Write-Host "removed $installDir" -ForegroundColor Green

# Remove from PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -like "*$installDir*") {
    $newPath = ($userPath -split ";" | Where-Object { $_ -ne $installDir }) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "removed $installDir from PATH" -ForegroundColor Green
}

# Ask before removing one folder
function Ask-Remove($dir, $what) {
    if (-not (Test-Path $dir)) { return }
    $answer = Read-Host "Remove $what in $dir? [y/N]"
    if ($answer -eq "y" -or $answer -eq "Y") {
        Remove-Item $dir -Recurse -Force
        Write-Host "removed $dir" -ForegroundColor Green
    } else {
        Write-Host "kept $dir" -ForegroundColor Yellow
    }
}

# What the user picked: the drawer's items and the alias (terminu design E3)
if ($env:CCDRAWU__CONFIG) {
    $configDir = $env:CCDRAWU__CONFIG
} elseif ($env:XDG_CONFIG_HOME) {
    $configDir = Join-Path $env:XDG_CONFIG_HOME "ccdrawu"
} else {
    $configDir = Join-Path $env:USERPROFILE ".config\ccdrawu"
}
Ask-Remove $configDir "ccdrawu settings (the drawer's items and the alias)"

# What the plugin writes: the catalog and vim's state per session
if ($env:CCDRAWU__DATA) {
    $dataDir = $env:CCDRAWU__DATA
} else {
    $dataDir = Join-Path $env:USERPROFILE ".ccdrawu"
}
Ask-Remove $dataDir "ccdrawu data (the catalog and vim's state)"

Write-Host ""
Write-Host "ccdrawu uninstalled. Restart your terminal for PATH changes to take effect." -ForegroundColor Green
Write-Host "The drawer itself is a Claude Code plugin: remove it in Claude Code with" -ForegroundColor Cyan
Write-Host "/plugin uninstall ccdrawu, and the statusLine entry in" -ForegroundColor Cyan
Write-Host "%USERPROFILE%\.claude\settings.json if you pointed it at 'ccdrawu statusline'." -ForegroundColor Cyan
