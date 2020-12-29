package util

import "errors"

type NonEmptyString string

func NewNonEmptyString(justString string) (NonEmptyString, error) {
	if len(justString) == 0 {
		return NonEmptyString(""), errors.New("can't be empty")
	}
	return NonEmptyString(justString), nil
}
