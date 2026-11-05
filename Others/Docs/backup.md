# Backup e restauração do MariaDB

As rotinas Windows estão em `Backend/scripts/backup_database.ps1` e `restore_database.ps1`. O backup é um dump lógico de um banco, com snapshot InnoDB (`--single-transaction`), rotinas, eventos, triggers e arquivo de manifesto SHA-256. As senhas são lidas de um arquivo de opções do cliente; não aparecem na linha de comando. Não aplique migrações enquanto o dump estiver em execução.

O caminho local foi validado em **27/10/2026 com MariaDB 13.0.2 no Windows**. O agendamento não foi criado nesta entrega: os comandos abaixo permitem configurá-lo com a frequência e a identidade desejadas. A existência dos scripts não significa que haja backup periódico operacional ou cópia fora da máquina.

## Credencial e destino

Prepare um arquivo local ignorado pelo Git, por exemplo `C:\Projetos\Radar-Literario\.local\backup.cnf`, usando a identidade de backup autorizada para o ambiente:

```ini
[client]
host=127.0.0.1
port=3306
user=usuario_de_backup
password=defina_localmente
```

A identidade precisa ler as tabelas, os triggers, eventos e metadados das rotinas do banco. As permissões de metadados de rotinas dependem da versão do MariaDB; verifique-as com o administrador. O ensaio desta entrega usou a credencial administrativa local somente nos bancos de teste. A API continua usando seu próprio usuário, sem root. Proteja tanto o arquivo de opções quanto os dumps com as permissões do Windows e o armazenamento protegido definido para o ambiente. Para restauração da fila de e-mail, preserve também a chave `OUTBOX_ENCRYPTION_KEY` por canal seguro; o dump não inclui o `.env`.

## Executar um backup

Na raiz do repositório, em PowerShell:

```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force
$backup = .\Backend\scripts\backup_database.ps1 `
  -DefaultsFile '.local\backup.cnf' `
  -Database 'radar_literario' `
  -MariaDBBin 'C:\Program Files\MariaDB 13.0\bin' `
  -BackupDirectory 'D:\Backups\RadarLiterario' `
  -RetentionDays 14
$backup.BackupFile
```

O diretório é criado quando necessário. Cada execução gera um nome único com instante UTC e um manifesto `.sql.json`. Após um dump bem-sucedido, a retenção remove apenas arquivos `radar_literario_*.sql` mais antigos que o prazo configurado e seus manifestos, diretamente no diretório escolhido. Não há remoção recursiva. Para copiar o backup para outro volume ou serviço, copie também seu manifesto.

Na instalação local desta sessão, o cliente MariaDB tentou TLS automaticamente e retornou `SEC_E_NO_CREDENTIALS`. O ensaio de loopback usou o parâmetro adicional **`-DisableTLS`**, que acrescenta `--skip-ssl`. Configure TLS para destinos que o exijam; o parâmetro não é ativado por padrão.

## Agendar e consultar

Este exemplo configura uma execução diária às 02:00, com retenção de 14 dias. Altere `$everyDays`, `$at` e `$retentionDays` conforme a política do ambiente e execute sob a identidade que acessará o arquivo de opções e o destino:

```powershell
$everyDays = 1
$at = '02:00'
$retentionDays = 14
$script = 'C:\Projetos\Radar-Literario\Backend\scripts\backup_database.ps1'
$options = 'C:\Projetos\Radar-Literario\.local\backup.cnf'
$destination = 'D:\Backups\RadarLiterario'
$arguments = '-NoProfile -ExecutionPolicy Bypass -File "{0}" -DefaultsFile "{1}" -Database radar_literario -BackupDirectory "{2}" -RetentionDays {3}' -f $script, $options, $destination, $retentionDays
$action = New-ScheduledTaskAction -Execute 'powershell.exe' -Argument $arguments
$trigger = New-ScheduledTaskTrigger -Daily -DaysInterval $everyDays -At $at
Register-ScheduledTask -TaskName 'RadarLiterario-Backup' -Action $action -Trigger $trigger -Description 'Dump lógico do Radar Literário'
Get-ScheduledTaskInfo -TaskName 'RadarLiterario-Backup'
```

Se o destino for esta conexão local sem TLS, acrescente ` -DisableTLS` a `$arguments`. No Agendador de Tarefas, confira a identidade, as condições de execução sem sessão iniciada e o resultado da última execução. Teste o agendamento e a restauração antes de registrar o backup como operacional. As instruções de agendamento não foram executadas nesta entrega.

## Restaurar para um banco de ensaio novo

A rotina recusa banco já existente e exige sufixo `_restore_test`. Ela confere SHA-256 antes de criar o destino, não executa `DROP DATABASE` e não sobrescreve os bancos da aplicação. O dump deve ter sido produzido pela rotina de backup, sem `CREATE DATABASE` ou `USE`.

```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force
.\Backend\scripts\restore_database.ps1 `
  -DefaultsFile '.local\restore.cnf' `
  -BackupFile 'D:\Backups\RadarLiterario\radar_literario_arquivo.sql' `
  -Database 'radar_literario_novo_restore_test' `
  -MariaDBBin 'C:\Program Files\MariaDB 13.0\bin'
```

Use uma identidade autorizada a criar esse banco e suas tabelas, rotinas, eventos e triggers. Durante a importação, cláusulas `DEFINER` das definições são removidas: os objetos passam a pertencer à identidade que restaura. Isso evita que o usuário de migração da origem, com privilégios limitados ao banco original, deixe os triggers inutilizáveis no novo banco. O arquivo original e seu manifesto permanecem intactos. Essa transformação não altera linhas de dados `INSERT`.

Depois de restaurar, confira a versão em `goose_db_version`, os dados, FKs, triggers e as regras da aplicação antes de escolher qualquer promoção para outro ambiente. Uma importação malsucedida preserva o novo banco de ensaio para diagnóstico. A rotina não promove o banco nem altera a conexão da API.

## Evidência executada

Foram feitos dumps reais de `radar_literario_test` para `.local/backups`, com manifesto SHA-256, e restaurações em bancos novos. O ensaio final está em **`radar_literario_verified_restore_test`**:

| Verificação | Resultado |
| --- | --- |
| Estrutura restaurada | 34 tabelas, 55 FKs, 2 triggers e migração 4, iguais à origem do dump |
| Dados fictícios UTF-8 | Biografia e diário com acentos preservados byte a byte, conferidos por `HEX` |
| Relações e leitura | Progresso 12, crédito 12 e dois comentários com vínculo pai/resposta preservados |
| Trigger restaurado | Alterar o pai de comentário foi rejeitado com erro 1644 (`Comment parent and post are immutable`) |
| CHECK restaurado | Página negativa foi rejeitada com erro 4025 |
| FK restaurada | Referência a leitor inexistente foi rejeitada com erro 1452 |
| Proteção de destino | Restaurar sobre banco existente foi recusado |
| Integridade do arquivo | Dump adulterado foi recusado por SHA-256 antes de criar o banco |
| Retenção | Uma fixture de arquivo com 15 dias foi removida com prazo 14; demais arquivos preservados |

A fixture própria foi removida do banco-fonte após o ensaio. Os bancos descartáveis de restauração e os arquivos locais foram preservados para inspeção. O primeiro ensaio identificou a necessidade de ajustar `DEFINER`; o ensaio final validou os efeitos dos triggers sob a identidade correta. Não foram copiados dados de produção, agendadas tarefas, executados contêineres ou configuradas cópias externas.

As tags do Compose foram conferidas em fontes oficiais: [MariaDB 11.8.5](https://mariadb.com/docs/release-notes/community-server/11.8/11.8.5), [imagem oficial MariaDB 11.8.5](https://hub.docker.com/layers/library/mariadb/11.8.5/images/sha256-5962c1371daf877b4131403b9d4c27aaff52702867b0fb1f483d6be488dfd2eb) e [publicação Mailpit v1.27.8](https://github.com/axllent/mailpit/actions/runs/17710003161). O Compose permaneceu sem execução porque Docker não estava disponível neste ambiente.
