package util_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/util"
	"testing"
)

func TestNewNonEmptyString(t *testing.T) {
	justString := "12sdad"
	result, err := util.NewNonEmptyString(justString)
	require.NoError(t, err)
	assert.Equal(t, justString, string(result))
}

func TestNonEmptyStringCantBeEmpty(t *testing.T) {
	result, err := util.NewNonEmptyString("")
	assert.Error(t, err)
	assert.Zero(t, result)
}
