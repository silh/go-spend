package expenses_test

import (
	"context"
	"errors"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"testing"
)

type mockExpensesRepository struct {
	mock.Mock
}

func (m *mockExpensesRepository) Create(ctx context.Context, db pgxtype.Querier, req expenses.NewExpense) (expenses.Expense, error) {
	args := m.Called(ctx, db, req)
	return args.Get(0).(expenses.Expense), args.Error(1)
}

func (m *mockExpensesRepository) CreateShares(ctx context.Context, db pgxtype.Querier, req expenses.CreateExpenseShares) error {
	args := m.Called(ctx, db, req)
	return args.Error(0)
}

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

func TestExpensesServiceFailedToStartTx(t *testing.T) {
	// given
	ctx := context.Background()
	db := new(mockTxQuerier)
	service := expenses.NewDefaultService(db, new(mockGroupRepository), new(mockExpensesRepository))

	db.On("Begin", ctx).Return(nil, errors.New("expected"))

	// when
	_, err := service.Create(ctx, expenses.CreateExpenseContext{})
	require.EqualError(t, err, "expected")
}

func TestExpensesServiceExpensesCreateFails(t *testing.T) {
	// given
	ctx := context.Background()
	db := new(mockTxQuerier)
	tx := new(mockTx)
	expensesRepository := new(mockExpensesRepository)
	groupRepository := new(mockGroupRepository)
	service := expenses.NewDefaultService(db, groupRepository, expensesRepository)

	expenseContext := expenses.CreateExpenseContext{
		UserID:  1,
		GroupID: 2,
		Amount:  10.0,
		Shares: expenses.ExpenseShares{
			1: 100,
		},
	}
	db.On("Begin", ctx).Return(tx, nil)
	groupRepository.On("FindByIDWithUsers", ctx, tx, expenseContext.GroupID).
		Return(expenses.GroupResponse{
			ID:   2,
			Name: "2",
			Users: []expenses.UserResponse{
				{
					ID:    1,
					Email: "mm@mm.mm",
				},
			},
		}, nil)
	expensesRepository.On("Create", ctx, tx, mock.Anything).
		Return(expenses.Expense{}, errors.New("expected"))

	// when
	_, err := service.Create(ctx, expenseContext)

	// then
	require.EqualError(t, err, "expected")
}

func TestExpensesServiceCreateSharesFails(t *testing.T) {
	// given
	ctx := context.Background()
	db := new(mockTxQuerier)
	tx := new(mockTx)
	expensesRepository := new(mockExpensesRepository)
	groupRepository := new(mockGroupRepository)
	service := expenses.NewDefaultService(db, groupRepository, expensesRepository)
	expenseContext := expenses.CreateExpenseContext{
		UserID:  1,
		GroupID: 2,
		Amount:  10.0,
		Shares: expenses.ExpenseShares{
			1: 100,
		},
	}

	db.On("Begin", ctx).Return(tx, nil)
	groupRepository.On("FindByIDWithUsers", ctx, tx, expenseContext.GroupID).
		Return(expenses.GroupResponse{
			ID:   2,
			Name: "2",
			Users: []expenses.UserResponse{
				{
					ID:    1,
					Email: "mm@mm.mm",
				},
			},
		}, nil)
	expensesRepository.On("Create", ctx, tx, mock.Anything).
		Return(expenses.Expense{}, errors.New("expected"))

	// when
	_, err := service.Create(ctx, expenseContext)

	// then
	require.EqualError(t, err, "expected")
}

func TestExpensesServiceFindUsersFails(t *testing.T) {
	// given
	ctx := context.Background()
	db := new(mockTxQuerier)
	tx := new(mockTx)
	expensesRepository := new(mockExpensesRepository)
	groupRepository := new(mockGroupRepository)
	service := expenses.NewDefaultService(db, groupRepository, expensesRepository)
	expenseContext := expenses.CreateExpenseContext{
		UserID:  1,
		GroupID: 2,
		Amount:  10.0,
		Shares: expenses.ExpenseShares{
			1: 100,
		},
	}

	db.On("Begin", ctx).Return(tx, nil)
	groupRepository.On("FindByIDWithUsers", ctx, tx, expenseContext.GroupID).
		Return(expenses.GroupResponse{}, errors.New("expected"))

	// when
	_, err := service.Create(ctx, expenseContext)

	// then
	require.EqualError(t, err, "expected")
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
