package main

import (
	"flag"
	"fmt"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/openapi"
	"radar-literario/backend/internal/router"
)

func main() {
	output := flag.String("out", "../Docs/api.yaml", "Destino do contrato OpenAPI (JSON também é YAML 1.2).")
	check := flag.Bool("check", false, "Compara contrato versionado sem modificá-lo.")
	flag.Parse()
	engine := router.New(config.Config{CookieName: "radar_session"}, router.Dependencies{})
	data, err := openapi.Build(engine.Routes(), "radar_session")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *check {
		existing, err := os.ReadFile(*output)
		if err != nil || string(existing) != string(data) {
			fmt.Fprintln(os.Stderr, "Contrato desatualizado; execute go run ./cmd/openapi.")
			os.Exit(1)
		}
		fmt.Println("OpenAPI válido e correspondente às rotas e DTOs.")
		return
	}
	if err = os.WriteFile(*output, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Contrato OpenAPI gerado em", *output)
}
