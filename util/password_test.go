package util_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/util"
	"testing"
)

func TestPrintRawPassword(t *testing.T) {
	result := fmt.Sprintf("%s", util.Password("123"))
	assert.Equal(t, "***", result)
}

func TestValidPassword(t *testing.T) {
	password, err := util.ValidPassword("124")
	require.NoError(t, err)
	assert.Equal(t, "124", string(password))
}

func TestValidPasswordCantByEmpty(t *testing.T) {
	password, err := util.ValidPassword("")
	assert.Error(t, err)
	assert.Zero(t, password)
}
