param(
    [Parameter(Mandatory = $true)][string]$DefaultsFile,
    [Parameter(Mandatory = $true)][string]$BackupFile,
    [string]$Database = 'radar_literario_restore_test',
    [string]$MariaDBBin = 'C:\Program Files\MariaDB 13.0\bin',
    [switch]$DisableTLS
)
$ErrorActionPreference = 'Stop'
if ($Database -notmatch '^[a-zA-Z][a-zA-Z0-9_]{0,50}_restore_test$') { throw 'A restauração de ensaio exige banco com sufixo _restore_test.' }
$defaultsPath = (Resolve-Path -LiteralPath $DefaultsFile).ProviderPath
$backupPath = (Resolve-Path -LiteralPath $BackupFile).ProviderPath
$clientExecutable = Join-Path $MariaDBBin 'mariadb.exe'
if (-not (Test-Path -LiteralPath $clientExecutable -PathType Leaf)) { throw 'Informe o diretório bin do MariaDB.' }
if (-not (Test-Path -LiteralPath "$backupPath.json" -PathType Leaf)) { throw 'O manifesto SHA-256 do backup está ausente.' }
$manifest = Get-Content -LiteralPath "$backupPath.json" -Encoding UTF8 -Raw | ConvertFrom-Json
$actualHash = (Get-FileHash -LiteralPath $backupPath -Algorithm SHA256).Hash.ToLowerInvariant()
if ($manifest.sha256 -ne $actualHash) { throw 'O SHA-256 do backup não confere com o manifesto.' }
# Dumps feitos por backup_database.ps1 não contêm CREATE DATABASE ou USE. Não
# aceite essas instruções em um arquivo que será restaurado em banco isolado.
foreach ($line in [IO.File]::ReadLines($backupPath)) {
    if ($line -match '^\s*(CREATE\s+DATABASE|USE\s+)') { throw 'O dump seleciona outro banco; use um backup gerado por backup_database.ps1.' }
}
$connectionArguments = @("--defaults-extra-file=$defaultsPath", '--default-character-set=utf8mb4')
if ($DisableTLS) { $connectionArguments += '--skip-ssl' }
$existing = & $clientExecutable @connectionArguments --batch --skip-column-names "--execute=SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$Database'"
if ($LASTEXITCODE -ne 0) { throw 'Não foi possível verificar o banco de destino.' }
if ([int]$existing -ne 0) { throw 'O banco de destino já existe. Escolha um novo banco de ensaio; a rotina não sobrescreve bancos.' }
& $clientExecutable @connectionArguments "--execute=CREATE DATABASE $Database CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
if ($LASTEXITCODE -ne 0) { throw 'Não foi possível criar o banco de ensaio.' }
# Envie o SQL em UTF-8 diretamente para stdin, evitando redirecionamento por
# outro shell. DEFINER das definições passa a ser o usuário da restauração:
# o usuário de migração da origem não tem privilégios no novo banco de ensaio.
$startInfo = [Diagnostics.ProcessStartInfo]::new()
$startInfo.FileName = $clientExecutable
$processArguments = $connectionArguments + @('--binary-mode', $Database)
foreach ($argument in $processArguments) { if ($argument.Contains('"') -or $argument.Contains("`r") -or $argument.Contains("`n")) { throw 'Argumento de processo inválido.' } }
$startInfo.Arguments = (($processArguments | ForEach-Object { '"' + $_ + '"' }) -join ' ')
$startInfo.UseShellExecute = $false
$startInfo.CreateNoWindow = $true
$startInfo.RedirectStandardInput = $true
$startInfo.RedirectStandardOutput = $true
$startInfo.RedirectStandardError = $true
$process = [Diagnostics.Process]::new()
$process.StartInfo = $startInfo
$reader = $null
try {
    if (-not $process.Start()) { throw 'Não foi possível iniciar o cliente MariaDB.' }
    $stdout = $process.StandardOutput.ReadToEndAsync()
    $stderr = $process.StandardError.ReadToEndAsync()
    $reader = [IO.StreamReader]::new($backupPath, [Text.UTF8Encoding]::new($false, $true), $true)
    $writer = [IO.StreamWriter]::new($process.StandardInput.BaseStream, [Text.UTF8Encoding]::new($false), 65536, $true)
    $writer.NewLine = "`n"
    try {
        while ($null -ne ($line = $reader.ReadLine())) {
            if ($line -match '^\s*(CREATE\b|/\*!\d+\s+(CREATE|DEFINER)\b)') {
                $line = [regex]::Replace($line, 'DEFINER\s*=\s*`(?:``|[^`])*`@`(?:``|[^`])*`', '')
                if ($line -match 'DEFINER\s*=') { throw 'A definição SQL usa um formato de DEFINER não suportado.' }
            }
            $writer.WriteLine($line)
        }
        $writer.Flush()
    }
    finally { $writer.Dispose(); $reader.Dispose(); $reader = $null }
    $process.StandardInput.Close()
    $process.WaitForExit()
    $outputText = $stdout.GetAwaiter().GetResult()
    $errorText = $stderr.GetAwaiter().GetResult()
    if ($process.ExitCode -ne 0) { throw "A restauração falhou com código $($process.ExitCode). O banco de ensaio foi preservado para diagnóstico; confira o dump e os privilégios sem publicar dados do SQL." }
}
finally { if ($null -ne $reader) { $reader.Dispose() }; $process.Dispose() }
$tables = & $clientExecutable @connectionArguments --batch --skip-column-names "--execute=SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$Database'"
if ($LASTEXITCODE -ne 0) { throw 'Não foi possível conferir as tabelas restauradas.' }
[pscustomobject]@{ Database = $Database; Tables = [int]$tables; BackupFile = $backupPath; SHA256 = $actualHash }
