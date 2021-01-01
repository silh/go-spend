package authentication_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"testing"
)

func TestNewBCryptPasswordEncoder(t *testing.T) {
	assert.NotNil(t, authentication.NewBCryptPasswordEncoder())
}

func TestBCryptPasswordEncoder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "12414681s"},
		{name: "T^E*&"},
		{name: "оылфврфлор(*ЦУ*?"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoder := authentication.BCryptPasswordEncoder{}
			password := test.name
			hashed1, err := encoder.Encode(password)
			require.NoError(t, err)
			hashed2, err := encoder.Encode(password)
			require.NoError(t, err)
			assert.True(t, encoder.Check(hashed1, password))
			assert.True(t, encoder.Check(hashed2, password))
			assert.NotEqual(t, hashed1, hashed2)
		})
	}
}

func TestNoAcPasswordEncoder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "12414681s"},
		{name: "T^E*&"},
		{name: "оылфврфлор(*ЦУ*?"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoder := authentication.NoAcPasswordEncoder{}
			password := test.name
			hashed1, err := encoder.Encode(password)
			require.NoError(t, err)
			hashed2, err := encoder.Encode(password)
			require.NoError(t, err)
			assert.Equal(t, hashed1, password)
			assert.Equal(t, hashed1, hashed2)
			assert.Equal(t, hashed2, password)
			assert.True(t, encoder.Check(hashed1, password))
			assert.True(t, encoder.Check(hashed2, password))
		})
	}
}
