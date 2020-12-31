package expenses_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

func TestValidateCreateExpenseContext(t *testing.T) {
	tests := []struct {
		name          string
		expense       expenses.CreateExpenseContext
		expectedError error
	}{
		{
			name: "only one spender",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  100.20,
				Shares: expenses.ExpenseShares{
					1: 100,
				},
			},
			expectedError: nil,
		},
		{
			name: "multiple spenders",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  200.20,
				Shares: expenses.ExpenseShares{
					1: 65,
					2: 25,
					3: 10,
				},
			},
			expectedError: nil,
		},
		{
			name: "no user",
			expense: expenses.CreateExpenseContext{
				UserID:  0,
				GroupID: 1,
				Amount:  200.20,
				Shares: expenses.ExpenseShares{
					1: 65,
					2: 25,
					3: 10,
				},
			},
			expectedError: errors.New("incorrect user"),
		},
		{
			name: "no group",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 0,
				Amount:  200.20,
				Shares: expenses.ExpenseShares{
					1: 65,
					2: 25,
					3: 10,
				},
			},
			expectedError: errors.New("incorrect group"),
		},
		{
			name: "empty shares",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  200.20,
				Shares:  expenses.ExpenseShares{},
			},
			expectedError: errors.New("shares should contain at least one share"),
		},
		{
			name: "no shares",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  200.20,
			},
			expectedError: errors.New("shares should contain at least one share"),
		},
		{
			name: "incorrect total less than 100",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  200.20,
				Shares: expenses.ExpenseShares{
					1: 65,
					2: 25,
					3: 1,
				},
			},
			expectedError: errors.New("total percent for shares incorrect"),
		},
		{
			name: "incorrect total more than 100",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  200.20,
				Shares: expenses.ExpenseShares{
					1: 65,
					2: 25,
					3: 50,
				},
			},
			expectedError: errors.New("total percent for shares incorrect"),
		},
		{
			name: "negative amount",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  -100.20,
				Shares: expenses.ExpenseShares{
					1: 100,
				},
			},
			expectedError: errors.New("amount should be positive number"),
		},
		{
			name: "zero amount",
			expense: expenses.CreateExpenseContext{
				UserID:  1,
				GroupID: 1,
				Amount:  0,
				Shares: expenses.ExpenseShares{
					1: 100,
				},
			},
			expectedError: errors.New("amount should be positive number"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := expenses.ValidateCreateExpenseContext(test.expense)
			if test.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}
