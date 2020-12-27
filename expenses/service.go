package expenses

import "context"

type UserService interface {
	Create(ctx context.Context, request CreateUser)
}

type DefaultUserService struct {
}

func (d *DefaultUserService) Create(ctx context.Context, request CreateUser) {
	//user := User{ID: }
	panic("implement me")
}
