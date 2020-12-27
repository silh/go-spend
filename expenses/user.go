package expenses

import (
	"go-spend/util"
)

type User struct {
	ID       uint
	Email    util.Email
	Password string
}

type CreateUserRequest struct {
	Email       util.Email
	RawPassword string
}

type UserResponse struct {
	ID    uint
	Email util.Email
}
