package authentication_test

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/authentication/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockTokenRetriever struct {
	mock.Mock
}

func (m *mockTokenRetriever) Retrieve(uuid string) (authentication.UserContext, error) {
	args := m.Called(uuid)
	return args.Get(0).(authentication.UserContext), args.Error(1)
}

func TestNewJWTAuthorizer(t *testing.T) {
	require.NotNil(t, authentication.NewJWTAuthorizer(accessAlg, new(mockTokenRetriever)))
}

func TestJWTAuthorizerAuthorize(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	accessUUID, accessJWT := prepareValidJWT(t)
	expectedContext := authentication.UserContext{
		UserID:  11,
		GroupID: 22,
	}
	tokenRetriever.On("Retrieve", accessUUID).Return(expectedContext, nil)

	handler := func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value("user")
		require.NotNil(t, value)
		userContext, ok := value.(authentication.UserContext)
		require.True(t, ok)
		require.Equal(t, userContext, expectedContext)
		w.WriteHeader(http.StatusOK)
	}

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(handler).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestJWTAuthorizerNoHeaderForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerNoBearerForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	_, accessJWT := prepareValidJWT(t)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerNotBearerForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	_, accessJWT := prepareValidJWT(t)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearere "+accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerMoreThenOneSpaceInAuthorizationHeaderForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	_, accessJWT := prepareValidJWT(t)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+accessJWT+" a")
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerNoAccessUUIDClaimForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	accessJWT := prepareNoAccessUUIDJWT(t)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerAccessUUIDNotAStringClaimForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	accessJWT := prepareUUIDNotStringJWT(t)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerTokenNotFoundInStorageForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)
	accessUUID, accessJWT := prepareValidJWT(t)
	tokenRetriever.On("Retrieve", accessUUID).Return(authentication.UserContext{}, errors.New("expected"))

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+accessJWT)
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(http.ResponseWriter, *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestJWTAuthorizerTokenIncorrectForbidden(t *testing.T) {
	// given
	tokenRetriever := new(mockTokenRetriever)
	authorizer := authentication.NewJWTAuthorizer(accessAlg, tokenRetriever)

	request := httptest.NewRequest(http.MethodGet, "/target", nil)
	request.Header.Set("Authorization", "Bearer "+"something")
	recorder := httptest.NewRecorder()

	// when
	authorizer.Authorize(func(http.ResponseWriter, *http.Request) {}).ServeHTTP(recorder, request)

	// then
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func prepareValidJWT(t *testing.T) (string, string) {
	claims := jwt.NewClaims()
	accessUUID := "uuid-id"
	claims["access_uuid"] = accessUUID
	accessJWT, err := accessAlg.Encode(claims)
	require.NoError(t, err)
	return accessUUID, accessJWT
}

func prepareNoAccessUUIDJWT(t *testing.T) string {
	claims := jwt.NewClaims()
	accessJWT, err := accessAlg.Encode(claims)
	require.NoError(t, err)
	return accessJWT
}

func prepareUUIDNotStringJWT(t *testing.T) string {
	claims := jwt.NewClaims()
	claims["access_uuid"] = 123
	accessJWT, err := accessAlg.Encode(claims)
	require.NoError(t, err)
	return accessJWT
}
