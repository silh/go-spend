package authentication

import (
	"go-spend/authentication/jwt"
	"go-spend/authentication/uuid"
	"time"
)

// TokenPair contains info about both access and refresh tokens
type TokenPair struct {
	AccessToken  Token
	RefreshToken Token
}

// Token contains information about either access or refresh token
type Token struct {
	Encoded   string
	UUID      string
	ExpiresAt int64
}

type TokenCreator struct {
	accessAlgorithm        *jwt.Algorithm
	refreshAlgorithm       *jwt.Algorithm
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
}

func NewTokenCreator(accessAlgorithm *jwt.Algorithm, refreshAlgorithm *jwt.Algorithm) *TokenCreator {
	return &TokenCreator{
		accessAlgorithm:        accessAlgorithm,
		refreshAlgorithm:       refreshAlgorithm,
		accessTokenExpiration:  15 * time.Minute, // hardcoded for now, if necessary can be made configurable
		refreshTokenExpiration: 30 * time.Minute,
	}
}

func (t *TokenCreator) CreateTokenPair(userID uint, groupID uint) (TokenPair, error) {
	token, err := t.createAccessToken(userID, groupID)
	if err != nil {
		return TokenPair{}, nil
	}
	refreshToken, err := t.createRefreshToken(userID, groupID)
	return TokenPair{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (t *TokenCreator) createAccessToken(userID uint, groupID uint) (Token, error) {
	token := Token{}
	token.ExpiresAt = time.Now().Add(t.accessTokenExpiration).Unix()
	newUUID, err := uuid.NewV4()
	if err != nil {
		return Token{}, err
	}
	token.UUID = newUUID.String()
	atClaims := jwt.NewClaims()
	atClaims["access_uuid"] = token.UUID
	atClaims["user_id"] = userID
	atClaims["group_id"] = groupID
	atClaims["exp"] = token.ExpiresAt
	token.Encoded, err = t.accessAlgorithm.Encode(atClaims)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

func (t *TokenCreator) createRefreshToken(userID uint, groupID uint) (Token, error) {
	token := Token{}
	token.ExpiresAt = time.Now().Add(t.refreshTokenExpiration).Unix()
	newUUID, err := uuid.NewV4()
	if err != nil {
		return Token{}, err
	}
	token.UUID = newUUID.String()
	claims := jwt.NewClaims()
	claims["refresh_uuid"] = token.UUID
	claims["user_id"] = userID
	claims["group_id"] = groupID
	claims["exp"] = token.ExpiresAt
	token.Encoded, err = t.accessAlgorithm.Encode(claims)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}
