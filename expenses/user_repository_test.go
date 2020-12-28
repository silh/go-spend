package expenses_test

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"strings"
	"testing"
)

const (
	pgImage    = "postgres:13.1"
	pgUser     = "user"
	pgPassword = "password"
	pgDb       = "expenses"
	pgPort     = nat.Port("5432/tcp")

	deleteAllUsersQuery = "DELETE FROM users"
)

var pgDB = createContainerAndGetDbUrl(context.Background())

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	created, err := expenses.NewPgUserRepository().Create(ctx, pgDB, user)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
}

func TestCantCreateTwoUsersWithSameEmail(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	repository := expenses.NewPgUserRepository()
	_, _ = repository.Create(ctx, pgDB, user)
	created2, err := repository.Create(ctx, pgDB, user)
	assert.Zero(t, created2)
	assert.EqualError(t, err, expenses.ErrEmailAlreadyExists.Error())
}

func TestCantStoreTooLongEmail(t *testing.T) {
	// this should not happen in real application as an email should be validated, added to check the constraint
	ctx := context.Background()
	cleanUpDB(t, ctx)

	user := expenses.CreateUserRequest{Email: createLongEmail(), Password: "password"}
	repository := expenses.NewPgUserRepository()
	created, err := repository.Create(ctx, pgDB, user)
	assert.Zero(t, created)
	assert.Error(t, err)
}

func TestFindById(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	// Create user to retrieve it later
	repository := expenses.NewPgUserRepository()
	user := expenses.CreateUserRequest{Email: "expenses@mail.com", Password: "password"}
	created, err := repository.Create(ctx, pgDB, user)
	require.NoError(t, err)

	foundUser, err := repository.FindById(ctx, pgDB, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, foundUser)
}

func TestFindByIdNonExistentUser(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	// Create user to retrieve it later
	repository := expenses.NewPgUserRepository()
	foundUser, err := repository.FindById(ctx, pgDB, 1)
	assert.EqualError(t, err, expenses.ErrUserNotFound.Error())
	assert.Zero(t, foundUser)
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
