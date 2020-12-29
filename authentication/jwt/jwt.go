package jwt

import (
	"time"
)

// Header contains important information for encrypting / decrypting
type Header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
	Cty string `json:"cty"`
}

// Claims contains the claims of a jwt.
type Claims map[string]interface{}

// NewClaims returns a new map representing the claims.
func NewClaims() Claims {
	claimsMap := Claims(make(map[string]interface{}))
	return claimsMap
}

// SetTime sets the claim given to the specified time.
func (c Claims) SetTime(key string, value time.Time) {
	c[key] = value.Unix()
}

// GetTime attempts to return a claim as a time.
func (c Claims) GetTime(key string) (time.Time, bool) {
	raw, ok := c[key]
	if !ok {
		return time.Unix(0, 0), false
	}
	timeFloat, ok := raw.(float64)
	if !ok {
		return time.Unix(0, 0), false
	}

	return time.Unix(int64(timeFloat), 0), true
}

// HasClaim returns if the claims map has the specified key.
func (c Claims) HasClaim(key string) bool {
	_, ok := c[key]
	return ok
}
