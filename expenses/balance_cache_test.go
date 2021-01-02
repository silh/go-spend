package expenses_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
	"time"
)

func TestRedisBalanceCacheGetSetRemove(t *testing.T) {
	// given
	defer clearRedis()
	cache := expenses.NewRedisBalanceCache(redisClient, time.Minute)
	key1 := expenses.BalanceCacheKey(1)
	key2 := expenses.BalanceCacheKey(2)
	balance1 := expenses.Balance{
		1: 20,
		3: -10.0,
	}
	balance2 := expenses.Balance{
		2: 100.0,
	}
	// when and then - set and retrieve
	require.NoError(t, cache.Set(key1, balance1))
	require.NoError(t, cache.Set(key2, balance2))
	foundBalance1, err := cache.Get(key1)
	require.NoError(t, err)
	assert.Equal(t, balance1, foundBalance1)
	foundBalance2, err := cache.Get(key2)
	require.NoError(t, err)
	assert.Equal(t, balance2, foundBalance2)
	// reset
	newBalance1 := balance2
	require.NoError(t, cache.Set(key1, newBalance1))
	foundBalance1, err = cache.Get(key1)
	require.NoError(t, err)
	assert.Equal(t, newBalance1, foundBalance1)
	// delete all
	require.NoError(t, cache.Remove(key1, key2))
	foundBalance1, err = cache.Get(key1)
	require.Error(t, err)
	foundBalance2, err = cache.Get(key2)
	require.Error(t, err)
}

func TestRedisBalanceCacheTimeout(t *testing.T) {
	cache := expenses.NewRedisBalanceCache(redisClient, 100*time.Millisecond)
	key := expenses.BalanceCacheKey(10)
	require.NoError(t, cache.Set(key, expenses.Balance{1: 2}))
	time.Sleep(100 * time.Millisecond)
	_, err := cache.Get(key)
	require.Error(t, err)
}
