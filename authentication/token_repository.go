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

type RedisTokenRepository struct {
	redis redis.UniversalClient
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
