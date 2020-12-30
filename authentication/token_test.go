package authentication_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/authentication/jwt"
	"testing"
	"time"
)

var (
	accessAlg  = jwt.HmacSha256("accessKey")
	refreshAlg = jwt.HmacSha256("refreshKey")
)

func TestNewTokenCreator(t *testing.T) {
	creator := authentication.NewTokenCreator(accessAlg, refreshAlg)
	assert.NotNil(t, creator)
}

func TestCreateTokenPair(t *testing.T) {
	creator := authentication.NewTokenCreator(accessAlg, refreshAlg)
	var userID uint = 123
	var groupId uint = 1
	tokenPair, err := creator.CreateTokenPair(userID, groupId)
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	assert.NotZero(t, userID, tokenPair.AccessToken.UUID)
	assert.NotZero(t, userID, tokenPair.AccessToken.Encoded)
	accessTimeLeft := tokenPair.AccessToken.ExpiresAt - time.Now().Unix()
	assert.True(t, accessTimeLeft > 0)
	refreshTimeLeft := tokenPair.RefreshToken.ExpiresAt - time.Now().Unix()
	assert.True(t, refreshTimeLeft > 0)
}
