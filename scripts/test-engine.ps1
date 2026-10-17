#requires -Version 7.4
param(
    [Parameter(Mandatory)][string]$Collector,
    [Parameter(Mandatory)][string]$FixtureImage,
    [string]$Socket=""
)
$ErrorActionPreference='Stop'
$collectorPath=(Resolve-Path -LiteralPath $Collector).Path
$fixtureId=''
function Invoke-FixtureOperation([string]$Action,[string[]]$Grants=@()) {
    $invocation=@{protocol=2;trigger='manual';action=$Action;input=@{containerId=$fixtureId};config=@{socket=$Socket};grants=$Grants} | ConvertTo-Json -Depth 5 -Compress
    $response=$invocation | & $collectorPath
    if($LASTEXITCODE -ne 0){throw "Collector failed: $Action"}
    return ($response | ConvertFrom-Json)
}
function Get-FixtureState {
    $value=& docker inspect --format '{{.State.Status}}' $fixtureId
    if($LASTEXITCODE -ne 0){throw 'Could not inspect fixture'}
    return $value.Trim()
}
try {
    # Existing image only; no ports, network, host volumes or production targets.
    $fixtureId=(& docker run --pull never --detach --network none --restart no --entrypoint sh $FixtureImage -c 'echo pulse-module-isolated-fixture; exec sleep 600').Trim()
    if($LASTEXITCODE -ne 0 -or $fixtureId -notmatch '^[a-f0-9]{64}$'){throw 'Could not create isolated fixture'}
    $logs=Invoke-FixtureOperation 'logs'
    if($logs.status -ne 'ok' -or $logs.data.containerId -ne $fixtureId -or ($logs.data.logs -join "`n") -notmatch 'pulse-module-isolated-fixture'){throw ("Selected-container logs failed: status={0}, error={1}, lines={2}" -f $logs.status,$logs.error,($logs.data.logs -join " | "))}
    $denied=Invoke-FixtureOperation 'stop'
    if($denied.status -ne 'error' -or (Get-FixtureState) -ne 'running'){throw 'Read-only request changed container state'}
    foreach($action in @('stop','start','restart')) {
        $response=Invoke-FixtureOperation $action @('actions.execute')
        $expected=if($action -eq 'stop'){'exited'}else{'running'}
        if($response.status -ne 'ok' -or $response.data.containerId -ne $fixtureId -or (Get-FixtureState) -ne $expected){throw "Container action failed: $action"}
    }
    Write-Output 'Real Docker Engine fixture: logs, denied stop, stop, start and restart passed.'
} finally {
    if($fixtureId -match '^[a-f0-9]{64}$') {
        & docker rm --force $fixtureId | Out-Null
        if($LASTEXITCODE -ne 0){Write-Warning "Could not remove isolated fixture $fixtureId"}
    }
}
