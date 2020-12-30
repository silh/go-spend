package authentication

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const userContextSeparator = "_"

var (
	ErrIncorrectValue = errors.New("incorrect value")
)

// UserContext contains info of the current user context. Can be stored in RedisTokenRepository as a value
type UserContext struct {
	UserID  uint
	GroupID uint
}

// Value converts to a format that can be stored in Redis
func (u *UserContext) Value() string {
	return fmt.Sprintf("%d%s%d", u.UserID, userContextSeparator, u.GroupID)
}

// ParseUserContext from a string, returns error if could not parse
func ParseUserContext(value string) (UserContext, error) {
	split := strings.Split(value, userContextSeparator)
	if len(split) != 2 {
		return UserContext{}, ErrIncorrectValue
	}
	userID, err := strconv.ParseUint(split[0], 10, 0)
	if err != nil {
		return UserContext{}, err
	}
	groupID, err := strconv.ParseUint(split[1], 10, 0)
	if err != nil {
		return UserContext{}, err
	}
	return UserContext{
		UserID:  uint(userID),
		GroupID: uint(groupID),
	}, nil
}
