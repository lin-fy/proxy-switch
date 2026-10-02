param(
    [string]$InstallRoot = (Join-Path $env:LOCALAPPDATA 'Programs\Codex Provider Hub')
)

$ErrorActionPreference = 'Stop'
$sourceRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$targetRoot = [IO.Path]::GetFullPath($InstallRoot)
$exe = Join-Path $sourceRoot 'codex-provider-hub.exe'
if (-not (Test-Path -LiteralPath $exe -PathType Leaf)) {
    throw "Application file not found: $exe"
}

New-Item -ItemType Directory -Force -Path $targetRoot | Out-Null
Copy-Item -LiteralPath $exe -Destination (Join-Path $targetRoot 'codex-provider-hub.exe') -Force
$marker = Join-Path $targetRoot '.codex-provider-hub-install'
Set-Content -LiteralPath $marker -Value 'Codex Provider Hub user install' -Encoding ASCII
$webviewBootstrapper = Join-Path $sourceRoot 'MicrosoftEdgeWebview2Setup.exe'
if (Test-Path -LiteralPath $webviewBootstrapper -PathType Leaf) {
    Copy-Item -LiteralPath $webviewBootstrapper -Destination (Join-Path $targetRoot 'MicrosoftEdgeWebview2Setup.exe') -Force
}
Copy-Item -LiteralPath (Join-Path $sourceRoot 'uninstall.ps1') -Destination (Join-Path $targetRoot 'uninstall.ps1') -Force

$startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
New-Item -ItemType Directory -Force -Path $startMenu | Out-Null
$shortcutPath = Join-Path $startMenu 'Codex Provider Hub.lnk'
$shell = New-Object -ComObject WScript.Shell
$shortcut = $shell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = Join-Path $targetRoot 'codex-provider-hub.exe'
$shortcut.WorkingDirectory = $targetRoot
$shortcut.Description = 'Codex Provider Hub'
$shortcut.Save()

Write-Output "Installed to: $targetRoot"
Write-Output "Start menu shortcut: $shortcutPath"
