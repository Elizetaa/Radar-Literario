package interfaces

import "context"

// Store is the transaction-bound persistence gateway used by concrete repositories.
// SQL belongs to repository implementations; application services use domain interfaces.
type Store interface {
	Query(context.Context, any, string, ...any) error
	Exec(context.Context, string, ...any) (int64, error)
	Insert(context.Context, string, map[string]any) (uint64, error)
	Transaction(context.Context, func(Store) error) error
	Ready(context.Context) error
}
