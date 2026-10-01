#Requires -Version 5.1
param(
    [switch]$BuildWindows
)

$ErrorActionPreference = 'Stop'
$verify = Join-Path $PSScriptRoot '..\..\..\..\scripts\ci\verify.ps1'
& $verify @PSBoundParameters
