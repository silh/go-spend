package expenses_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"math/rand"
	"testing"
	"time"
)

const (
	pizzaPrice  = 44
	coffeePrice = 8
)

func TestPgBalanceRepositoryGetBalanceManyEntries(t *testing.T) {
	t.Skip("small test to check how long it takes to fetch")
	// it actually takes a lot of time to insert...
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)
	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()
	expensesRepository := expenses.NewPgRepository()
	balanceRepository := expenses.NewPgBalanceRepository()

	// Create user and group
	user1 := createProperUser(ctx, t, "1", userRepository)
	user2 := createProperUser(ctx, t, "2", userRepository)
	user3 := createProperUser(ctx, t, "3", userRepository)
	user4 := createProperUser(ctx, t, "4", userRepository)
	group1 := createGroup(ctx, t, groupRepository, "1")
	group2 := createGroup(ctx, t, groupRepository, "2")
	addToGroup(ctx, t, groupRepository, group1.ID, user1, user2, user3)
	addToGroup(ctx, t, groupRepository, group2.ID, user4)
	prepareNExpenses(t, expensesRepository, ctx, 1_000)
	now := time.Now()
	_, err := balanceRepository.Get(ctx, pgdb, user1.ID)
	require.NoError(t, err)
	spent := time.Now().Sub(now)
	fmt.Printf("%++v\n", spent)
}

func TestPgBalanceRepositoryGetBalance(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)
	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()
	expensesRepository := expenses.NewPgRepository()
	balanceRepository := expenses.NewPgBalanceRepository()

	// Create user and group
	user1 := createProperUser(ctx, t, "1", userRepository)
	user2 := createProperUser(ctx, t, "2", userRepository)
	user3 := createProperUser(ctx, t, "3", userRepository)
	user4 := createProperUser(ctx, t, "4", userRepository)
	group1 := createGroup(ctx, t, groupRepository, "1")
	group2 := createGroup(ctx, t, groupRepository, "2")
	addToGroup(ctx, t, groupRepository, group1.ID, user1, user2, user3)
	addToGroup(ctx, t, groupRepository, group2.ID, user4)
	payForPizza(t, expensesRepository, ctx, user1.ID, user2.ID, user3.ID)
	payForCoffee(t, expensesRepository, ctx, user2.ID, user1.ID)
	balance1, err := balanceRepository.Get(ctx, pgdb, user1.ID)
	require.NoError(t, err)
	user1expectedBalance := expenses.Balance{ // well it's not exact because we have uint percentages
		2: pizzaPrice*0.33 - coffeePrice*0.5,
		3: pizzaPrice * 0.34,
	}
	assert.InDelta(t, user1expectedBalance[2], balance1[2], 0.01)
	assert.InDelta(t, user1expectedBalance[3], balance1[3], 0.01)
	balance2, err := balanceRepository.Get(ctx, pgdb, user2.ID)
	user2expectedBalance := expenses.Balance{
		1: -pizzaPrice*0.33 + coffeePrice*0.5,
	}
	assert.InDelta(t, user2expectedBalance[1], balance2[1], 0.01)
	assert.Zero(t, user2expectedBalance[3])
	user3expectedBalance := expenses.Balance{
		1: -pizzaPrice * 0.34,
	}
	balance3, err := balanceRepository.Get(ctx, pgdb, user3.ID)
	require.NoError(t, err)
	assert.InDelta(t, user3expectedBalance[1], balance3[1], 0.01)
}

func prepareNExpenses(t *testing.T, expensesRepository *expenses.PgRepository, ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		userID := uint(rand.Intn(2)) + 1
		amount := uint(rand.Intn(500)) + 1
		req := expenses.NewExpense{
			UserID: userID,
			Amount: float32(amount),
		}
		expense, err := expensesRepository.Create(ctx, pgdb, req)
		require.NoError(t, err)
		var createExpenseShares expenses.CreateExpenseShares
		if userID == 4 {
			createExpenseShares = expenses.CreateExpenseShares{
				ExpenseID: expense.ID,
				Shares: expenses.ExpenseShares{
					userID: 100,
				},
			}
		} else {
			shares := make(expenses.ExpenseShares)
			numberOfParts := rand.Intn(2) + 1
			left := 100
			var part uint
			for part = 1; part <= uint(numberOfParts) && left > 1; part++ {
				randPercent := rand.Intn(left) + 1
				left -= randPercent
				shares[part] = expenses.Percent(randPercent)
			}
			if left > 0 {
				shares[part] = expenses.Percent(left)
			}
			createExpenseShares = expenses.CreateExpenseShares{
				ExpenseID: expense.ID,
				Shares:    shares,
			}
		}
		err = expensesRepository.CreateShares(ctx, pgdb, createExpenseShares)
		require.NoError(t, err)
	}
}

func payForPizza(
	t *testing.T,
	expensesRepository *expenses.PgRepository,
	ctx context.Context,
	user1, user2, user3 uint,
) {
	req := expenses.NewExpense{
		UserID: user1,
		Amount: pizzaPrice,
	}
	expense, err := expensesRepository.Create(ctx, pgdb, req)
	require.NoError(t, err)
	createExpenseShares := expenses.CreateExpenseShares{
		ExpenseID: expense.ID,
		Shares: expenses.ExpenseShares{
			user1: 33,
			user2: 33,
			user3: 34,
		},
	}
	err = expensesRepository.CreateShares(ctx, pgdb, createExpenseShares)
	require.NoError(t, err)
}

func payForCoffee(
	t *testing.T,
	expensesRepository *expenses.PgRepository,
	ctx context.Context,
	user1, user2 uint,
) {
	req := expenses.NewExpense{
		UserID: user1,
		Amount: coffeePrice,
	}
	expense, err := expensesRepository.Create(ctx, pgdb, req)
	require.NoError(t, err)
	createExpenseShares := expenses.CreateExpenseShares{
		ExpenseID: expense.ID,
		Shares: expenses.ExpenseShares{
			user1: 50,
			user2: 50,
		},
	}
	err = expensesRepository.CreateShares(ctx, pgdb, createExpenseShares)
	require.NoError(t, err)
}
