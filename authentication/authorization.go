package authentication

import (
	"context"
	"errors"
	"go-spend/authentication/jwt"
	"net/http"
	"strings"
)

const (
	NotAuthorized = "Not Authorized"
)

var (
	ErrUserContextNotFound = errors.New("user context not found")
)

// Authorizer adds UserContext to a request
type Authorizer interface {
	Authorize(http.HandlerFunc) http.HandlerFunc
}

// JWTAuthorizer extracts credentials from Authorization header and expects them to be a JWT.
// The accessUUIDClaim is check against the values stored in redis.
type JWTAuthorizer struct {
	accessAlgorithm *jwt.Algorithm
	tokenRetriever  TokenRetriever
}

// NewJWTAuthorizer creates new instance of JWTAuthorizer
func NewJWTAuthorizer(accessAlgorithm *jwt.Algorithm, tokenRetriever TokenRetriever) *JWTAuthorizer {
	return &JWTAuthorizer{accessAlgorithm: accessAlgorithm, tokenRetriever: tokenRetriever}
}

func (a *JWTAuthorizer) Authorize(realHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, NotAuthorized, http.StatusForbidden)
			return
		}
		claims, err := a.accessAlgorithm.Decode(parts[1])
		if err != nil {
			http.Error(w, NotAuthorized, http.StatusForbidden)
			return
		}
		claim, ok := claims[accessUUIDClaim]
		if !ok {
			http.Error(w, NotAuthorized, http.StatusForbidden)
			return
		}
		uuid, ok := claim.(string)
		if !ok {
			http.Error(w, NotAuthorized, http.StatusForbidden)
			return
		}
		accessContext, err := a.tokenRetriever.Retrieve(uuid)
		if err != nil {
			http.Error(w, NotAuthorized, http.StatusForbidden)
			return
		}
		contextWithUser := context.WithValue(r.Context(), "user", accessContext)
		requestWithUser := r.WithContext(contextWithUser)
		realHandler.ServeHTTP(w, requestWithUser)
	}
}

// ExtractUser from request. It should be put there by Authorizer
func ExtractUser(r *http.Request) (UserContext, error) {
	value := r.Context().Value("user")
	if value == nil {
		return UserContext{}, ErrUserContextNotFound
	}
	userContext, ok := value.(UserContext)
	if !ok {
		return UserContext{}, ErrUserContextNotFound
	}
	return userContext, nil
}
