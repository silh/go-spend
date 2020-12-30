package authentication

import (
	"go-spend/authentication/jwt"
	"go-spend/authentication/uuid"
	"time"
)

const (
	// claims
	accessUUIDClaim  = "access_uuid"
	refreshUUIDClaim = "refresh_uuid"
	userIDClaim      = "user_id"
	groupIDClaim     = "group_id"
	expClaim         = "exp"
)

// TokenResponse is a struct to map token values to JSON
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

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

// TokenCreator creates tokens based on provided algorithms
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
		refreshTokenExpiration: 24 * time.Hour,
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
	atClaims[accessUUIDClaim] = token.UUID
	atClaims[userIDClaim] = userID
	atClaims[groupIDClaim] = groupID
	atClaims[expClaim] = token.ExpiresAt
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
	claims[refreshUUIDClaim] = token.UUID
	claims[userIDClaim] = userID
	claims[groupIDClaim] = groupID
	claims[expClaim] = token.ExpiresAt
	token.Encoded, err = t.accessAlgorithm.Encode(claims)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}
