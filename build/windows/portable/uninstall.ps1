$ErrorActionPreference = 'Stop'
$installRoot = [IO.Path]::GetFullPath((Split-Path -Parent $MyInvocation.MyCommand.Path))
$shortcutPath = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\Codex Provider Hub.lnk'
if (Test-Path -LiteralPath $shortcutPath) {
    Remove-Item -LiteralPath $shortcutPath -Force
}

$marker = Join-Path $installRoot '.codex-provider-hub-install'
if (-not (Test-Path -LiteralPath $marker -PathType Leaf)) {
    throw "Refusing to remove unmarked install directory: $installRoot"
}
Remove-Item -LiteralPath $installRoot -Recurse -Force
Write-Output 'Codex Provider Hub has been uninstalled.'
