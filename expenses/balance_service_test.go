package expenses_test

import (
	"context"
	"errors"
	"github.com/go-redis/redis"
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

type mockBalanceCacheGetterSetter struct {
	mock.Mock
}

func (m *mockBalanceCacheGetterSetter) Get(key expenses.BalanceCacheKey) (expenses.Balance, error) {
	args := m.Called(key)
	return args.Get(0).(expenses.Balance), args.Error(1)
}

func (m *mockBalanceCacheGetterSetter) Set(key expenses.BalanceCacheKey, balance expenses.Balance) error {
	args := m.Called(key, balance)
	return args.Error(0)
}

func TestNewDefaultBalanceService(t *testing.T) {
	cache := new(mockBalanceCacheGetterSetter)
	assert.NotNil(t, expenses.NewDefaultBalanceService(new(mockTxQuerier), cache, new(mockBalanceRepository)))
}

func TestDefaultBalanceServiceReturnsBalanceFromRepo(t *testing.T) {
	// given
	ctx := context.Background()
	balanceRepository := new(mockBalanceRepository)
	querier := new(mockTxQuerier)
	cache := new(mockBalanceCacheGetterSetter)
	balanceService := expenses.NewDefaultBalanceService(querier, cache, balanceRepository)
	balance := expenses.Balance{
		1: 10.0,
		2: -20.0,
	}
	cache.On("Get", expenses.BalanceCacheKey(1)).Return(expenses.Balance{}, redis.Nil)
	balanceRepository.On("Get", ctx, querier, uint(1)).Return(balance, nil)
	cache.On("Set", expenses.BalanceCacheKey(1), balance).Return(nil)

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
	cache := new(mockBalanceCacheGetterSetter)
	balanceService := expenses.NewDefaultBalanceService(querier, cache, balanceRepository)
	cache.On("Get", expenses.BalanceCacheKey(1)).Return(expenses.Balance{}, redis.Nil)
	balanceRepository.On("Get", ctx, querier, uint(1)).Return(expenses.Balance{}, errors.New("expected"))

	// when
	_, err := balanceService.Get(ctx, 1)

	// then
	require.Error(t, err)
}

func TestDefaultBalanceServiceGetFromCache(t *testing.T) {
	// given
	ctx := context.Background()
	balanceRepository := new(mockBalanceRepository)
	querier := new(mockTxQuerier)
	cache := new(mockBalanceCacheGetterSetter)
	balanceService := expenses.NewDefaultBalanceService(querier, cache, balanceRepository)
	balance := expenses.Balance{
		1: 10.0,
		2: -20.0,
	}
	cache.On("Get", expenses.BalanceCacheKey(1)).Return(balance, nil)

	// when
	result, err := balanceService.Get(ctx, 1)

	// then
	require.NoError(t, err)
	assert.Equal(t, balance, result)
}

func TestDefaultBalanceServiceReturnsBalanceFromRepoEvenIfSetToCacheFails(t *testing.T) {
	// given
	ctx := context.Background()
	balanceRepository := new(mockBalanceRepository)
	querier := new(mockTxQuerier)
	cache := new(mockBalanceCacheGetterSetter)
	balanceService := expenses.NewDefaultBalanceService(querier, cache, balanceRepository)
	balance := expenses.Balance{
		1: 10.0,
		2: -20.0,
	}
	cache.On("Get", expenses.BalanceCacheKey(1)).Return(expenses.Balance{}, redis.Nil)
	balanceRepository.On("Get", ctx, querier, uint(1)).Return(balance, nil)
	cache.On("Set", expenses.BalanceCacheKey(1), balance).Return(errors.New("expected"))

	// when
	result, err := balanceService.Get(ctx, 1)

	// then
	require.NoError(t, err)
	assert.Equal(t, balance, result)
}
