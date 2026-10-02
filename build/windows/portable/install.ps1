param(
    [string]$InstallRoot = (Join-Path $env:LOCALAPPDATA 'Programs\Proxy Switch')
)

$ErrorActionPreference = 'Stop'
$sourceRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$targetRoot = [IO.Path]::GetFullPath($InstallRoot)
$exe = Join-Path $sourceRoot 'proxy-switch.exe'
if (-not (Test-Path -LiteralPath $exe -PathType Leaf)) {
    throw "Application file not found: $exe"
}

New-Item -ItemType Directory -Force -Path $targetRoot | Out-Null
Copy-Item -LiteralPath $exe -Destination (Join-Path $targetRoot 'proxy-switch.exe') -Force
$marker = Join-Path $targetRoot '.proxy-switch-install'
Set-Content -LiteralPath $marker -Value 'Proxy Switch user install' -Encoding ASCII
$webviewBootstrapper = Join-Path $sourceRoot 'MicrosoftEdgeWebview2Setup.exe'
if (Test-Path -LiteralPath $webviewBootstrapper -PathType Leaf) {
    Copy-Item -LiteralPath $webviewBootstrapper -Destination (Join-Path $targetRoot 'MicrosoftEdgeWebview2Setup.exe') -Force
}
Copy-Item -LiteralPath (Join-Path $sourceRoot 'uninstall.ps1') -Destination (Join-Path $targetRoot 'uninstall.ps1') -Force

$startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
New-Item -ItemType Directory -Force -Path $startMenu | Out-Null
$shortcutPath = Join-Path $startMenu 'Proxy Switch.lnk'
$shell = New-Object -ComObject WScript.Shell
$shortcut = $shell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = Join-Path $targetRoot 'proxy-switch.exe'
$shortcut.WorkingDirectory = $targetRoot
$shortcut.Description = 'Proxy Switch'
$shortcut.Save()

Write-Output "Installed to: $targetRoot"
Write-Output "Start menu shortcut: $shortcutPath"
