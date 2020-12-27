package util

import (
	"errors"
	"regexp"
	"strings"
)

var (
	simpleEmailValidationRegexp = regexp.MustCompile("^[^@\\s]+@[^@\\s.]+\\.[^@.\\s]+$")

	ErrIncorrectEmail = errors.New("invalid email")
)

// Wrapper for a string that has been validated to be an email. Only incoming strings (from request are validated). All
// others are considered to be checked before.
type Email string

// Validates an email address.
func NewEmail(email string) (Email, error) {
	if !simpleEmailValidationRegexp.MatchString(email) {
		return "", ErrIncorrectEmail
	}
	parts := strings.Split(email, "@")
	if len(parts[0]) > 64 || len(parts[1]) > 255 {
		return "", ErrIncorrectEmail
	}
	return Email(email), nil
}
