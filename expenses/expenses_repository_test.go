package expenses_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

func TestPgRepositoryCreate(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repo := new(expenses.PgRepository)
	userRepo := new(expenses.PgUserRepository)

	// Need to create a user first
	user, err := userRepo.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "mail@mail.com", Password: "128c76xz"})
	require.NoError(t, err)
	req := expenses.NewExpense{
		UserID: user.ID,
		Amount: 20.20,
	}

	// when
	createdExpense, err := repo.Create(ctx, pgdb, req)

	// then
	require.NoError(t, err)
	assert.NotZero(t, createdExpense)
	assert.NotZero(t, createdExpense.ID)
	assert.NotZero(t, createdExpense.Timestamp)
	assert.Equal(t, req.UserID, createdExpense.UserID)
	assert.Equal(t, req.Amount, createdExpense.Amount)
}
