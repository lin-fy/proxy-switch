#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

function Invoke-Native([string]$Name, [scriptblock]$Command) {
    & $Command
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw "$Name failed with exit code $exitCode"
    }
}

$guard = Join-Path $PSScriptRoot 'validate-release.ps1'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$tempRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
$fixture = Join-Path $tempRoot ("proxy-switch-release-test-" + [guid]::NewGuid())

function Assert-Rejected([string]$Tag) {
    $rejected = $false
    try { & $guard -Tag $Tag -MainRef main | Out-Null }
    catch { $rejected = $true }
    if (-not $rejected) { throw "Release guard accepted invalid case: $Tag" }
}

New-Item -ItemType Directory -Path $fixture | Out-Null
Push-Location $fixture
try {
    Invoke-Native 'git init' { git init --quiet --initial-branch=main }
    Invoke-Native 'git config user.name' { git config user.name 'Release Guard Test' }
    Invoke-Native 'git config user.email' { git config user.email 'release-test@example.invalid' }
    New-Item -ItemType Directory -Force build/windows/nsis | Out-Null
    foreach ($path in @('build/config.yml', 'build/windows/info.json', 'build/windows/nsis/wails_tools.nsh')) {
        Copy-Item -LiteralPath (Join-Path $repoRoot $path) -Destination $path
    }
    # These checks use the real tracked version metadata as their positive fixture.
    $config = Get-Content build/config.yml -Raw
    if ($config -notmatch '(?m)^  version:\s*"([^"]+)"') { throw 'Fixture version missing' }
    $tag = "v$($Matches[1])"
    Invoke-Native 'git add' { git add build }
    Invoke-Native 'git commit' { git commit --quiet -m 'test fixture' }
    Invoke-Native 'git tag' { git tag $tag }
    if ((& $guard -Tag $tag -MainRef main) -ne $tag.Substring(1)) {
        throw 'Release guard rejected a valid release'
    }
    Assert-Rejected "$tag-rc.1"
    Assert-Rejected 'v65536.0.0'
    Assert-Rejected 'v999.998.997'
    Invoke-Native 'git tag' { git tag 'v9.8.7' }
    Assert-Rejected 'v9.8.7'
    Set-Content build/windows/info.json '{"fixed":{"file_version":"9.8.7"},"info":{"0000":{"ProductVersion":"9.8.7"}}}'
    Assert-Rejected $tag
    Copy-Item -LiteralPath (Join-Path $repoRoot 'build/windows/info.json') -Destination build/windows/info.json
    Invoke-Native 'git checkout' { git checkout --quiet -b feature }
    Invoke-Native 'git commit' { git commit --quiet --allow-empty -m 'unmerged change' }
    Invoke-Native 'git tag' { git tag 'v9.8.6' }
    Assert-Rejected 'v9.8.6'
    Assert-Rejected $tag
    Write-Host 'Release guard: valid release accepted; 7 invalid cases rejected.'
}
finally {
    Pop-Location
    $resolved = [IO.Path]::GetFullPath($fixture)
    if (-not $resolved.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Test cleanup escaped the temporary directory'
    }
    Remove-Item -LiteralPath $resolved -Recurse -Force
}
