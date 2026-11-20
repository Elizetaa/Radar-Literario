package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

func main() {
	weak := flag.Bool("allow-weak-local-password", false, "exceção explícita de senha somente em development")
	flag.Parse()
	if e := run(*weak); e != nil {
		log.Print(e)
		os.Exit(1)
	}
}
func run(weak bool) error {
	cfg, e := config.Load()
	if e != nil {
		return e
	}
	db, pool, e := database.Open(cfg)
	if e != nil {
		return e
	}
	defer pool.Close()
	in := schemas.RegisterReader{Name: os.Getenv("ADMIN_NAME"), Nickname: os.Getenv("ADMIN_NICKNAME"), Email: os.Getenv("ADMIN_EMAIL"), Password: os.Getenv("ADMIN_PASSWORD"), BirthDate: os.Getenv("ADMIN_BIRTH_DATE")}
	accounts := services.NewAccounts(tools.NewAccountRepository(tools.NewStore(db)), cfg, tools.NewSMTPMailer(cfg))
	id, e := accounts.Bootstrap(context.Background(), in, weak)
	if e == nil {
		fmt.Printf("Administrador de sistema disponível, id=%d.\n", id)
	}
	return e
}
