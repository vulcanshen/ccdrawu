# ccdrawu uninstaller for Windows
# Usage: irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$installDir = "$env:LOCALAPPDATA\ccdrawu"

if (-not (Test-Path "$installDir\ccdrawu.exe")) {
    Write-Host "ccdrawu not found in $installDir" -ForegroundColor Red
    exit 1
}

# Take the drawer out of Claude Code while the command is still here; with no
# plugin there (or no Claude Code) it says so and the uninstall goes on.
& "$installDir\ccdrawu.exe" uninstall claude

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

# What the user picked: the drawer's items and the alias (tdp E3)
if ($env:CCDRAWU__CONFIG) {
    $configDir = $env:CCDRAWU__CONFIG
} elseif ($env:XDG_CONFIG_HOME) {
    $configDir = Join-Path $env:XDG_CONFIG_HOME "ccdrawu"
} else {
    $configDir = Join-Path $env:USERPROFILE ".config\ccdrawu"
}
Ask-Remove $configDir "ccdrawu settings (the drawer's items and the alias)"

# What the plugin writes for ccdrawu, a cache (tdp E3): in %LOCALAPPDATA%\ccdrawu\cache, gone above
# with the folder ccdrawu.exe is in, unless CCDRAWU__CACHE names another
if ($env:CCDRAWU__CACHE) {
    Ask-Remove $env:CCDRAWU__CACHE "ccdrawu cache (the catalog, the status line's values)"
}

# The data (tdp E3): the theme and status line ccdrawu replaced, put back above by uninstall claude;
# versions before kept the catalog there too
if ($env:CCDRAWU__DATA) {
    $dataDir = $env:CCDRAWU__DATA
} else {
    $dataDir = Join-Path $env:USERPROFILE ".ccdrawu"
}
Ask-Remove $dataDir "ccdrawu data (the theme and status line it replaced, to put back)"

Write-Host ""
Write-Host "ccdrawu uninstalled. Restart your terminal for PATH changes to take effect." -ForegroundColor Green
