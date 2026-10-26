$ErrorActionPreference='Stop'
$workflow=[IO.File]::ReadAllText((Join-Path $PSScriptRoot '../.github/workflows/release.yml'))
$match=[regex]::Match($workflow,'(?ms)^\s{10}\$assets = .*?^\s{10}if \(\$LASTEXITCODE -ne 0\) \{ exit \$LASTEXITCODE \}')
if(-not $match.Success){throw 'Release publication command not found'}
function gh { $script:captured=@($args); $global:LASTEXITCODE=0 }
function Get-ChildItem { [pscustomobject]@{Extension='.zip';Name='release.zip';FullName='release.zip'} }
$oldRef=$env:GITHUB_REF_NAME
$oldTag=$env:RELEASE_TAG
try {
 foreach($tag in @('v3.0.0-alpha1','v3.0.0')){
  $env:GITHUB_REF_NAME=$tag;$env:RELEASE_TAG=$tag
  & ([scriptblock]::Create($match.Value))
  if(($script:captured -contains '--prerelease') -ne $tag.Contains('-')){throw 'Preview flag was split or missing'}
  if($script:captured -contains '-'){throw 'Native CLI received a split flag'}
  if($script:captured[2] -ne $tag){throw 'Wrong release tag'}
 }
 Write-Output 'Release CLI arguments: preview and stable passed'
} finally {$env:GITHUB_REF_NAME=$oldRef;$env:RELEASE_TAG=$oldTag}
