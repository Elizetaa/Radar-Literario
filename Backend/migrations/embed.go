package migrations

import "embed"

// FS includes the exact versioned SQL files shipped with the application.
//
//go:embed *.sql
var FS embed.FS

const Latest int64 = 4
