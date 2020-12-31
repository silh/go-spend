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

func TestPgRepositoryCreateShares(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repo := new(expenses.PgRepository)
	userRepo := new(expenses.PgUserRepository)

	// Need to create a users first
	user1, err := userRepo.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "mail@mail.com", Password: "128c76xz"})
	require.NoError(t, err)
	user2, err := userRepo.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "mail2@mail.com", Password: "128c76xz"})
	require.NoError(t, err)
	// And expense
	req := expenses.NewExpense{
		UserID: user1.ID,
		Amount: 20.20,
	}
	createdExpense, err := repo.Create(ctx, pgdb, req)
	require.NoError(t, err)

	// when
	createExpenseShares := expenses.CreateExpenseShares{
		ExpenseID: createdExpense.ID,
		Shares: expenses.ExpenseShares{
			user1.ID: 10,
			user2.ID: 90,
		},
	}
	err = repo.CreateShares(ctx, pgdb, createExpenseShares)

	// then
	require.NoError(t, err)
}

func TestPgRepositoryCreateSharesOneUserDoesntExist(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repo := new(expenses.PgRepository)
	userRepo := new(expenses.PgUserRepository)

	// Need to create a users first
	user1, err := userRepo.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "mail@mail.com", Password: "128c76xz"})
	require.NoError(t, err)
	// And expense
	req := expenses.NewExpense{
		UserID: user1.ID,
		Amount: 20.20,
	}
	createdExpense, err := repo.Create(ctx, pgdb, req)
	require.NoError(t, err)

	// when
	createExpenseShares := expenses.CreateExpenseShares{
		ExpenseID: createdExpense.ID,
		Shares: expenses.ExpenseShares{
			user1.ID: 10,
			99:       90,
		},
	}
	err = repo.CreateShares(ctx, pgdb, createExpenseShares)

	// then
	require.EqualError(t, err, expenses.ErrUserOrExpenseDoesntExist.Error())
}

func TestPgRepositoryCreateSharesExpenseDoesntExist(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repo := new(expenses.PgRepository)
	userRepo := new(expenses.PgUserRepository)

	// Need to create a users first
	user1, err := userRepo.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "mail@mail.com", Password: "128c76xz"})
	require.NoError(t, err)

	// when
	createExpenseShares := expenses.CreateExpenseShares{
		ExpenseID: 10,
		Shares: expenses.ExpenseShares{
			user1.ID: 100,
		},
	}
	err = repo.CreateShares(ctx, pgdb, createExpenseShares)

	// then
	require.EqualError(t, err, expenses.ErrUserOrExpenseDoesntExist.Error())
}
