package main_test

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

const (
	pgImage    = "postgres:13.1"
	pgUser     = "user"
	pgPassword = "password"
	pgDb       = "expenses"
	pgPort     = "5432/tcp"

	deleteAllUsersQuery  = "DELETE FROM users"
	deleteAllGroupsQuery = "DELETE FROM groups"

	redisImage    = "redis:6.0.9-alpine3.12"
	redisPassword = "password"
)

// Creates Redis container and return its address
func createRedisContainer(ctx context.Context) string {
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        redisImage,
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
			Cmd:          []string{"redis-server", "--requirepass", redisPassword},
		},
		Started: true,
	})
	if err != nil {
		panic(err)
	}
	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}
	return endpoint
}

// Creates PG container, applies necessary schema. If there is any error - it will panic
func createPGContainerAndGetDbUrl(ctx context.Context) string {
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

	return fmt.Sprintf("postgresql://%s:%s@%s/%s", pgUser, pgPassword, endpoint, pgDb)
}

func cleanUpDB(t *testing.T, ctx context.Context) {
	pgdb, err := pgxpool.Connect(ctx, defaultConfig.DB.ConnectionString)
	if err != nil {
		panic(err)
	}
	_, err = pgdb.Exec(ctx, deleteAllGroupsQuery)
	require.NoError(t, err)
	_, err = pgdb.Exec(ctx, deleteAllUsersQuery)
	require.NoError(t, err)
}
