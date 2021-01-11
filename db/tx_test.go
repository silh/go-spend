package db_test

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/db"
	"testing"
)

type mockTxProvider struct {
	mock.Mock
}

func (m *mockTxProvider) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).(pgx.Tx), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockTx struct {
	mock.Mock
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	panic("implement me")
}

func (m *mockTx) Commit(ctx context.Context) error {
	panic("implement me")
}

func (m *mockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	panic("implement me")
}

func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic("implement me")
}

func (m *mockTx) LargeObjects() pgx.LargeObjects {
	panic("implement me")
}

func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	panic("implement me")
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	panic("implement me")
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	panic("implement me")
}

func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	panic("implement me")
}

func (m *mockTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	panic("implement me")
}

func (m *mockTx) Conn() *pgx.Conn {
	panic("implement me")
}

func TestWithTxPanic(t *testing.T) {
	defer func() {
		err := recover()
		require.NotNil(t, err)
	}()
	txProvider := new(mockTxProvider)
	tx := new(mockTx)
	ctx := context.Background()
	txProvider.On("Begin", ctx).Return(tx, nil)
	tx.On("Rollback", ctx).Return(nil)
	db.WithTx(ctx, txProvider, func(tx pgxtype.Querier) error {
		panic(errors.New("expected"))
	})
	t.Fail() // should not get here
}
