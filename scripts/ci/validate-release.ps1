#Requires -Version 7.3
param(
    [Parameter(Mandatory)]
    [string]$Tag,
    [string]$MainRef = 'origin/main'
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

if ($Tag -notmatch '^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$') {
    throw 'Release tag must match vX.Y.Z without prerelease suffixes'
}
$version = $Tag.Substring(1)
foreach ($part in $version.Split('.')) {
    if ([decimal]$part -gt 65535) { throw 'Windows version components must be at most 65535' }
}

$tagExists = @(git tag --list $Tag)
if ($tagExists.Count -eq 0) { throw "Release tag does not exist: $Tag" }
$tagCommit = git rev-parse --verify "refs/tags/$Tag^{commit}"
$headCommit = git rev-parse HEAD
if ($tagCommit -ne $headCommit) { throw 'Checkout does not match the release tag' }
git merge-base --is-ancestor $tagCommit $MainRef
if ($LASTEXITCODE -ne 0) { throw "Release tag is not on $MainRef" }

$config = Get-Content build/config.yml -Raw
if ($config -notmatch '(?m)^  version:\s*"([^"]+)"') {
    throw 'Could not read info.version from build/config.yml'
}
if ($Matches[1] -ne $version) { throw 'Tag does not match build/config.yml info.version' }

$info = Get-Content build/windows/info.json -Raw | ConvertFrom-Json
if ($info.fixed.file_version -ne $version -or $info.info.'0000'.ProductVersion -ne $version) {
    throw 'Update build/windows/info.json to match the release version'
}
$nsis = Get-Content build/windows/nsis/wails_tools.nsh -Raw
if ($nsis -notmatch '(?m)^\s*!define INFO_PRODUCTVERSION "([^"]+)"' -or $Matches[1] -ne $version) {
    throw 'Update NSIS INFO_PRODUCTVERSION to match the release version'
}

# Emit only the validated version, for GITHUB_ENV and local callers.
Write-Output $version
