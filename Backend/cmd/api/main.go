package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
	"radar-literario/backend/internal/router"
	"syscall"
	"time"
)

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e = database.SchemaReady(ctx, pool); e != nil {
		slog.Warn("banco exige aplicação de migrações; /ready retornará 503")
	}
	store := tools.NewStore(db)
	accounts := services.NewAccounts(tools.NewAccountRepository(store), cfg, tools.NewSMTPMailer(cfg))
	catalog := services.NewCatalog(tools.NewCatalogRepository(store), cfg.MinimumReviews)
	reading := services.NewReading(tools.NewReadingRepository(store), services.ReadingOptions{MaxCommentDepth: cfg.MaxCommentDepth, BlockedTerms: cfg.BlockedTerms, BlockedDomains: cfg.BlockedDomains, Recommendations: catalog.Recommendations, PreferredWorks: catalog.Preferred})
	admin := services.NewAdmin(tools.NewAdminRepository(store))
	handler := router.New(cfg, router.Dependencies{Accounts: accounts, Catalog: catalog, Reading: reading, Admin: admin, Ready: func(c context.Context) error { return database.SchemaReady(c, pool) }})
	server := &http.Server{Addr: cfg.Address(), Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: cfg.RequestTimeout, WriteTimeout: cfg.RequestTimeout + time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if e := accounts.DispatchMail(ctx); e != nil && ctx.Err() == nil {
					slog.Warn("processamento da fila de e-mail indisponível")
				}
			}
		}
	}()
	listener, e := net.Listen("tcp", cfg.Address())
	if e != nil {
		stop()
		<-workerDone
		return e
	}
	slog.Info("API iniciada", "address", listener.Addr().String(), "environment", cfg.Env)
	e = serve(ctx, server, listener)
	stop()
	<-workerDone
	if e == nil {
		slog.Info("API encerrada; requisições concluídas e recursos liberados")
	}
	return e
}
func serve(ctx context.Context, server *http.Server, listener net.Listener) error {
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(listener) }()
	select {
	case <-ctx.Done():
	case e := <-errCh:
		if e != http.ErrServerClosed {
			return e
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := server.Shutdown(shutdown); e != nil {
		_ = server.Close()
		return e
	}
	return nil
}
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if e := run(); e != nil {
		slog.Error("não foi possível executar API", "message", e.Error())
		os.Exit(1)
	}
}
