# Backend do Radar Literário

API Go/Gin, serviços MVC, repositórios GORM/SQL e MariaDB. Contrato completo: [OpenAPI](../Docs/api.yaml); requisitos e evidências: [cobertura](../Docs/cobertura_backend.md); políticas: [decisões](../Docs/decisoes.md). O protótipo React continua independente e ainda não consome esta API.

## 1. Pré-requisitos

- Go **1.27.3**, conforme `go.mod`; [distribuição oficial](https://go.dev/dl/). Nesta máquina também foi preparado `../.local/tools/go/bin/go.exe`, sem instalação global.
- MariaDB: validado no Windows com **13.0.2**, serviço **MariaDB**, TCP **127.0.0.1:3306**. Alternativa Compose: MariaDB **11.8.5**, portas 3307/3308, volumes separados.
- PowerShell e, para backup/restauração, os executáveis da pasta `C:\Program Files\MariaDB 13.0\bin`. Docker Compose somente para a alternativa em contêiner.
- DBeaver é opcional. Nix também é opcional e não foi executado no Windows.

Todos os comandos seguintes partem de `Backend`. Para usar a toolchain local desta máquina:

```powershell
cd C:\Projetos\Radar-Literario\Backend
$env:PATH = (Resolve-Path '..\.local\tools\go\bin').Path + ';' + $env:PATH
$env:GOPATH = (Resolve-Path '..\.local').Path + '\gopath'
$env:GOMODCACHE = (Resolve-Path '..\.local').Path + '\gomodcache'
$env:GOCACHE = (Resolve-Path '..\.local').Path + '\gocache'
go version
```

Com Go instalado normalmente, basta `cd Backend`. O helper `scripts/with_env.ps1` também encontra a toolchain local. Se a política do PowerShell impedir scripts, a liberação abaixo afeta somente o terminal atual:

```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force
go mod download
```

## 2. Configuração

**Nesta máquina a preparação já foi executada:** bancos `radar_literario` e `radar_literario_test`; usuários `radar_app`, `radar_migrate`, `radar_test`; `.env`, `.env.migrate`, `.env.test` privados; primeiro administrador criado. Não repita o setup para recriar esses dados.

Para uma instalação local **nova**, inicie o banco (seção 3), defina a credencial administrativa e execute:

```powershell
$credential = Get-Credential -UserName 'root' -Message 'Acesso administrativo ao MariaDB local'
$env:DB_ADMIN_PASSWORD = $credential.GetNetworkCredential().Password
go run ./cmd/setup
Remove-Item Env:DB_ADMIN_PASSWORD
```

O comando cria bancos/usuários com senhas aleatórias, gera uma chave de 32 bytes para a outbox e grava os três arquivos de ambiente. Recusa bancos, usuários ou arquivos já existentes; não altera credenciais antigas. DDL não é transacional no MariaDB: em uma falha parcial, preserve os arquivos gerados para revisar as etapas aplicadas.

Se os bancos já existem, copie `.env.example` para `.env`, configure host/porta/usuário/senha próprios e mantenha outra configuração `.env.migrate` para o usuário com DDL. Gere a chave, sem colocar seu valor no Git:

```powershell
$keyBytes = New-Object byte[] 32
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($keyBytes)
[Convert]::ToBase64String($keyBytes) # colocar em OUTBOX_ENCRYPTION_KEY no arquivo privado
$rng.Dispose()
```

Não troque essa chave enquanto houver e-mails pendentes: eles são cifrados com AES-GCM. Arquivos `.env*` privados e `.local` estão ignorados pelo Git; `.env.example` contém apenas exemplos. `config.Load` carrega `.env` explicitamente e preserva variáveis já exportadas. O helper carrega o arquivo indicado apenas durante o comando Go.

| Variáveis efetivamente usadas | Default/unidade e finalidade |
| --- | --- |
| `APP_ENV` | `development`, `test` ou `production`; default development |
| `HTTP_HOST`, `HTTP_PORT` | `127.0.0.1`, `8080` |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | 127.0.0.1:3306, radar_literario, radar_app; senha obrigatória; API recusa root |
| `DB_MAX_OPEN`, `DB_MAX_IDLE`, `DB_CONN_MAX_LIFETIME`, `DB_TIMEOUT` | 20, 5, 5m, 5s; pool e timeouts de conexão/leitura/escrita |
| `OUTBOX_ENCRYPTION_KEY` | Obrigatória: 32 bytes aleatórios em base64; cifra e-mails e deriva CSRF autenticado |
| `CORS_ALLOWED_ORIGINS`, `CSRF_ALLOWED_ORIGINS` | Lista separada por vírgula, sem wildcard; default http://localhost:5173 |
| `SESSION_COOKIE_NAME`, `SESSION_IDLE_TIMEOUT`, `SESSION_ABSOLUTE_TIMEOUT` | radar_session, 30m, 168h; expiração validada no servidor |
| `COOKIE_SECURE`, `COOKIE_SAME_SITE` | false/lax em desenvolvimento; SameSite none exige Secure |
| `PASSWORD_RESET_TTL` | 30m, token de recuperação de uso único |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_TLS`, `SMTP_FROM` | localhost:1025, sem autenticação/TLS em desenvolvimento; remetente nao-responda@example.test; TLS significa STARTTLS |
| `PUBLIC_FRONTEND_URL` | http://localhost:5173; gera link `/reset-password?token=...` para o consumidor implementar |
| `REQUEST_TIMEOUT`, `MAX_BODY_BYTES`, `AUTH_RATE_LIMIT` | 15s, 1048576 bytes, 30 tentativas/IP/minuto nos fluxos de conta |
| `RANKING_MINIMUM_REVIEWS`, `COMMENT_MAX_DEPTH` | 1 avaliação mínima, profundidade 5; raiz tem profundidade zero |
| `CONTENT_BLOCKED_TERMS`, `CONTENT_BLOCKED_DOMAINS` | Listas opcionais separadas por vírgula; filtro determinístico, além de limites de texto, links e repetição |
| `RETENTION_DAYS` | 180 dias; usado por `cmd/maintenance` |
| `ADMIN_NAME`, `ADMIN_NICKNAME`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_BIRTH_DATE` | Apenas bootstrap administrativo; nascimento opcional; senha não aparece em argumentos ou logs |
| `TEST_DB_DSN` | Somente testes de integração; banco separado com nome terminado em `_test` |

Produção exige cookie seguro, SMTP com TLS, configuração adequada de remetente/frontend HTTPS e credenciais próprias. A idade é autodeclarada; não há verificação documental. Não há fluxo de verificação de e-mail habilitado, nem provedor externo de ISBN. Importação manual com ISBN válido e dados completos é suportada.

## 3. Iniciar o banco

No Windows local, confira o serviço e inicie-o em terminal administrativo apenas se estiver parado:

```powershell
Get-Service MariaDB
Start-Service MariaDB
```

**Alternativa Compose para instalação nova:** configure as duas senhas administrativas em variáveis privadas e use portas diferentes da instalação Windows:

```powershell
$env:MARIADB_ROOT_PASSWORD = (Get-Credential -UserName root -Message 'Senha root do contêiner de desenvolvimento').GetNetworkCredential().Password
$env:MARIADB_TEST_ROOT_PASSWORD = (Get-Credential -UserName root -Message 'Senha root do contêiner de testes').GetNetworkCredential().Password
docker compose up -d db mail
$env:DB_ADMIN_PASSWORD = $env:MARIADB_ROOT_PASSWORD
go run ./cmd/setup -port 3307 -account-host '%'
Remove-Item Env:DB_ADMIN_PASSWORD
```

O capturador Mailpit abre em http://localhost:8025 e recebe SMTP em localhost:1025. Para o banco de testes em contêiner, `docker compose --profile test up -d test-db`, porta 3308; prepare o banco/usuário exclusivo e ajuste `.env.test` para essa instância. A configuração padrão gerada por setup utiliza o banco isolado na mesma instância local. Os serviços Compose têm volumes e healthchecks; a execução dos contêineres não foi validada nesta máquina, que não possui Docker disponível.

No DBeaver, selecione MariaDB, host `127.0.0.1`, porta `3306` (ou 3307 em Compose), banco `radar_literario` e credenciais de `.env` para consulta da aplicação ou de `.env.migrate` para administração do esquema. O nome de serviço Windows `MariaDB` não é o nome do banco. Para o CLI local MariaDB 13, se ocorrer erro de TLS automático do Windows, use `--skip-ssl` **apenas nessa conexão local de desenvolvimento**.

## 4. Migrações

O executor embute **Goose v3.29.0** e os SQLs versionados. A API não executa DDL nem `AutoMigrate` na inicialização.

```powershell
.\scripts\with_env.ps1 -EnvFile .env.migrate -GoArgs @('run','./cmd/migrate','-command','up')
.\scripts\with_env.ps1 -EnvFile .env.migrate -GoArgs @('run','./cmd/migrate','-command','status')
.\scripts\with_env.ps1 -EnvFile .env.migrate -GoArgs @('run','./cmd/migrate','-command','version')
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('run','./cmd/migrate','-command','up')
```

`down` reverte a última migração e remove suas tabelas; execute apenas em banco descartável após revisar o SQL:

```powershell
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('run','./cmd/migrate','-command','down')
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('run','./cmd/migrate','-command','up')
```

Migrações já aplicadas em ambiente compartilhado são corrigidas por nova migração. IDs são BIGINT UNSIGNED, timestamps DATETIME(6) em UTC, collation padrão utf8mb4_unicode_ci e identidades/tokens usam collation binária. E-mail/apelido são normalizados em minúsculas; apelidos aceitam `[a-z0-9_]{3,30}`.

## 5. Primeiro administrador

```powershell
$env:ADMIN_NAME = 'Administrador Local'
$env:ADMIN_NICKNAME = 'admin_local'
$env:ADMIN_EMAIL = 'admin@example.test'
$env:ADMIN_PASSWORD = (Get-Credential -UserName admin_local -Message 'Senha do primeiro administrador').GetNetworkCredential().Password
go run ./cmd/admin
Remove-Item Env:ADMIN_PASSWORD
```

O comando usa as credenciais locais de banco, cria leitor+administrador de sistema, registra auditoria e não oferece promoção pública. Repeti-lo para o mesmo administrador é idempotente; outros administradores são geridos pela API autorizada. Protege-se o último administrador de sistema ativo. Nesta sessão foi criada a conta local solicitada, `elizeta`, com a exceção explícita `-allow-weak-local-password`, disponível somente em development. Cadastros públicos e bootstrap normal exigem de 8 a 72 bytes; bcrypt custo 12.

Papéis: `catalog_manager` mantém catálogo/sugestões; `user_manager` moderação/punições/FAQ/atendimento; `administrator` reúne os dois; `system_administrator` também gerencia administradores. A auditoria é limitada ao domínio de cada papel. Administradores ativos não podem ser punidos por outro administrador. Elevação/revogação de papel invalida sessões.

## 6. Seed opcional

```powershell
go run ./cmd/seed
```

Adiciona uma obra, autora, editora, gênero e edição inteiramente fictícios, sem contas nem senhas fixas. É idempotente e permitido somente em development/test após bootstrap. Não é executado pelas migrações ou pelo startup da API.

## 7. Executar a API

```powershell
go run ./cmd/api
```

Endereço padrão http://127.0.0.1:8080. `GET /health` verifica vivacidade; `GET /ready` verifica conexão e versão das migrações, retornando 503 sem detalhes internos em caso de falha. Ctrl+C solicita encerramento gracioso e fecha o pool. As antigas rotas fictícias `/api/usuarios` foram removidas.

Exemplo de sessão e CSRF no PowerShell, sem imprimir a senha:

```powershell
$base = 'http://127.0.0.1:8080/api/v1'
$preauth = Invoke-RestMethod "$base/auth/csrf" -SessionVariable radar
$credentials = Get-Credential -Message 'Conta do Radar Literário'
$json = @{ email=$credentials.UserName; password=$credentials.GetNetworkCredential().Password } | ConvertTo-Json
$login = Invoke-RestMethod "$base/auth/login" -Method Post -WebSession $radar -ContentType application/json -Headers @{'X-CSRF-Token'=$preauth.data.csrf_token} -Body $json
Invoke-RestMethod "$base/me" -WebSession $radar
Invoke-RestMethod "$base/auth/logout" -Method Post -WebSession $radar -Headers @{'X-CSRF-Token'=$login.data.csrf_token}
```

O protocolo CSRF também protege cadastro/login/recuperação: obtenha `GET /auth/csrf`, preserve o cookie preauth e envie o token no header. Login troca a sessão e devolve novo token. Em sessão existente, `GET /auth/csrf` fornece o token derivado dela. Requisições de navegador usam uma origem permitida; clientes sem Origin também precisam do par cookie+header. Cookies são HttpOnly, Path=/, SameSite definido, Secure em produção. Frontend deve enviar `credentials: 'include'`.

Listas usam `data` e `pagination` (page>=1, page_size 1–100); detalhes usam `data`; 204 não tem corpo. JSON desconhecido é recusado. Datas civis usam YYYY-MM-DD, timestamps ISO8601 UTC. O OpenAPI descreve filtros/enums e os erros estáveis. `reveal_spoilers=true`/`reveal_sensitive=true` revela texto somente quando a idade também permite. A classificação efetiva considera a classificação atual da obra.

## 8. Testes e contrato

```powershell
go mod tidy
go mod verify
go fmt ./...
go vet ./...
go test ./...
.\scripts\with_env.ps1 -EnvFile .env.test -GoArgs @('test','-tags=integration','-p=1','./...','-count=1')
go build ./cmd/api
go build ./cmd/admin
go run ./cmd/openapi -check
```

Após alterar uma rota ou DTO, atualize seus metadados em `internal/openapi` e execute `go run ./cmd/openapi`; o contrato é JSON compatível com YAML. Os testes verificam referências, parâmetros, schemas e correspondência de todas as operações com as rotas, também com o validador kin-openapi. Integrações usam `-p=1` para serializar pacotes: os cenários de retenção alteram o escopo inteiro do banco exclusivo de testes. A concorrência interna de cada cenário continua usando conexões independentes e fixtures próprias. Testes de SMTP abrem capturadores TCP locais; não comprovam entrega em provedor externo. `go test -race` depende de CGO e compilador C compatível no Windows; o resultado/limitação está registrado na matriz de cobertura.

No Linux, Go carrega `.env` igualmente. Para usar outra configuração em um subshell:

```sh
(set -a; . ./.env.migrate; set +a; go run ./cmd/migrate -command up)
(set -a; . ./.env.test; set +a; go test -tags=integration -p=1 ./... -count=1)
```

## 9. Backup, retenção e limites

Veja [backup e restauração](../Docs/backup.md), com scripts executáveis, frequência/retenção configuráveis e procedimento de ensaio. O backup foi restaurado em bancos descartáveis; **nenhuma tarefa de backup foi agendada** nesta sessão.

Excluir conta exige senha, revoga sessões/tokens, remove estante/eventos/interesses/preferências/outbox, elimina o suporte privado e anonimiza nome/e-mail/apelido/nascimento. Posts/comentários deixam de aparecer. IDs mínimos e evidências de moderação permanecem restritos pelo prazo técnico. Execute periodicamente:

```powershell
go run ./cmd/maintenance
```

Após RETENTION_DAYS, o comando elimina texto original de posts/comentários removidos, evidências de denúncias encerradas, motivos antigos de punição/auditoria, conteúdo de atendimento encerrado e propostas decididas; remove sessões/tokens/outbox antigos. Mantém IDs/estados mínimos para integridade. A rotina de retenção também precisa de agendamento operacional; não foi agendada. Uma política jurídica específica deve ser revisada pelo responsável.

Estatísticas usam períodos UTC `[início,fim)`: conclusões atuais pela data de término; páginas por eventos, sem somar snapshots. Correções/troca de edição não geram crédito. Redução explícita exige `change_kind=correction`; voltar a avançar até o máximo já creditado não duplica páginas. Remover item da estante apaga seus eventos e efeitos nas estatísticas, preservando posts. Múltiplas releituras independentes não são modeladas. Recomendações e rankings são calculados no SQL, sem dados fixos; fórmulas e desempates constam nas decisões por módulo.

Pendências operacionais: integração do frontend; SMTP real e entrega externa; HTTPS/hospedagem; agendamento de backup e retenção; revisão jurídica da política. A aplicação implementa controles técnicos e não declara conformidade legal automática.
