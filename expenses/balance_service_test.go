package expenses_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/db"
	"go-spend/expenses"
	"testing"
)

type mockBalanceRepository struct {
	mock.Mock
}

func (m *mockBalanceRepository) Get(ctx context.Context, db db.TxQuerier, userID uint) (expenses.Balance, error) {
	args := m.Called(ctx, db, userID)
	return args.Get(0).(expenses.Balance), args.Error(1)
}

func TestNewDefaultBalanceService(t *testing.T) {
	assert.NotNil(t, expenses.NewDefaultBalanceService(new(mockTxQuerier), new(mockBalanceRepository)))
}

func TestDefaultBalanceServiceReturnsBalance(t *testing.T) {
	// given
	ctx := context.Background()
	balanceRepository := new(mockBalanceRepository)
	querier := new(mockTxQuerier)
	balanceService := expenses.NewDefaultBalanceService(querier, balanceRepository)
	balance := expenses.Balance{
		1: 10.0,
		2: -20.0,
	}
	balanceRepository.On("Get", ctx, querier, uint(1)).Return(balance, nil)

	// when
	result, err := balanceService.Get(ctx, 1)

	// then
	require.NoError(t, err)
	assert.Equal(t, balance, result)
}

func TestDefaultBalanceServiceReturnsError(t *testing.T) {
	// given
	ctx := context.Background()
	balanceRepository := new(mockBalanceRepository)
	querier := new(mockTxQuerier)
	balanceService := expenses.NewDefaultBalanceService(querier, balanceRepository)
	balanceRepository.On("Get", ctx, querier, uint(1)).Return(expenses.Balance{}, errors.New("expected"))

	// when
	_, err := balanceService.Get(ctx, 1)

	// then
	require.Error(t, err)
}
