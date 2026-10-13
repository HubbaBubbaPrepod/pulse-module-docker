#requires -Version 7.4
param([string]$GoExe='go',[string]$Version='1.0.2',[string]$GoToolchain='go1.27.2')
$ErrorActionPreference='Stop'
$Version=$Version.TrimStart('v')
if($Version -notmatch '^\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?$'){throw 'Invalid module version'}
$root=Split-Path $PSScriptRoot -Parent
$output=Join-Path $root 'artifacts/binaries'
$release=Join-Path $root 'artifacts/releases'
New-Item -ItemType Directory -Force $output,$release | Out-Null
$manifest=Get-Content -LiteralPath (Join-Path $root 'manifest.json') -Raw | ConvertFrom-Json
$manifest.version=$Version
$binaries=@{}
$previousOS=$env:GOOS; $previousArch=$env:GOARCH; $previousCGO=$env:CGO_ENABLED; $previousToolchain=$env:GOTOOLCHAIN
Push-Location $root
try {
    $env:CGO_ENABLED='0'
    $env:GOTOOLCHAIN=$GoToolchain
    foreach($target in @(@{os='linux';arch='amd64';suffix=''},@{os='linux';arch='arm64';suffix=''},@{os='windows';arch='amd64';suffix='.exe'})){
        $env:GOOS=$target.os; $env:GOARCH=$target.arch
        $key=$target.os+'-'+$target.arch
        $path=Join-Path $output ('docker-'+$key+$target.suffix)
        & $GoExe build -trimpath -ldflags '-s -w' -o $path .
        if($LASTEXITCODE -ne 0){throw "Docker collector build failed: $key"}
        $binaries[$key]=@{sha256=(Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant();content=[Convert]::ToBase64String([IO.File]::ReadAllBytes($path))}
    }
    $manifest.binaries=$binaries
    $package=Join-Path $release 'docker-module.json'
    [IO.File]::WriteAllText($package,($manifest | ConvertTo-Json -Depth 20 -Compress),[Text.UTF8Encoding]::new($false))
    foreach($required in @('README.md','LICENSE','CHANGELOG.md')){if(-not (Test-Path $required)){throw "Source release requires $required"}}
    $stage=Join-Path $root ('artifacts/source-'+[guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory $stage | Out-Null
    foreach($file in Get-ChildItem -LiteralPath $root -File){if($file.Extension -eq '.go' -or $file.Name -in @('go.mod','go.sum','manifest.json','README.md','LICENSE','CHANGELOG.md','CONTRIBUTING.md','.gitignore')){Copy-Item -LiteralPath $file.FullName -Destination $stage}}
    foreach($directory in @('scripts','.github','docs')){if(Test-Path (Join-Path $root $directory)){Copy-Item -LiteralPath (Join-Path $root $directory) -Destination $stage -Recurse}}
    $archive=Join-Path $release "pulse-module-docker_v${Version}_source.zip"
    Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $archive -Force
    $checksums=@(@($package,$archive) | ForEach-Object {('{0}  {1}' -f (Get-FileHash -LiteralPath $_ -Algorithm SHA256).Hash.ToLowerInvariant(),(Split-Path $_ -Leaf))})
    [IO.File]::WriteAllLines((Join-Path $release 'checksums.txt'),$checksums,[Text.UTF8Encoding]::new($false))
    Write-Output "Module release files: $release"
} finally {
    $env:GOOS=$previousOS; $env:GOARCH=$previousArch; $env:CGO_ENABLED=$previousCGO; $env:GOTOOLCHAIN=$previousToolchain
    Pop-Location
}