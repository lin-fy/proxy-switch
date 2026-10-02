$ErrorActionPreference = 'Stop'
$installRoot = [IO.Path]::GetFullPath((Split-Path -Parent $MyInvocation.MyCommand.Path))
$shortcutPath = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\Proxy Switch.lnk'
if (Test-Path -LiteralPath $shortcutPath) {
    Remove-Item -LiteralPath $shortcutPath -Force
}

$marker = Join-Path $installRoot '.proxy-switch-install'
if (-not (Test-Path -LiteralPath $marker -PathType Leaf)) {
    throw "Refusing to remove unmarked install directory: $installRoot"
}
Remove-Item -LiteralPath $installRoot -Recurse -Force
Write-Output 'Proxy Switch has been uninstalled.'
