package db

import (
	"context"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
)

// Combines pgxtype.Querier and TxProvider. In app usually represented as DB connection or connection pool
type TxQuerier interface {
	pgxtype.Querier
	TxProvider
}

// An interface abstraction to enable mocking of TX creation
type TxProvider interface {
	// Begin a new transaction
	Begin(ctx context.Context) (pgx.Tx, error)
}

const (
	// PG Error codes
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
)

// TxFunc is any function that should be executed in transaction context
type TxFunc func(tx pgxtype.Querier) error

// WithTx executes provided function in a transaction using a given TxProvider instance to start and commit/rollback
// transaction
func WithTx(ctx context.Context, db TxProvider, f TxFunc) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if panicErr := recover(); panicErr != nil {
			tx.Rollback(ctx)
			panic(panicErr)
		}
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()
	return f(tx)
}
