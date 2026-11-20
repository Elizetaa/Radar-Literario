param(
    [Parameter(Mandatory = $true)][string]$DefaultsFile,
    [string]$Database = 'radar_literario',
    [string]$MariaDBBin = 'C:\Program Files\MariaDB 13.0\bin',
    [string]$BackupDirectory = (Join-Path $PSScriptRoot '../backups'),
    [ValidateRange(1, 3650)][int]$RetentionDays = 14,
    [switch]$DisableTLS
)
$ErrorActionPreference = 'Stop'
if ($Database -notmatch '^[a-zA-Z][a-zA-Z0-9_]{0,63}$') { throw 'Nome de banco inválido.' }
$defaultsPath = (Resolve-Path -LiteralPath $DefaultsFile).ProviderPath
$dumpExecutable = Join-Path $MariaDBBin 'mariadb-dump.exe'
$clientExecutable = Join-Path $MariaDBBin 'mariadb.exe'
if (-not (Test-Path -LiteralPath $dumpExecutable -PathType Leaf) -or -not (Test-Path -LiteralPath $clientExecutable -PathType Leaf)) { throw 'Informe o diretório bin do MariaDB.' }
$directoryPath = [IO.Path]::GetFullPath($BackupDirectory)
[IO.Directory]::CreateDirectory($directoryPath) | Out-Null
$backupName = '{0}_{1}_{2}.sql' -f $Database, [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssfffZ'), [Guid]::NewGuid().ToString('N').Substring(0, 8)
$backupPath = Join-Path $directoryPath $backupName
$connectionArguments = @("--defaults-extra-file=$defaultsPath", '--default-character-set=utf8mb4')
if ($DisableTLS) { $connectionArguments += '--skip-ssl' }
$serverVersion = & $clientExecutable @connectionArguments --batch --skip-column-names '--execute=SELECT VERSION()'
if ($LASTEXITCODE -ne 0) { throw 'Não foi possível consultar a versão do MariaDB.' }
$dumpArguments = $connectionArguments + @('--single-transaction', '--quick', '--routines', '--triggers', '--events', '--hex-blob', "--result-file=$backupPath", $Database)
& $dumpExecutable @dumpArguments
if ($LASTEXITCODE -ne 0) {
    if ((Test-Path -LiteralPath $backupPath) -and [IO.Path]::GetDirectoryName([IO.Path]::GetFullPath($backupPath)) -eq $directoryPath) { Remove-Item -LiteralPath $backupPath }
    throw 'O backup falhou; nenhum arquivo completo foi registrado.'
}
if ((Get-Item -LiteralPath $backupPath).Length -eq 0) { throw 'O backup gerou um arquivo vazio.' }
$hash = (Get-FileHash -LiteralPath $backupPath -Algorithm SHA256).Hash.ToLowerInvariant()
$manifest = [ordered]@{ database = $Database; created_at = [DateTime]::UtcNow.ToString('o'); server_version = "$serverVersion"; sha256 = $hash; retention_days = $RetentionDays; file = $backupName }
$manifest | ConvertTo-Json | Set-Content -LiteralPath "$backupPath.json" -Encoding UTF8
$cutoff = [DateTime]::UtcNow.AddDays(-$RetentionDays)
foreach ($oldBackup in Get-ChildItem -LiteralPath $directoryPath -Filter "${Database}_*.sql" -File) {
    if ($oldBackup.LastWriteTimeUtc -ge $cutoff) { continue }
    $oldPath = [IO.Path]::GetFullPath($oldBackup.FullName)
    if ([IO.Path]::GetDirectoryName($oldPath) -ne $directoryPath) { throw 'O arquivo de retenção está fora do diretório de backup.' }
    Remove-Item -LiteralPath $oldPath
    if (Test-Path -LiteralPath "$oldPath.json" -PathType Leaf) { Remove-Item -LiteralPath "$oldPath.json" }
}
[pscustomobject]@{ BackupFile = $backupPath; SHA256 = $hash; ServerVersion = "$serverVersion"; RetentionDays = $RetentionDays }
