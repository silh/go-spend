package authentication_test

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"testing"
	"time"
)

func TestNewRedisTokenRepository(t *testing.T) {
	tokenRepository := authentication.NewRedisTokenRepository(redisClient)
	require.NotNil(t, tokenRepository)
}

func TestRedisTokenRepositorySaveRetrieve(t *testing.T) {
	clearRedis()
	// given
	tokenRepository := authentication.NewRedisTokenRepository(redisClient)
	userContext := authentication.UserContext{
		UserID:  1111,
		GroupID: 21,
	}
	tokenPair := authentication.TokenPair{
		AccessToken: authentication.Token{
			Encoded:   "jshgfhja12",
			UUID:      "aa-22s",
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
		RefreshToken: authentication.Token{
			Encoded:   "aaaazxczcxz",
			UUID:      "222ds2",
			ExpiresAt: time.Now().Add(16 * time.Minute).Unix(),
		},
	}

	// when and then
	require.NoError(t, tokenRepository.Save(tokenPair, userContext))
	accessContext, err := tokenRepository.Retrieve(tokenPair.AccessToken.UUID)
	require.NoError(t, err)
	assert.Equal(t, userContext, accessContext)
	refreshContext, err := tokenRepository.Retrieve(tokenPair.RefreshToken.UUID)
	require.NoError(t, err)
	assert.Equal(t, refreshContext, refreshContext)
}

// Due to the nature of redis methods return type only simple test is present.
// It is possible to wrap redis client, but that was omitted for now.
func TestRedisTokenRepositorySaveErrors(t *testing.T) {
	// given
	localClient := redis.NewClient(&redis.Options{})
	tokenRepository := authentication.NewRedisTokenRepository(localClient)

	userContext := authentication.UserContext{
		UserID:  1111,
		GroupID: 21,
	}
	tokenPair := authentication.TokenPair{
		AccessToken: authentication.Token{
			Encoded:   "jshgfhja12",
			UUID:      "aa-22s",
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
		RefreshToken: authentication.Token{
			Encoded:   "aaaazxczcxz",
			UUID:      "222ds2",
			ExpiresAt: time.Now().Add(16 * time.Minute).Unix(),
		},
	}
	require.Error(t, tokenRepository.Save(tokenPair, userContext))
}
