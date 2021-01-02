package expenses

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// BalanceCache provides operation to Set key-value in cache, Get value by key, Remove n key-values by provided keys
type BalanceCache interface {
	BalanceCacheGetterSetter
	BalanceCacheCleaner
}

// BalanceCacheGetterSetter provides get and set operations on cache
type BalanceCacheGetterSetter interface {
	// Get value by key. If value is not present - error is returned.
	Get(key BalanceCacheKey) (Balance, error)
	// Set key-value
	Set(key BalanceCacheKey, balance Balance) error
}

// BalanceCacheCleaner provides operation to clean the cache by key
type BalanceCacheCleaner interface {
	// Remove n key-values by provided keys
	Remove(keys ...BalanceCacheKey) error
}

// RedisBalanceCache is a BalanceCache that uses redis as a cache backend.
type RedisBalanceCache struct {
	redisClient   redis.UniversalClient
	cacheDuration time.Duration
}

// NewRedisBalanceCache creates new instance of RedisBalanceCache. The cache will be stored with duration of
// provided cacheDuration.
func NewRedisBalanceCache(redisClient redis.UniversalClient, cacheDuration time.Duration) *RedisBalanceCache {
	return &RedisBalanceCache{redisClient: redisClient, cacheDuration: cacheDuration}
}

// Get value from cache by key, if nothing is found redis.Nil error will be returned
func (r *RedisBalanceCache) Get(key BalanceCacheKey) (Balance, error) {
	result, err := r.redisClient.Get(key.AsKey()).Result() // no refresh of expire
	if err != nil {
		return nil, err
	}
	balance := make(Balance)
	if err = json.Unmarshal([]byte(result), &balance); err != nil {
		return nil, err
	}
	return balance, nil
}

// Set key-value
func (r *RedisBalanceCache) Set(key BalanceCacheKey, balance Balance) error {
	data, err := json.Marshal(&balance)
	if err != nil {
		return err
	}
	return r.redisClient.Set(key.AsKey(), string(data), r.cacheDuration).Err()
}

// Remove n key-values by provided keys
func (r *RedisBalanceCache) Remove(keys ...BalanceCacheKey) error {
	stringKeys := make([]string, len(keys))
	for i, key := range keys {
		stringKeys[i] = key.AsKey()
	}
	return r.redisClient.Del(stringKeys...).Err()
}

// BalanceCacheKey contains cache key information
type BalanceCacheKey uint

// AsKey creates a string key to be used with cache.
func (b *BalanceCacheKey) AsKey() string {
	return fmt.Sprintf("%d_balance", *b) // should be ok without nil check
}
