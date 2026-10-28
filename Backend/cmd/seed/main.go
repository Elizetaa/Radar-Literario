package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

func ptr[T any](v T) *T { return &v }
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
	if cfg.Env != "development" && cfg.Env != "test" {
		return fmt.Errorf("seed fictício permitido apenas em development/test")
	}
	db, pool, e := database.Open(cfg)
	if e != nil {
		return e
	}
	defer pool.Close()
	ctx := context.Background()
	store := tools.NewStore(db)
	var adminID uint64
	if e = store.Query(ctx, &adminID, "SELECT a.user_id FROM administrators a JOIN users u ON u.id=a.user_id WHERE a.role='system_administrator' AND a.status='active' AND u.status='active' ORDER BY a.user_id LIMIT 1"); e != nil {
		return e
	}
	if adminID == 0 {
		return fmt.Errorf("crie o primeiro administrador antes do seed")
	}
	s := services.NewCatalog(tools.NewCatalogRepository(store), 1)
	actor := core.Actor{ID: adminID, Role: "system_administrator", Reader: true}
	var workID uint64
	if e = store.Query(ctx, &workID, "SELECT id FROM works WHERE title=? ORDER BY id LIMIT 1", "Viagem ao Farol de Papel (demonstração)"); e != nil {
		return e
	}
	if workID != 0 {
		fmt.Printf("Seed já disponível, obra id=%d.\n", workID)
		return nil
	}
	var authorID, publisherID, genreID uint64
	if e = store.Query(ctx, &authorID, "SELECT id FROM authors WHERE name=? LIMIT 1", "Autora Fictícia Radar"); e != nil {
		return e
	}
	if authorID == 0 {
		v, e := s.SaveAuthor(ctx, actor, 0, schemas.AuthorInput{Name: ptr("Autora Fictícia Radar"), Biography: ptr("Personagem fictícia usada apenas para demonstrar o catálogo.")})
		if e != nil {
			return e
		}
		authorID = v.ID
	}
	if e = store.Query(ctx, &publisherID, "SELECT id FROM publishers WHERE name=? LIMIT 1", "Editora Fictícia Radar"); e != nil {
		return e
	}
	if publisherID == 0 {
		v, e := s.SavePublisher(ctx, actor, 0, schemas.PublisherInput{Name: ptr("Editora Fictícia Radar")})
		if e != nil {
			return e
		}
		publisherID = v.ID
	}
	if e = store.Query(ctx, &genreID, "SELECT id FROM genres WHERE name=? LIMIT 1", "Ficção de demonstração"); e != nil {
		return e
	}
	if genreID == 0 {
		v, e := s.SaveGenre(ctx, actor, 0, schemas.GenreInput{Name: ptr("Ficção de demonstração"), Active: ptr(true)})
		if e != nil {
			return e
		}
		genreID = v.ID
	}
	v, e := s.SaveWork(ctx, actor, 0, schemas.WorkInput{Title: ptr("Viagem ao Farol de Papel (demonstração)"), Synopsis: ptr("Uma história inteiramente fictícia para explorar os fluxos do Radar Literário."), OriginalLanguage: ptr("pt-BR"), Status: ptr("published"), MinimumAge: ptr(0), AuthorIDs: ptr([]uint64{authorID}), GenreIDs: ptr([]uint64{genreID}), InitialEdition: &schemas.EditionInput{PublisherID: &publisherID, Title: ptr("Edição de demonstração"), ISBNAbsenceReason: ptr("Obra fictícia de demonstração, sem ISBN comercial."), EditionNumber: ptr(1), PageCount: ptr(180), Language: ptr("pt-BR"), Status: ptr("published"), MinimumAge: ptr(0)}})
	if e != nil {
		return e
	}
	fmt.Printf("Seed fictício disponível, obra id=%d.\n", v.ID)
	return nil
}
