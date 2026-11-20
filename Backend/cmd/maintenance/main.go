package main

import (
	"context"
	"log"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

func main() {
	if e := run(); e != nil {
		log.Print(e)
		os.Exit(1)
	}
}
func run() error {
	cfg, e := config.Load()
	if e != nil {
		return e
	}
	db, pool, e := database.Open(cfg)
	if e != nil {
		return e
	}
	defer pool.Close()
	return services.NewAccounts(tools.NewAccountRepository(tools.NewStore(db)), cfg, tools.NewSMTPMailer(cfg)).Purge(context.Background())
}
