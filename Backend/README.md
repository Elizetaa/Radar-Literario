# Backend — Radar Literário

API Go com MariaDB. Execute os comandos da aplicação na pasta `Backend`.
Requisitos: Go conforme `go.mod` (atualmente **1.27.3**) e Node.js **22.12+** com npm.

## 1. Instalar as ferramentas

### Ubuntu

```bash
sudo apt update
sudo apt install mariadb-server mariadb-client golang-go
sudo systemctl enable --now mariadb
```

Instale Node.js 22.12+ pelo [site oficial](https://nodejs.org/en/download).
Se o Go do Ubuntu for anterior a 1.21, instale pelo [site oficial](https://go.dev/doc/install).
Com Go 1.21+, `go env -w GOTOOLCHAIN=auto` permite baixar a versão exigida pelo projeto.

### Windows

Instale [Go](https://go.dev/doc/install), [Node.js LTS com npm](https://nodejs.org/en/download)
e [MariaDB Server](https://mariadb.org/download/). No instalador do MariaDB,
defina a senha de `root`, habilite o serviço e mantenha a porta `3306`.
Reabra o PowerShell; adicione a pasta `bin` do MariaDB ao PATH para usar `mariadb`.

### Nix / NixOS

Na raiz que contém `flake.nix`:

```bash
nix develop
```

O ambiente fornece Go, Node.js/npm e MariaDB; o servidor precisa ser iniciado separadamente.
No NixOS, adicione à configuração do sistema:

```nix
services.mysql = {
  enable = true;
  package = pkgs.mariadb;
  settings.mysqld.bind-address = "127.0.0.1";
};
```

Aplique com `sudo nixos-rebuild switch`. Em Ubuntu com Nix, use o serviço da seção Ubuntu.
O flake atual atende `x86_64-linux`; mantenha o terminal de `nix develop` aberto.

Confira: `go version`, `node --version` e `npm --version`.

## 2. Criar o banco

Abra o MariaDB como administrador:

```bash
# Ubuntu / NixOS
sudo mariadb
```

```powershell
# Windows
mariadb -u root -p
```

Execute o SQL abaixo para uma instalação local nova. Se banco e usuário já existem,
use as credenciais existentes e pule esta etapa.

```sql
CREATE DATABASE radar_literario CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'radar'@'localhost' IDENTIFIED BY '123';
GRANT ALL PRIVILEGES ON radar_literario.* TO 'radar'@'localhost';
EXIT;
```

A senha `123` é apenas para desenvolvimento local. O usuário recebe permissão
para executar as migrações nesse banco.

## 3. Configurar o .env

Na pasta `Backend`, copie o modelo se ainda não houver `.env`:

```bash
# Ubuntu / Nix
cp .env.example .env
```

```powershell
# Windows
Copy-Item .env.example .env
```

Edite os campos abaixo e mantenha as demais opções do modelo:

```dotenv
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=radar_literario
DB_USER=radar
DB_PASSWORD=123
```

Gere a chave e copie a saída para `OUTBOX_ENCRYPTION_KEY` no `.env`:

```bash
# Ubuntu / Nix
head -c 32 /dev/urandom | base64
```

```powershell
# Windows
$bytes = New-Object byte[] 32
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes)
[Convert]::ToBase64String($bytes)
$rng.Dispose()
```

Preserve a chave se ela já estiver configurada. O `.env` é privado e ignorado pelo Git.

## 4. Executar

Na pasta `Backend`, em qualquer sistema:

```bash
go mod download
go run ./cmd/migrate -command up
go run ./cmd/api
```

API: http://127.0.0.1:8080. `/health` verifica o processo; `/ready` verifica
banco e migrações. Se indicar migrações pendentes, execute o comando de migração acima.

O `.env.example` usa SMTP em `localhost:1025`. Para enviar e-mails, configure um
servidor SMTP ou um capturador local, como Mailpit. Com Docker instalado:

```bash
docker compose up -d mail
```

Os e-mails capturados aparecem em http://localhost:8025.

Em outro terminal, partindo de `Backend`, inicie o frontend:

```bash
cd ../Frontend/radar_literario
npm ci
npm run dev
```

Abra o endereço exibido pelo Vite (normalmente http://localhost:5173).

## Referências

- [Contrato da API](../Docs/api.yaml)
- [Backup e restauração](../Docs/backup.md)
- [Decisões do backend](../Docs/decisoes.md)
- Instalação: [MariaDB](https://mariadb.com/docs/server/mariadb-quickstart-guides/installing-mariadb-server-guide), [MariaDB no NixOS](https://wiki.nixos.org/wiki/Mysql), [toolchains Go](https://go.dev/doc/toolchain).
