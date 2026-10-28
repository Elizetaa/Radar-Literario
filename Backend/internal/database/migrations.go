package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
	"radar-literario/backend/migrations"
)

func Migration(ctx context.Context, db *sql.DB, command string) error {
	goose.SetBaseFS(migrations.FS)
	if e := goose.SetDialect("mysql"); e != nil {
		return e
	}
	switch command {
	case "up":
		return goose.UpContext(ctx, db, ".")
	case "status":
		return goose.StatusContext(ctx, db, ".")
	case "version":
		v, e := goose.GetDBVersionContext(ctx, db)
		if e == nil {
			fmt.Println(v)
		}
		return e
	case "down":
		return goose.DownContext(ctx, db, ".")
	default:
		return fmt.Errorf("comando deve ser up, status, version ou down")
	}
}
func SchemaReady(ctx context.Context, db *sql.DB) error {
	if e := db.PingContext(ctx); e != nil {
		return e
	}
	var v int64
	if e := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version_id),0) FROM goose_db_version WHERE is_applied=1").Scan(&v); e != nil {
		return e
	}
	if v != migrations.Latest {
		return fmt.Errorf("migrações pendentes")
	}
	return nil
}
