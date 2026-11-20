param(
    [string]$EnvFile = '.env',
    [Parameter(Mandatory = $true)][string[]]$GoArgs
)
$ErrorActionPreference = 'Stop'
$previous = @{}
try {
    foreach ($line in Get-Content -LiteralPath $EnvFile -Encoding UTF8) {
        $value = $line.Trim()
        if ($value.Length -eq 0 -or $value.StartsWith('#')) { continue }
        $pair = $value.Split('=', 2)
        if ($pair.Length -ne 2 -or $pair[0] -notmatch '^[A-Z][A-Z0-9_]*$') { throw "Linha inválida em $EnvFile" }
        $key = $pair[0]
        $previous[$key] = [Environment]::GetEnvironmentVariable($key, 'Process')
        [Environment]::SetEnvironmentVariable($key, $pair[1].Trim().Trim('"').Trim("'"), 'Process')
    }
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($null -ne $goCommand) { $goExecutable = $goCommand.Source }
    else {
        $workspaceRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
        $goExecutable = Join-Path $workspaceRoot '.local/tools/go/bin/go.exe'
        if (-not (Test-Path -LiteralPath $goExecutable)) { throw 'Instale Go 1.27.3 ou coloque go.exe no PATH.' }
        $env:GOPATH = Join-Path $workspaceRoot '.local/gopath'
        $env:GOMODCACHE = Join-Path $workspaceRoot '.local/gomodcache'
        $env:GOCACHE = Join-Path $workspaceRoot '.local/gocache'
        $env:GOTOOLCHAIN = 'local'
    }
    & $goExecutable @GoArgs
    if ($LASTEXITCODE -ne 0) { throw "Comando Go terminou com código $LASTEXITCODE" }
}
finally { foreach ($key in $previous.Keys) { [Environment]::SetEnvironmentVariable($key, $previous[$key], 'Process') } }
