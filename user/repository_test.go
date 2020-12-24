package user_test

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go-spend/user"
	"io/ioutil"
	"testing"
)

const (
	pgImage    = "postgres:13.1"
	pgUser     = "user"
	pgPassword = "password"
	pgDb       = "expenses"
	pgPort     = nat.Port("5432/tcp")
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
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
	require.NoError(t, err)
	endpoint, err := postgres.Endpoint(ctx, "")
	require.NoError(t, err)

	url := fmt.Sprintf("postgresql://%s:%s@%s/%s", pgUser, pgPassword, endpoint, pgDb)
	db, err := pgxpool.Connect(ctx, url)
	require.NoError(t, err)
	schema, err := ioutil.ReadFile("../db/001_schema.sql")
	require.NoError(t, err)
	_, err = db.Exec(ctx, string(schema))
	require.NoError(t, err)

	_, err = user.NewRepositoryImpl(db).Create(user.User{ID: 1, Email: "user@mail.com", Password: "password"})
	require.NoError(t, err)
}
