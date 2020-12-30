package authentication_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"math"
	"testing"
)

func TestUserContextValue(t *testing.T) {
	tests := []struct {
		expected string
		userID   uint
		groupID  uint
	}{
		{
			expected: "0_0",
			userID:   0,
			groupID:  0,
		},
		{
			expected: "9223372036854775807_10",
			userID:   math.MaxInt64,
			groupID:  10,
		},
	}
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			userContext := authentication.UserContext{
				UserID:  test.userID,
				GroupID: test.groupID,
			}
			assert.Equal(t, test.expected, userContext.Value())
		})
	}
}

func TestUserContextParse(t *testing.T) {
	tests := []struct {
		value    string
		expected authentication.UserContext
	}{
		{
			value: "1_1",
			expected: authentication.UserContext{
				UserID:  1,
				GroupID: 1,
			},
		},
		{
			value: "10_50",
			expected: authentication.UserContext{
				UserID:  10,
				GroupID: 50,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			userContext, err := authentication.ParseUserContext(test.value)
			require.NoError(t, err)
			assert.Equal(t, test.expected, userContext)
		})
	}
}

func TestUserContextParseErrors(t *testing.T) {
	tests := []struct {
		value string
	}{
		{
			value: "a1_1",
		},
		{
			value: "1_1a",
		},
		{
			value: "10_",
		},
		{
			value: "10",
		},
		{
			value: "_",
		},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			_, err := authentication.ParseUserContext(test.value)
			require.Error(t, err)
		})
	}
}
