package authentication

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordEncoderChecker hashes the password and provides the means to check a password against a hashed password.
// Depending on the implementation Encode can be a no-action, the Check then will be an equality check
type PasswordEncoderChecker interface {
	PasswordEncoder
	PasswordChecker
}

// PasswordEncoder hashes the password
type PasswordEncoder interface {
	Encode(password string) (string, error)
}

// PasswordChecker compares hashed password and plain password. Returns true if they match
type PasswordChecker interface {
	Check(hashed string, toTest string) bool
}

type BCryptPasswordEncoder struct {
}

func (*BCryptPasswordEncoder) Encode(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (*BCryptPasswordEncoder) Check(hashed string, toTest string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(toTest)) == nil
}

type NoAcPasswordEncoder struct {
}

func (NoAcPasswordEncoder) Encode(password string) (string, error) {
	return password, nil
}

func (NoAcPasswordEncoder) Check(hashed string, toTest string) bool {
	return hashed == toTest
}
