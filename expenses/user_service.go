package expenses

import (
	"context"
	"go-spend/db"
)

type UserService interface {
	Create(ctx context.Context, request CreateUserRequest) (UserResponse, error)
}

// DefaultUserService responsible for business logic with User type.
type DefaultUserService struct {
	db         db.TxQuerier
	repository UserRepository
}

// Create DefaultUserService
func NewDefaultUserService(db db.TxQuerier, repository UserRepository) *DefaultUserService {
	return &DefaultUserService{db: db, repository: repository}
}

// Store a new user in repository. CreateUserRequest is expected to be valid.
func (d *DefaultUserService) Create(ctx context.Context, request CreateUserRequest) (UserResponse, error) {
	createdUser, err := d.repository.Create(ctx, d.db, request)
	if err != nil {
		return UserResponse{}, err
	}
	return UserResponse{ID: createdUser.ID, Email: createdUser.Email}, nil
}
