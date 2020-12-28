package expenses_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

func TestValidCreateUserRequest(t *testing.T) {
	rawRequest := expenses.RawCreateUserRequest{Email: "some@mail.com", Password: "1234"}
	request, err := expenses.ValidCreateUserRequest(rawRequest)
	require.NoError(t, err)
	assert.Equal(t, rawRequest.Email, string(request.Email))
	assert.Equal(t, rawRequest.Password, string(request.Password))
}
