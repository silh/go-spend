package authentication

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"go-spend/authentication/jwt"
	"go-spend/db"
	"go-spend/expenses"
	"time"
)

var (
	ErrEmailOrPasswordIncorrect = errors.New("email or password incorrect")
)

// Performs user authentication
type AuthService interface {
	Authenticate(email expenses.Email, password expenses.Password) error
}

type DefaultAuthService struct {
	accessAlgorithm  *jwt.Algorithm
	refreshAlgorithm *jwt.Algorithm
	userRepository   expenses.UserRepository
	db               db.TxQuerier
	passwordChecker  PasswordChecker
}

func NewDefaultAuthService(
	accessAlgorithm *jwt.Algorithm,
	refreshAlgorithm *jwt.Algorithm,
	userRepository expenses.UserRepository,
	db db.TxQuerier,
	passwordChecker PasswordChecker,
) *DefaultAuthService {
	return &DefaultAuthService{
		accessAlgorithm:  accessAlgorithm,
		refreshAlgorithm: refreshAlgorithm,
		userRepository:   userRepository,
		db:               db,
		passwordChecker:  passwordChecker,
	}
}

func (a *DefaultAuthService) Authenticate(ctx context.Context, email expenses.Email, password expenses.Password) error {
	user, err := a.userRepository.FindByEmail(ctx, a.db, email)
	if err != nil {
		return err
	}
	if ok := a.passwordChecker.Check(string(user.Password), string(password)); !ok {
		return ErrEmailOrPasswordIncorrect
	}
	//a.createAccessToken(user.ID, )
	return nil
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

func (a *DefaultAuthService) CreateTokenPair(userID uint, groupID uint) (TokenPair, error) {
	token, err := a.createAccessToken(userID, groupID)
	if err != nil {
		return TokenPair{}, nil
	}
	refreshToken, err := a.createRefreshToken(userID, groupID)
	return TokenPair{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (a *DefaultAuthService) createAccessToken(userID uint, groupID uint) (Token, error) {
	token := Token{}
	token.ExpiresAt = time.Now().Add(time.Minute * 15).Unix()
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
	token.Encoded, err = a.accessAlgorithm.Encode(atClaims)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

func (a *DefaultAuthService) createRefreshToken(userID uint, groupID uint) (Token, error) {
	token := Token{}
	token.ExpiresAt = time.Now().Add(time.Minute * 15).Unix()
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
	token.Encoded, err = a.accessAlgorithm.Encode(claims)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}
