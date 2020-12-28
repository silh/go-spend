package util_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/util"
	"strings"
	"testing"
)

func TestValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "simple email",
			email: "email@email.com",
		},
		{
			name:  "with russian letters",
			email: "someemail@почта.рф",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := util.ValidEmail(test.email)
			require.NoError(t, err)
			assert.Equal(t, test.email, string(result))
		})
	}
}

func TestValidEmailErrors(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "no at",
			email: "emailemail.com",
		},
		{
			name:  "no dot",
			email: "someemail@mailcom",
		},
		{
			name:  "double dot",
			email: "someemail@mail..com",
		},
		{
			name: "local part too long",
			email: func() string {
				builder := strings.Builder{}
				for i := 0; i < 65; i++ {
					builder.WriteRune('i')
				}
				builder.WriteString("@email.com")
				return builder.String()
			}(),
		},
		{
			name: "domain part too long",
			email: func() string {
				builder := strings.Builder{}
				builder.WriteString("someone@")
				for i := 0; i < 256-len(".com"); i++ {
					builder.WriteRune('i')
				}
				builder.WriteString(".com")
				return builder.String()
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := util.ValidEmail(test.email)
			require.Error(t, err)
			assert.Zero(t, result)
		})
	}
}
