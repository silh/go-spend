package expenses_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

func TestPrintRawPassword(t *testing.T) {
	result := fmt.Sprintf("%s", expenses.Password("123"))
	assert.Equal(t, "***", result)
}

func TestValidPassword(t *testing.T) {
	password, err := expenses.ValidPassword("124")
	require.NoError(t, err)
	assert.Equal(t, "124", string(password))
}

func TestValidPasswordCantByEmpty(t *testing.T) {
	password, err := expenses.ValidPassword("")
	assert.Error(t, err)
	assert.Zero(t, password)
}
