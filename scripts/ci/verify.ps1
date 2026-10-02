#Requires -Version 5.1
param(
    [ValidateSet('all', 'frontend', 'backend')]
    [string]$Scope = 'all',
    [switch]$BuildWindows
)

$ErrorActionPreference = 'Stop'

function Invoke-Native([string]$Name, [scriptblock]$Command) {
    & $Command
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw "$Name failed with exit code $exitCode"
    }
}

function Get-GitDiff([string[]]$Paths) {
    $tracked = @(Invoke-Native 'git diff' { git diff --binary -- @Paths })
    $untracked = @(Invoke-Native 'git ls-files' { git ls-files --others --exclude-standard -- @Paths })
    return (@($tracked + $untracked) | Out-String)
}

Push-Location (Join-Path $PSScriptRoot '..\..')
try {
if ($BuildWindows -and $Scope -ne 'all') {
    throw '-BuildWindows requires the default all scope'
}

if ($Scope -in @('all', 'frontend')) {
    Push-Location (Join-Path $PSScriptRoot '..\..\frontend')
    try {
        Invoke-Native 'npm run format:check' { npm run format:check }
        Invoke-Native 'npm run lint' { npm run lint }
        Invoke-Native 'npm run build' { npm run build }
    }
    finally {
        Pop-Location
    }
}

if ($Scope -in @('all', 'backend')) {
    $goFiles = @(Invoke-Native 'git ls-files' { git ls-files --cached --others --exclude-standard '*.go' } | Sort-Object -Unique)
    if ($goFiles.Count -gt 0) {
        $unformatted = @(Invoke-Native 'gofmt' { gofmt -l $goFiles })
        if ($unformatted.Count -gt 0) {
            $unformatted | ForEach-Object { Write-Host $_ }
            throw 'Go files are not formatted'
        }
    }

    $modulePaths = @('go.mod', 'go.sum')
    $moduleDiff = Get-GitDiff $modulePaths
    Invoke-Native 'go mod tidy' { go mod tidy }
    if ((Get-GitDiff $modulePaths) -ne $moduleDiff) {
        throw 'go mod tidy changed go.mod or go.sum'
    }
    Invoke-Native 'go vet' { go vet . ./internal/... }
    Invoke-Native 'go test' { go test . ./internal/... -count=1 }
}

if ($BuildWindows) {
    if (-not (Get-Command wails3 -ErrorAction SilentlyContinue)) {
        $goBin = Join-Path (Invoke-Native 'go env GOPATH' { go env GOPATH }) 'bin'
        $env:PATH = "$goBin;$env:PATH"
    }
    if (-not (Get-Command wails3 -ErrorAction SilentlyContinue)) {
        throw 'Wails CLI not found; install v3.0.0-beta.27 and ensure it is available in go env GOPATH\bin'
    }

    $generatedPaths = @('go.mod', 'go.sum', 'frontend/package-lock.json', 'frontend/bindings')
    $generatedDiff = Get-GitDiff $generatedPaths
    Invoke-Native 'wails3 build' { wails3 build }
    if ((Get-GitDiff $generatedPaths) -ne $generatedDiff) {
        throw 'Wails build changed tracked module or generated binding files'
    }
    Invoke-Native 'portable package script' {
        powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot '..\..\build\windows\portable\package.ps1')
    }
}
}
finally {
    Pop-Location
}
