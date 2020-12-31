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
