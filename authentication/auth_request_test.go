package authentication_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/expenses"
	"testing"
)

func TestAuthRequestUnmarshalJSON(t *testing.T) {
	authJSON := `{"email": "mail@mail.com", "password": "password"}`
	var req authentication.AuthRequest
	err := json.Unmarshal([]byte(authJSON), &req)
	require.NoError(t, err)
	assert.Equal(t, expenses.Email("mail@mail.com"), req.Email)
	assert.Equal(t, expenses.Password("password"), req.Password)
}

func TestAuthRequestUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "unexpected fields",
			json: `{"email": "mail@mail.com", "password": "password", "one": "two"}`,
		},
		{
			name: "no email field",
			json: `{"password": "password"}`,
		},
		{
			name: "incorrect email",
			json: `{"email": "somemail", password": "password"}`,
		},
		{
			name: "no password field",
			json: `{"email": "mail@mail.com"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req authentication.AuthRequest
			err := json.Unmarshal([]byte(test.json), &req)
			require.Error(t, err)
		})
	}
}
