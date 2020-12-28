package expenses_test

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go-spend/expenses"
	"go-spend/util"
	"io/ioutil"
	"strings"
	"testing"
)

const (
	pgImage    = "postgres:13.1"
	pgUser     = "user"
	pgPassword = "password"
	pgDb       = "expenses"
	pgPort     = nat.Port("5432/tcp")

	deleteAllUsersQuery = "DELETE FROM USERS"
)

var db = createContainerAndGetDbUrl(context.Background())

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	cleanUpUsers(t, ctx)

	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	created, err := expenses.NewPgRepository(db).Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
}

func TestCantCreateTwoUsersWithSameEmail(t *testing.T) {
	ctx := context.Background()
	cleanUpUsers(t, ctx)

	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	repository := expenses.NewPgRepository(db)
	_, _ = repository.Create(ctx, user)
	created2, err := repository.Create(ctx, user)
	assert.Zero(t, created2)
	assert.EqualError(t, err, expenses.ErrEmailAlreadyExists.Error())
}

func TestCantStoreTooLongEmail(t *testing.T) {
	// this should not happen in real application as an email should be validated, added to check the constraint
	ctx := context.Background()
	cleanUpUsers(t, ctx)

	user := expenses.CreateUserRequest{Email: createLongEmail(), Password: "password"}
	repository := expenses.NewPgRepository(db)
	created, err := repository.Create(ctx, user)
	assert.Zero(t, created)
	assert.Error(t, err)
}

func TestFindById(t *testing.T) {
	ctx := context.Background()
	cleanUpUsers(t, ctx)

	// Create user to retrieve it later
	repository := expenses.NewPgRepository(db)
	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	created, err := repository.Create(ctx, user)
	require.NoError(t, err)

	foundUser, err := repository.FindById(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, foundUser)
}

func cleanUpUsers(t *testing.T, ctx context.Context) {
	rows, _ := db.Query(ctx, deleteAllUsersQuery)
	defer rows.Close()
	require.NoError(t, rows.Err())
}

func createLongEmail() util.Email {
	suffix := "@email.com"
	builder := strings.Builder{}
	for i := 0; i < 321-len(suffix); i++ {
		builder.WriteRune('c')
	}
	builder.WriteString(suffix)
	return util.Email(builder.String())
}

// Creates PG container, applies necessary schema. If there is any error - it will panic
func createContainerAndGetDbUrl(ctx context.Context) pgxtype.Querier {
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
