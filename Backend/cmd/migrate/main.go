package main

import (
	"context"
	"flag"
	"log"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
)

func main() {
	command := flag.String("command", "status", "up, status, version, down")
	flag.Parse()
	if e := run(*command); e != nil {
		log.Print(e)
		os.Exit(1)
	}
}
func run(command string) error {
	cfg, e := config.Load()
	if e != nil {
		return e
	}
	_, pool, e := database.Open(cfg)
	if e != nil {
		return e
	}
	defer pool.Close()
	return database.Migration(context.Background(), pool, command)
}
