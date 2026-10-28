// Command setup prepares a new local development installation without fixed application passwords.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host MariaDB local")
	port := flag.Int("port", 3306, "porta TCP")
	user := flag.String("admin-user", "root", "usuário administrativo MariaDB")
	accountHost := flag.String("account-host", "127.0.0.1", "host dos usuários SQL: 127.0.0.1, localhost ou % (Compose)")
	flag.Parse()
	if e := run(*host, *port, *user, *accountHost); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func secret(n int) (string, error) {
	b := make([]byte, n)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func run(host string, port int, adminUser, accountHost string) error {
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		return fmt.Errorf("setup destina-se somente à instalação local de desenvolvimento")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("porta inválida")
	}
	if accountHost != "127.0.0.1" && accountHost != "localhost" && accountHost != "%" {
		return fmt.Errorf("account-host inválido")
	}
	password, present := os.LookupEnv("DB_ADMIN_PASSWORD")
	if !present {
		return fmt.Errorf("defina DB_ADMIN_PASSWORD sem colocá-la no código ou no Git")
	}
	for _, file := range []string{".env", ".env.migrate", ".env.test"} {
		if _, e := os.Stat(file); e == nil {
			return fmt.Errorf("%s já existe; setup não sobrescreve configurações locais", file)
		} else if !os.IsNotExist(e) {
			return e
		}
	}
	d := mysql.NewConfig()
	d.Net = "tcp"
	d.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	d.User = adminUser
	d.Passwd = password
	d.Timeout = 5 * time.Second
	db, e := sql.Open("mysql", d.FormatDSN())
	if e != nil {
		return fmt.Errorf("falha ao configurar MariaDB")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if e = db.PingContext(ctx); e != nil {
		return fmt.Errorf("falha ao conectar; confira host, porta e credencial administrativa")
	}
	var n int
	if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME IN ('radar_literario','radar_literario_test')").Scan(&n); e != nil {
		return fmt.Errorf("não foi possível inspecionar os bancos")
	}
	if n > 0 {
		return fmt.Errorf("um dos bancos já existe; configure credenciais locais existentes seguindo o README, sem recriar dados")
	}
	if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.user WHERE User IN ('radar_app','radar_migrate','radar_test')").Scan(&n); e != nil {
		return fmt.Errorf("não foi possível inspecionar usuários administrativos")
	}
	if n > 0 {
		return fmt.Errorf("usuários radar já existem; revise-os manualmente para preservar suas credenciais")
	}
	keyBytes := make([]byte, 32)
	if _, e = rand.Read(keyBytes); e != nil {
		return e
	}
	key := base64.StdEncoding.EncodeToString(keyBytes)
	app, e := secret(30)
	if e != nil {
		return e
	}
	migration, e := secret(30)
	if e != nil {
		return e
	}
	test, e := secret(30)
	if e != nil {
		return e
	}
	env := func(name, user, pw string) string {
		return fmt.Sprintf("APP_ENV=development\nHTTP_HOST=127.0.0.1\nHTTP_PORT=8080\nDB_HOST=%s\nDB_PORT=%d\nDB_NAME=%s\nDB_USER=%s\nDB_PASSWORD=%s\nOUTBOX_ENCRYPTION_KEY=%s\nCORS_ALLOWED_ORIGINS=http://localhost:5173\nCSRF_ALLOWED_ORIGINS=http://localhost:5173\nSESSION_COOKIE_NAME=radar_session\nSESSION_IDLE_TIMEOUT=30m\nSESSION_ABSOLUTE_TIMEOUT=168h\nCOOKIE_SECURE=false\nCOOKIE_SAME_SITE=lax\nSMTP_HOST=localhost\nSMTP_PORT=1025\nSMTP_FROM=nao-responda@example.test\nPUBLIC_FRONTEND_URL=http://localhost:5173\n", host, port, name, user, pw, key)
	}
	files := map[string]string{".env": env("radar_literario", "radar_app", app), ".env.migrate": env("radar_literario", "radar_migrate", migration), ".env.test": strings.Replace(env("radar_literario_test", "radar_test", test), "APP_ENV=development", "APP_ENV=test", 1)}
	td := mysql.NewConfig()
	td.Net = "tcp"
	td.Addr = d.Addr
	td.User = "radar_test"
	td.Passwd = test
	td.DBName = "radar_literario_test"
	td.ParseTime = true
	td.Loc = time.UTC
	files[".env.test"] += "TEST_DB_DSN=" + td.FormatDSN() + "\n"
	for _, name := range []string{".env", ".env.migrate", ".env.test"} {
		f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		_, e = f.WriteString(files[name])
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	queries := []string{"CREATE DATABASE radar_literario CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "CREATE DATABASE radar_literario_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"}
	for _, q := range queries {
		if _, e = db.ExecContext(ctx, q); e != nil {
			return fmt.Errorf("falha ao criar banco; preserve os arquivos .env gerados para revisão")
		}
	}
	for _, a := range []struct{ name, pw, database, grants string }{{"radar_app", app, "radar_literario", "SELECT,INSERT,UPDATE,DELETE"}, {"radar_migrate", migration, "radar_literario", "ALL PRIVILEGES"}, {"radar_test", test, "radar_literario_test", "ALL PRIVILEGES"}} {
		// MariaDB does not accept a prepared placeholder in IDENTIFIED BY.
		// Names/hosts are fixed or allowlisted and passwords are generated base64url.
		if _, e = db.ExecContext(ctx, "CREATE USER '"+a.name+"'@'"+accountHost+"' IDENTIFIED BY '"+a.pw+"'"); e != nil {
			return fmt.Errorf("falha ao criar usuário %s; credencial permanece nos arquivos locais", a.name)
		}
		if _, e = db.ExecContext(ctx, "GRANT "+a.grants+" ON "+a.database+".* TO '"+a.name+"'@'"+accountHost+"'"); e != nil {
			return fmt.Errorf("falha ao conceder permissões para %s", a.name)
		}
	}
	fmt.Println("Bancos e usuários locais criados. Credenciais gravadas somente em .env, .env.migrate e .env.test. Aplique as migrações seguindo o README.")
	return nil
}
