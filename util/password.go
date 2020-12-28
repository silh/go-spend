package util

import "errors"

// Just a string that should not be printed
type Password string

func ValidPassword(password string) (Password, error) {
	if len(password) == 0 {
		return Password(""), errors.New("password can't be empty")
	}
	return Password(password), nil
}

func (r Password) String() string {
	return "***"
}
