package expenses

import "context"

type UserService interface {
	Create(ctx context.Context, request CreateUserRequest) (UserReponse, error)
}

// DefaultUserService responsible for business logic with User type.
type DefaultUserService struct {
	repository UserRepository
}

// Create DefaultUserService
func NewDefaultUserService(repository UserRepository) *DefaultUserService {
	return &DefaultUserService{repository: repository}
}

// Store a new user in repository. CreateUserRequest is expected to be valid.
func (d *DefaultUserService) Create(ctx context.Context, request CreateUserRequest) (UserReponse, error) {
	createdUser, err := d.repository.Create(ctx, request)
	if err != nil {
		return UserReponse{}, err
	}
	return UserReponse{ID: createdUser.ID, Email: createdUser.Email}, nil
}
