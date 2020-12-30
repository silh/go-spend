package authentication

import (
	"context"
	"errors"
	"go-spend/db"
	"go-spend/expenses"
)

var (
	ErrEmailOrPasswordIncorrect = errors.New("email or password incorrect")
)

// Authenticator performs user authentication
type Authenticator interface {
	Authenticate(
		ctx context.Context,
		email expenses.Email,
		password expenses.Password,
	) (TokenResponse, error)
}

// AuthService is a default implementation of Authenticator
type AuthService struct {
	tokenCreator    *TokenCreator
	userRepository  expenses.UserRepository
	db              db.TxQuerier
	passwordChecker PasswordChecker
}

func NewAuthService(
	tokenCreator *TokenCreator,
	userRepository expenses.UserRepository,
	db db.TxQuerier,
	passwordChecker PasswordChecker,
) *AuthService {
	return &AuthService{
		tokenCreator:    tokenCreator,
		userRepository:  userRepository,
		db:              db,
		passwordChecker: passwordChecker,
	}
}

// Authenticate performs user authentication. If user was not found or if password was incorrect -
// ErrEmailOrPasswordIncorrect is returned.
func (a *AuthService) Authenticate(
	ctx context.Context,
	email expenses.Email,
	password expenses.Password,
) (TokenResponse, error) {
	user, err := a.userRepository.FindByEmail(ctx, a.db, email)
	if err != nil {
		if err == expenses.ErrUserNotFound {
			return TokenResponse{}, ErrEmailOrPasswordIncorrect
		}
		return TokenResponse{}, err
	}
	if ok := a.passwordChecker.Check(string(user.Password), string(password)); !ok {
		return TokenResponse{}, ErrEmailOrPasswordIncorrect
	}
	tokenPair, err := a.tokenCreator.CreateTokenPair(user.ID, user.GroupID)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{
		AccessToken:  tokenPair.AccessToken.Encoded,
		RefreshToken: tokenPair.RefreshToken.Encoded,
	}, nil
}
