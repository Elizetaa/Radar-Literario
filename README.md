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
e [MariaDB Server](https://mariadb.org/download/)

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

Em outro terminal, partindo de `Backend`, inicie o frontend:

```bash
cd ../Frontend/radar_literario_integrado
npm ci
npm run dev
```

Abra o endereço exibido pelo Vite (http://localhost:5173).

## 5. Teste de Backend

Para executar o teste da API do Backend e verificar os registros no banco de dados, execute esta etapa com o Backend já iniciado.

```bash
cd ../Frontend/backend_tester
npm ci
npm run dev
```

Abra o endereço exibido pelo Vite (http://localhost:5174).

Para realizar alterações no Banco de Dados, utilize o login de usuário com permissões administrativas.

## 6. Carga de dados iniciais (Seed)

Para importar ao banco de dados e estar a API e o front, é necessário o registro de alguns dados no banco de dados. Para isso, foi criado o arquivo `/Data_seed/seed_catalog_populares.sql`, com alguns registro de obras, editoras, autores, categorias e edições de livros.

Execute o script via Dbeaver ou sistema gerenciador de banco de dados de preferência.