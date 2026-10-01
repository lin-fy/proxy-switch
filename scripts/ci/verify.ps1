#Requires -Version 7.3
param(
    [ValidateSet('all', 'frontend', 'backend')]
    [string]$Scope = 'all',
    [switch]$BuildWindows
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

function Get-GitDiff([string[]]$Paths) {
    return (@(git diff --binary -- @Paths; git ls-files --others --exclude-standard -- @Paths) | Out-String)
}

Push-Location (Join-Path $PSScriptRoot '..\..')
try {
if ($BuildWindows -and $Scope -ne 'all') {
    throw '-BuildWindows requires the default all scope'
}

if ($Scope -in @('all', 'frontend')) {
    Push-Location (Join-Path $PSScriptRoot '..\..\frontend')
    try {
        npm run format:check
        npm run lint
        npm run build
    }
    finally {
        Pop-Location
    }
}

if ($Scope -in @('all', 'backend')) {
    $goFiles = @(git ls-files --cached --others --exclude-standard '*.go' | Sort-Object -Unique)
    if ($goFiles.Count -gt 0) {
        $unformatted = @(gofmt -l $goFiles)
        if ($unformatted.Count -gt 0) {
            $unformatted | ForEach-Object { Write-Host $_ }
            throw 'Go files are not formatted'
        }
    }

    $modulePaths = @('go.mod', 'go.sum')
    $moduleDiff = Get-GitDiff $modulePaths
    go mod tidy
    if ((Get-GitDiff $modulePaths) -ne $moduleDiff) {
        throw 'go mod tidy changed go.mod or go.sum'
    }
    go vet . ./internal/...
    go test . ./internal/... -count=1
}

if ($BuildWindows) {
    if (-not (Get-Command wails3 -ErrorAction SilentlyContinue)) {
        $goBin = Join-Path (go env GOPATH) 'bin'
        $env:PATH = "$goBin;$env:PATH"
    }
    if (-not (Get-Command wails3 -ErrorAction SilentlyContinue)) {
        throw 'Wails CLI not found; install v3.0.0-beta.26 and ensure it is available in go env GOPATH\bin'
    }

    $generatedPaths = @('go.mod', 'go.sum', 'frontend/package-lock.json', 'frontend/bindings')
    $generatedDiff = Get-GitDiff $generatedPaths
    wails3 build
    if ((Get-GitDiff $generatedPaths) -ne $generatedDiff) {
        throw 'Wails build changed tracked module or generated binding files'
    }
    powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot '..\..\build\windows\portable\package.ps1')
}
}
finally {
    Pop-Location
}
