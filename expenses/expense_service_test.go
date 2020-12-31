package expenses_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"testing"
)

// integration tests
func TestDefaultServiceCreateExpense(t *testing.T) {
	// given
	ctx := context.Background()
	cleanUpDB(t, ctx)
	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	expensesService := expenses.NewDefaultService(pgdb, groupRepository, expenses.NewPgRepository())

	// Create user and group
	user1 := createProperUser(ctx, t, "1", userRepository)
	user2 := createProperUser(ctx, t, "2", userRepository)
	user3 := createProperUser(ctx, t, "3", userRepository)
	group := createGroup(ctx, t, groupRepository, "1")
	group2 := createGroup(ctx, t, groupRepository, "2")
	addToGroup(ctx, t, groupRepository, group.ID, user1, user2)
	addToGroup(ctx, t, groupRepository, group2.ID, user3)
	tests := []struct {
		name      string
		expense   expenses.CreateExpenseContext
		expectErr bool
	}{
		{
			name: "create for self only",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: group.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user1.ID: 100,
				},
			},
			expectErr: false,
		},
		{
			name: "for someone in a group",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: group.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user2.ID: 100,
				},
			},
			expectErr: false,
		},
		{
			name: "for multiple people in a group",
			expense: expenses.CreateExpenseContext{
				UserID:  user2.ID,
				GroupID: group.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user2.ID: 10,
					user1.ID: 90,
				},
			},
			expectErr: false,
		},
		{
			name: "person not in a group",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: group.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user3.ID: 100,
				},
			},
			expectErr: true,
		},
		{
			name: "for multiple people but one not in a group",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: group.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user3.ID: 10,
					user1.ID: 20,
					user2.ID: 70,
				},
			},
			expectErr: true,
		},
		{
			name: "for non-existen group",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: 4,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user1.ID: 30,
					user2.ID: 70,
				},
			},
			expectErr: true,
		},
		{
			name: "for another group",
			expense: expenses.CreateExpenseContext{
				UserID:  user1.ID,
				GroupID: group2.ID,
				Amount:  10021.00,
				Shares: expenses.ExpenseShares{
					user1.ID: 30,
					user2.ID: 70,
				},
			},
			expectErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := expensesService.Create(ctx, test.expense)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, result.Timestamp)
			}
		})
	}
}

func createProperUser(
	ctx context.Context,
	t *testing.T,
	mailSuffix string,
	userRepository expenses.UserRepository,
) expenses.User {
	user, err := userRepository.Create(ctx, pgdb, expenses.CreateUserRequest{
		Email:    expenses.Email("mail@mail.com" + mailSuffix),
		Password: "1dsac",
	})
	require.NoError(t, err)
	return user
}

func createGroup(ctx context.Context, t *testing.T, groupRepository expenses.GroupRepository, name string) expenses.Group {
	group, err := groupRepository.Create(ctx, pgdb, util.NonEmptyString(name))
	require.NoError(t, err)
	return group
}

func addToGroup(
	ctx context.Context,
	t *testing.T,
	groupRepository expenses.GroupRepository,
	groupID uint,
	users ...expenses.User,
) {
	for _, user := range users {
		require.NoError(t, groupRepository.AddUserToGroup(ctx, pgdb, user.ID, groupID))
	}
}
