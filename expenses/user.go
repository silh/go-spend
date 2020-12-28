package expenses

import (
	"go-spend/util"
)

// Internal user type, will not be shared outside of the application.
type User struct {
	ID       uint
	Email    util.Email
	Password util.Password
}

// Raw create user request
type RawCreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validated create user request
type CreateUserRequest struct {
	Email    util.Email
	Password util.Password
}

// Validates email and password
func ValidCreateUserRequest(rawRequest RawCreateUserRequest) (CreateUserRequest, error) {
	email, err := util.ValidEmail(rawRequest.Email)
	if err != nil {
		return CreateUserRequest{}, err
	}
	password, err := util.ValidPassword(rawRequest.Password)
	if err != nil {
		return CreateUserRequest{}, err
	}
	return CreateUserRequest{Email: email, Password: password}, nil
}

// contains information returned when the User information is requested
type UserResponse struct {
	ID    uint
	Email util.Email
}
