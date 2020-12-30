package authentication

import (
	"context"
	"errors"
	"go-spend/authentication/jwt"
	"go-spend/db"
	"go-spend/expenses"
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
