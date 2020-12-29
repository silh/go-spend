package jwt

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var secret = "my secret"

func TestEncodeAndValidateToken(t *testing.T) {
	algorithm := HmacSha256(secret)
	payload := NewClaims()
	payload.SetTime("nbf", time.Now().Add(time.Duration(-1)*time.Hour))
	payload.SetTime("exp", time.Now().Add(time.Duration(100)*time.Hour))

	token, err := algorithm.Encode(payload)
	require.NoError(t, err)
	require.NoError(t, algorithm.Validate(token))
}

func TestValidateToken(t *testing.T) {
	algorithm := HmacSha256(secret)
	payload := NewClaims()
	err := json.Unmarshal([]byte(`{"sub":"1234567890","name":"John Doe","admin":true}`), &payload)
	require.NoError(t, err)

	token, err := algorithm.Encode(payload)
	require.NoError(t, err)

	tokenComponents := strings.Split(token, ".")

	invalidSignature := "cBab30RMHrHDcEfxjoYZgeFONFh7Hg"
	invalidToken := tokenComponents[0] + "." + tokenComponents[1] + "." + invalidSignature

	require.Error(t, algorithm.Validate(invalidToken))
}

func TestVerifyTokenExp(t *testing.T) {
	algorithm := HmacSha256(secret)
	payload := NewClaims()
	payload["exp"] = fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).Unix())

	err := json.Unmarshal([]byte(`{"sub":"1234567890","name":"John Doe","admin":true}`), &payload)
	require.NoError(t, err)

	token, err := algorithm.Encode(payload)
	require.NoError(t, err)

	err = algorithm.Validate(token)
	require.Error(t, err)
}

func TestVerifyTokenNbf(t *testing.T) {
	algorithm := HmacSha256(secret)

	payload := NewClaims()
	payload.SetTime("nbf", time.Now().Add(time.Duration(1)*time.Hour))

	err := json.Unmarshal([]byte(`{"sub":"1234567890","name":"John Doe","admin":true}`), &payload)
	require.NoError(t, err)

	token, err := algorithm.Encode(payload)
	require.NoError(t, err)

	err = algorithm.Validate(token)
	require.Error(t, err)
}

func TestDecodeMalformedToken(t *testing.T) {
	algorithm := HmacSha256(secret)
	bogusTokens := []string{"", "abc", "czwmS6hE.NZLElvuy"}

	for _, bogusToken := range bogusTokens {
		_, err := algorithm.Decode(bogusToken)
		require.Error(t, err)
	}
}

func TestValidateExternalToken(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImp0aSI6ImZmNzJkMWM5LTMzMTktNGIyOS04YjlhLWU1OThkNGJhNDRlZCJ9.eyJpc3MiOiJodHRwOi8vbG9jYWwuaG9zdC5jb20iLCJhdWQiOiJodHRwOi8vbG9jYWwuaG9zdC5jb20iLCJqdGkiOiJmZjcyZDFjOS0zMzE5LTRiMjktOGI5YS1lNTk4ZDRiYTQ0ZWQiLCJpYXQiOjE1MTkzMjc2NDYsIm5iZiI6MTUxOTMyNzY1MCwiZXhwIjoxNjQwMzkwNDAwfQ.ASo8eiekkwZ7on43S9n697x-SqmdehY680GetK_KqpI"

	algorithm := HmacSha256("this-needs-a-test")
	require.NoError(t, algorithm.Validate(token))
}
