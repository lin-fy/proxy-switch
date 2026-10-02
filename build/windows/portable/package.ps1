$ErrorActionPreference = 'Stop'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\..'))
$binary = Join-Path $repoRoot 'bin\proxy-switch.exe'
if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) {
    throw "Run wails3 build first: $binary"
}

$staging = Join-Path $repoRoot 'bin\proxy-switch-windows-amd64'
$archive = Join-Path $repoRoot 'bin\proxy-switch-windows-amd64.zip'
if (Test-Path -LiteralPath $staging) { Remove-Item -LiteralPath $staging -Recurse -Force }
if (Test-Path -LiteralPath $archive) { Remove-Item -LiteralPath $archive -Force }
New-Item -ItemType Directory -Force -Path $staging | Out-Null
Copy-Item -LiteralPath $binary -Destination (Join-Path $staging 'proxy-switch.exe')
$bootstrapper = Join-Path $repoRoot 'build\windows\nsis\MicrosoftEdgeWebview2Setup.exe'
if (Test-Path -LiteralPath $bootstrapper -PathType Leaf) {
    Copy-Item -LiteralPath $bootstrapper -Destination (Join-Path $staging 'MicrosoftEdgeWebview2Setup.exe')
}
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'install.ps1') -Destination (Join-Path $staging 'install.ps1')
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'uninstall.ps1') -Destination (Join-Path $staging 'uninstall.ps1')
Compress-Archive -Path (Join-Path $staging '*') -DestinationPath $archive -CompressionLevel Optimal
Remove-Item -LiteralPath $staging -Recurse -Force
Write-Output "Created: $archive"
