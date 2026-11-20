package tools

import (
	"context"
	"errors"
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"regexp"
	"sort"
	"strings"
)

type store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) interfaces.Store { return &store{db: db} }
func persistenceError(e error) error {
	if e == nil {
		return nil
	}
	var m *mysql.MySQLError
	if errors.As(e, &m) {
		switch m.Number {
		case 1020, 1062, 1213, 1205:
			return core.Conflict("A operação conflita com outro registro ou alteração. Atualize e tente novamente.")
		case 1451:
			return core.Conflict("O registro possui vínculos e não pode ser removido.")
		case 1452, 3819, 4025, 1265, 1406, 1048:
			return core.Invalid("INVALID_RELATION", "Os dados violam uma restrição de integridade.")
		}
	}
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return core.NotFound()
	}
	return e
}
func (s *store) Query(ctx context.Context, dest any, q string, args ...any) error {
	return persistenceError(s.db.WithContext(ctx).Raw(q, args...).Scan(dest).Error)
}
func (s *store) Exec(ctx context.Context, q string, args ...any) (int64, error) {
	r := s.db.WithContext(ctx).Exec(q, args...)
	return r.RowsAffected, persistenceError(r.Error)
}

var identifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (s *store) Insert(ctx context.Context, table string, values map[string]any) (uint64, error) {
	if !identifier.MatchString(table) || len(values) == 0 {
		return 0, fmt.Errorf("invalid repository insert")
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if !identifier.MatchString(k) {
			return 0, fmt.Errorf("invalid repository column")
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	cols := make([]string, len(keys))
	binds := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		cols[i] = "`" + k + "`"
		binds[i] = "?"
		args[i] = values[k]
	}
	var id uint64
	e := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := tx.Exec("INSERT INTO `"+table+"` ("+strings.Join(cols, ",")+") VALUES ("+strings.Join(binds, ",")+")", args...)
		if r.Error != nil {
			return r.Error
		}
		return tx.Raw("SELECT LAST_INSERT_ID()").Scan(&id).Error
	})
	return id, persistenceError(e)
}
func (s *store) Transaction(ctx context.Context, fn func(interfaces.Store) error) error {
	return persistenceError(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(&store{db: tx}) }))
}
func (s *store) Ready(ctx context.Context) error {
	d, e := s.db.DB()
	if e != nil {
		return e
	}
	return d.PingContext(ctx)
}
