package router_test

import (
	"encoding/json"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/openapi"
	"radar-literario/backend/internal/router"
	"testing"
)

func TestOpenAPIContractMatchesRoutesAndDTOs(t *testing.T) {
	engine := router.New(config.Config{CookieName: "radar_session"}, router.Dependencies{})
	data, err := openapi.Build(engine.Routes(), "radar_session")
	if err != nil {
		t.Fatal(err)
	}
	versioned, err := os.ReadFile("../../../Docs/api.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(versioned) != string(data) {
		t.Fatal("Docs/api.yaml desatualizado: execute go run ./cmd/openapi em Backend")
	}
	if err = openapi.Validate(versioned); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	components := doc["components"].(map[string]any)["schemas"].(map[string]any)
	for _, name := range []string{"DiaryInput", "ReviewInput", "WorkCreate", "EditionCreate", "CreateReport", "ErrorEnvelope"} {
		if _, ok := components[name]; !ok {
			t.Errorf("missing schema %s", name)
		}
	}
	paths := doc["paths"].(map[string]any)
	write := paths["/api/v1/reading-entries"].(map[string]any)["post"].(map[string]any)
	requirements := write["security"].([]any)[0].(map[string]any)
	if _, ok := requirements["SessionAuth"]; !ok {
		t.Fatal("protected write missing session requirement")
	}
	if _, ok := requirements["CSRF"]; !ok {
		t.Fatal("protected write missing CSRF requirement")
	}
}
