package expenses

// Internal user type, will not be shared outside of the application. Every User can only be in one group.
type User struct {
	ID       uint
	Email    Email
	Password Password
}

// Raw create user request
type RawCreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validated create user request
type CreateUserContext struct {
	Email    Email
	Password Password
}

// Validates email and password
func ValidCreateUserRequest(rawRequest RawCreateUserRequest) (CreateUserContext, error) {
	email, err := ValidEmail(rawRequest.Email)
	if err != nil {
		return CreateUserContext{}, err
	}
	password, err := ValidPassword(rawRequest.Password)
	if err != nil {
		return CreateUserContext{}, err
	}
	return CreateUserContext{Email: email, Password: password}, nil
}

// contains information returned when the User information is requested
type UserResponse struct {
	ID    uint
	Email Email
}
