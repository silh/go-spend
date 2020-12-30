package authentication

import (
	"github.com/go-redis/redis"
	"time"
)

// TokenSaver stores both access and refresh token information
type TokenSaver interface {
	Save(pair TokenPair, userContext UserContext) error
}

// TokenRetriever retrieves UserContext for provided token if present
type TokenRetriever interface {
	Retrieve(token Token) (UserContext, error)
}

// Combines capabilities to store and retrieve values from the storage
type TokenRepository interface {
	TokenSaver
	TokenRetriever
}

// SimpleRedisClient provides only necessary methods to simplify testing, see redis.UniversalClient
type SimpleRedisClient interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
}

// RedisTokenRepository is TokenSaver which stores value in redis
type RedisTokenRepository struct {
	redis SimpleRedisClient
}

// NewRedisTokenRepository creates new instance of RedisTokenRepository that works with provided redis client
func NewRedisTokenRepository(redis SimpleRedisClient) *RedisTokenRepository {
	return &RedisTokenRepository{redis: redis}
}

func (r *RedisTokenRepository) Save(pair TokenPair, userContext UserContext) error {
	var err error
	now := time.Now()
	accessDuration := time.Unix(pair.AccessToken.ExpiresAt, 0).Sub(now)
	err = r.redis.Set(pair.AccessToken.UUID, userContext.Value(), accessDuration).Err()
	if err != nil {
		return err
	}
	refreshDuration := time.Unix(pair.RefreshToken.ExpiresAt, 0).Sub(now)
	return r.redis.Set(pair.RefreshToken.UUID, userContext.Value(), refreshDuration).Err()
}

func (r *RedisTokenRepository) Retrieve(token Token) (UserContext, error) {
	value, err := r.redis.Get(token.UUID).Result()
	if err != nil {
		return UserContext{}, err
	}
	return ParseUserContext(value)
}
