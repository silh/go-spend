package authentication

import (
	"context"
	"go-spend/db"
	"go-spend/expenses"
)

type UserService interface {
	Create(ctx context.Context, request expenses.CreateUserRequest) (expenses.UserResponse, error)
}

// DefaultUserService responsible for business logic with User type.
type DefaultUserService struct {
	db              db.TxQuerier
	passwordEncoder PasswordEncoder
	repository      expenses.UserRepository
}

// Create DefaultUserService
func NewDefaultUserService(
	db db.TxQuerier,
	passwordEncoder PasswordEncoder,
	repository expenses.UserRepository,
) *DefaultUserService {
	return &DefaultUserService{db: db, passwordEncoder: passwordEncoder, repository: repository}
}

// Store a new user in repository. CreateUserRequest is expected to be valid.
func (d *DefaultUserService) Create(ctx context.Context, request expenses.CreateUserRequest) (expenses.UserResponse, error) {
	encodedPassword, err := d.passwordEncoder.Encode(string(request.Password))
	request.Password = expenses.Password(encodedPassword)
	if err != nil {
		return expenses.UserResponse{}, err
	}
	createdUser, err := d.repository.Create(ctx, d.db, request)
	if err != nil {
		return expenses.UserResponse{}, err
	}
	return expenses.UserResponse{ID: createdUser.ID, Email: createdUser.Email}, nil
}
