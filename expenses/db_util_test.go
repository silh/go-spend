package expenses_test

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io/ioutil"
	"testing"
)

const (
	pgImage    = "postgres:13.1"
	pgUser     = "user"
	pgPassword = "password"
	pgDb       = "expenses"
	pgPort     = nat.Port("5432/tcp")

	deleteAllUsersQuery  = "DELETE FROM users"
	deleteAllGroupsQuery = "DELETE FROM groups"
)

var PGDB = CreateContainerAndGetDbUrl(context.Background())

// Creates PG container, applies necessary schema. If there is any error - it will panic
func CreateContainerAndGetDbUrl(ctx context.Context) *pgxpool.Pool {
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        pgImage,
			ExposedPorts: []string{string(pgPort)},
			Env: map[string]string{
				"POSTGRES_USER":     pgUser,
				"POSTGRES_PASSWORD": pgPassword,
				"POSTGRES_DB":       pgDb,
			},
			WaitingFor: wait.ForAll(
				wait.NewLogStrategy("database system is ready to accept connections").WithOccurrence(2),
				wait.ForListeningPort(pgPort),
			),
			Cmd: []string{"postgres", "-c", "fsync=off"},
		},
		Started: true,
	})
	if err != nil {
		panic(err)
	}
	endpoint, err := postgres.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("postgresql://%s:%s@%s/%s", pgUser, pgPassword, endpoint, pgDb)

	db, err := pgxpool.Connect(ctx, url)
	if err != nil {
		panic(err)
	}
	schema, err := ioutil.ReadFile("../db/001_schema.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(ctx, string(schema))
	if err != nil {
		panic(err)
	}
	return db
}

func cleanUpDB(t *testing.T, ctx context.Context) {
	_, err := PGDB.Exec(ctx, deleteAllGroupsQuery)
	require.NoError(t, err)
	_, err = PGDB.Exec(ctx, deleteAllUsersQuery)
	require.NoError(t, err)
}

type MockTxQuerier struct {
	mock.Mock
}

func (m *MockTxQuerier) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	panic("implement me")
}

func (m *MockTxQuerier) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	panic("implement me")
}

func (m *MockTxQuerier) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	panic("implement me")
}

func (m *MockTxQuerier) Begin(_ context.Context) (pgx.Tx, error) {
	panic("implement me")
}
