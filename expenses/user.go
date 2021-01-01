package expenses

import (
	"bytes"
	"encoding/json"
)

// Internal user type, will not be shared outside of the application. Every User can only be in one group.
type User struct {
	ID       uint
	Email    Email
	Password Password
	GroupID  uint
}

// CreateUserRequest contains information for User registration
type CreateUserRequest struct {
	Email    Email    `json:"email"`
	Password Password `json:"password"`
}

// UnmarshalJSON unmarshalls incoming JSON request and validates it.
func (r *CreateUserRequest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	type createRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req createRequest
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.DisallowUnknownFields()
	var err error
	if err = decoder.Decode(&req); err != nil {
		return err
	}
	r.Email, err = ValidEmail(req.Email)
	if err != nil {
		return err
	}
	r.Password, err = ValidPassword(req.Password)
	return err
}

// contains information returned when the User information is requested
type UserResponse struct {
	ID    uint  `json:"id"`
	Email Email `json:"email"`
}
