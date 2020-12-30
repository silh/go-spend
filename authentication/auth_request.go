package authentication

import (
	"bytes"
	"encoding/json"
	"go-spend/expenses"
)

// AuthRequest represents JSON body of authentication request
type AuthRequest struct {
	Email    expenses.Email    `json:"email"`
	Password expenses.Password `json:"password"`
}

// UnmarshalJSON performs unmarshalling and check of provided properties
func (a *AuthRequest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" { // by convention
		return nil
	}
	type authRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var auth authRequest
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var err error
	if err = decoder.Decode(&auth); err != nil {
		return err
	}
	a.Email, err = expenses.ValidEmail(auth.Email)
	if err != nil {
		return err
	}
	a.Password, err = expenses.ValidPassword(auth.Password)
	return err
}
