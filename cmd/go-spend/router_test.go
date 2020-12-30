package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/authentication/jwt"
	"go-spend/cmd/go-spend"
	"go-spend/expenses"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, request expenses.CreateUserRequest) (expenses.UserResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.UserResponse), args.Error(1)
}

type mockAuthenticator struct {
	mock.Mock
}

func (m *mockAuthenticator) Authenticate(
	ctx context.Context,
	email expenses.Email,
	password expenses.Password,
) (authentication.TokenResponse, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(authentication.TokenResponse), args.Error(1)
}

type mockGroupService struct {
	mock.Mock
}

func (m *mockGroupService) Create(ctx context.Context, request expenses.CreateGroupRequest) (expenses.GroupResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.GroupResponse), args.Error(1)
}

func (m *mockGroupService) FindByID(_ context.Context, _ uint) (expenses.GroupResponse, error) {
	panic("implement me")
}

type mockAuthorizer struct {
	mock.Mock
}

func (m *mockAuthorizer) Authorize(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return handlerFunc
}

type mockTokenRetriever struct {
	mock.Mock
}

func (m *mockTokenRetriever) Retrieve(uuid string) (authentication.UserContext, error) {
	args := m.Called(uuid)
	return args.Get(0).(authentication.UserContext), args.Error(1)
}

func TestNewRouter(t *testing.T) {
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockGroupService),
		new(mockUserService),
	)
	assert.NotNil(t, router)
}

func TestCreateUserWithProperParams(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockGroupService),
		userService,
	)

	createUserRequest := expenses.CreateUserRequest{Email: "some@mail.com", Password: "1234"}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()
	userResponse := expenses.UserResponse{ID: 1, Email: expenses.Email(createUserRequest.Email)}
	userService.On(
		"Create",
		mock.Anything,
		mock.AnythingOfType("expenses.CreateUserRequest"),
	).Return(userResponse, nil)

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var result expenses.UserResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, userResponse, result)
}

func TestCreateUserWithIncorrectMethod(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockGroupService),
		userService,
	)

	createUserRequest := expenses.CreateUserRequest{Email: "some@mail.com", Password: "1234"}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestCreateUserWithIncorrectBody(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockGroupService),
		userService,
	)

	createUserRequest := struct {
		something int64
	}{something: 10}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateUserWithEmptyFields(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockGroupService),
		userService,
	)

	createUserRequest := struct {
		something int64
	}{something: 10}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateUserWithSomeIncorrectFields(t *testing.T) {
	tests := []struct {
		name string
		body expenses.CreateUserRequest
	}{
		{
			name: "no email",
			body: expenses.CreateUserRequest{Password: "1234"},
		},
		{
			name: "no password",
			body: expenses.CreateUserRequest{Email: "some@email.com"},
		},
		{
			name: "invalid email",
			body: expenses.CreateUserRequest{Email: "someemail.com", Password: "12341"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			userService := new(mockUserService)
			router := main.NewRouter(
				new(mockAuthenticator),
				new(mockAuthorizer),
				new(mockGroupService),
				userService,
			)
			jsonBody, err := json.Marshal(&test.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
			recorder := httptest.NewRecorder()

			// when
			router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func TestCreateUserServiceError(t *testing.T) {
	tests := []struct {
		name         string
		prepareMock  func(userService *mockUserService)
		expectedCode int
	}{
		{
			name: "user already exists",
			prepareMock: func(userService *mockUserService) {
				userService.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("expenses.CreateUserRequest"),
				).Return(expenses.UserResponse{}, expenses.ErrEmailAlreadyExists)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "other errors",
			prepareMock: func(userService *mockUserService) {
				userService.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("expenses.CreateUserRequest"),
				).Return(expenses.UserResponse{}, errors.New("expected"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			userService := new(mockUserService)
			router := main.NewRouter(
				new(mockAuthenticator),
				new(mockAuthorizer),
				new(mockGroupService),
				userService,
			)

			createUserRequest := expenses.CreateUserRequest{Email: "some@mail.com", Password: "1234"}
			jsonBody, err := json.Marshal(createUserRequest)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
			recorder := httptest.NewRecorder()
			test.prepareMock(userService)
			// when
			router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	// given
	authenticator := new(mockAuthenticator)
	router := main.NewRouter(
		authenticator,
		new(mockAuthorizer),
		new(mockGroupService),
		new(mockUserService),
	)

	authRequest := authentication.AuthRequest{
		Email:    "mail@mail.com",
		Password: "password",
	}
	body, err := json.Marshal(&authRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/authenticate", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	expectedTokenResponse := authentication.TokenResponse{AccessToken: "asjkhdakj17", RefreshToken: "sakhadkj71"}
	authenticator.On("Authenticate", context.Background(), authRequest.Email, authRequest.Password).
		Return(expectedTokenResponse, nil)

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code)
	var response authentication.TokenResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	assert.Equal(t, expectedTokenResponse, response)
}

func TestAuthenticateFailed(t *testing.T) {
	authRequest := authentication.AuthRequest{
		Email:    "mail@mail.com",
		Password: "password",
	}
	tests := []struct {
		name         string
		expectedCode int
		method       string
		prepareMock  func(*mockAuthenticator)
	}{
		{
			name:         "incorrect creds",
			expectedCode: http.StatusUnauthorized,
			method:       http.MethodPost,
			prepareMock: func(authenticator *mockAuthenticator) {
				authenticator.On("Authenticate", context.Background(), authRequest.Email, authRequest.Password).
					Return(authentication.TokenResponse{}, authentication.ErrEmailOrPasswordIncorrect)
			},
		},
		{
			name:         "server error",
			expectedCode: http.StatusInternalServerError,
			method:       http.MethodPost,
			prepareMock: func(authenticator *mockAuthenticator) {
				authenticator.On("Authenticate", context.Background(), authRequest.Email, authRequest.Password).
					Return(authentication.TokenResponse{}, errors.New("some other error"))
			},
		},
		{
			name:         "incorrect HTTP method",
			expectedCode: http.StatusNotFound,
			method:       http.MethodGet,
			prepareMock: func(authenticator *mockAuthenticator) {
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			authenticator := new(mockAuthenticator)
			router := main.NewRouter(
				authenticator,
				new(mockAuthorizer),
				new(mockGroupService),
				new(mockUserService),
			)
			body, err := json.Marshal(&authRequest)
			require.NoError(t, err)
			req := httptest.NewRequest(test.method, "/authenticate", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()
			test.prepareMock(authenticator)

			// when
			router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}

func TestCreateGroup(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupRequest{
		Name:      "someName",
		CreatorID: 1,
	}
	expectedGroupResponse := expenses.GroupResponse{
		ID:   1,
		Name: groupRequest.Name,
		Users: []expenses.UserResponse{
			{
				ID:    1,
				Email: "some@mail.com",
			},
		},
	}
	groupService.On("Create", mock.Anything, groupRequest).Return(expectedGroupResponse, nil)

	body, err := json.Marshal(&groupRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var createdGroup expenses.GroupResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&createdGroup))
	assert.Equal(t, expectedGroupResponse, createdGroup)
}

func TestCreateGroupErrors(t *testing.T) {
	groupRequest := expenses.CreateGroupRequest{
		Name:      "someName",
		CreatorID: 1,
	}
	tests := []struct {
		name         string
		expectedCode int
		method       string
		prepareMock  func(service *mockGroupService)
	}{
		{
			name:         "creator not found",
			expectedCode: http.StatusBadRequest,
			method:       http.MethodPost,
			prepareMock: func(service *mockGroupService) {
				service.On("Create", mock.Anything, groupRequest).
					Return(expenses.GroupResponse{}, expenses.ErrUserNotFound)
			},
		},
		{
			name:         "group with such name exists",
			expectedCode: http.StatusBadRequest,
			method:       http.MethodPost,
			prepareMock: func(service *mockGroupService) {
				service.On("Create", mock.Anything, groupRequest).
					Return(expenses.GroupResponse{}, expenses.ErrGroupNameAlreadyExists)
			},
		},
		{
			name:         "user is already in a group",
			expectedCode: http.StatusBadRequest,
			method:       http.MethodPost,
			prepareMock: func(service *mockGroupService) {
				service.On("Create", mock.Anything, groupRequest).
					Return(expenses.GroupResponse{}, expenses.ErrUserIsInAnotherGroup)
			},
		},
		{
			name:         "incorrect method",
			expectedCode: http.StatusNotFound,
			method:       http.MethodGet,
			prepareMock: func(service *mockGroupService) {
			},
		},
		{
			name:         "server error on some other fail",
			expectedCode: http.StatusInternalServerError,
			method:       http.MethodPost,
			prepareMock: func(service *mockGroupService) {
				service.On("Create", mock.Anything, groupRequest).
					Return(expenses.GroupResponse{}, errors.New("some error"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			groupService := new(mockGroupService)
			router := main.NewRouter(
				new(mockAuthenticator),
				new(mockAuthorizer),
				groupService,
				new(mockUserService),
			)

			test.prepareMock(groupService)

			body, err := json.Marshal(&groupRequest)
			require.NoError(t, err)
			req := httptest.NewRequest(test.method, "/groups", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			// when
			router.ServeHTTP(recorder, req)

			// then
			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}

func TestCreateGroupWithoutAuthorizationWithJWTAuthorizerForbidden(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		authentication.NewJWTAuthorizer(jwt.HmacSha256("key"), new(mockTokenRetriever)),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupRequest{
		Name:      "someName",
		CreatorID: 1,
	}
	expectedGroupResponse := expenses.GroupResponse{
		ID:   1,
		Name: groupRequest.Name,
		Users: []expenses.UserResponse{
			{
				ID:    1,
				Email: "some@mail.com",
			},
		},
	}
	groupService.On("Create", mock.Anything, groupRequest).Return(expectedGroupResponse, nil)

	body, err := json.Marshal(&groupRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestCreateGroupWithAuthorizationWithJWTAuthorizerForbidden(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	tokenRetriever := new(mockTokenRetriever)
	alg := jwt.HmacSha256("key")
	tokenUUID, validJWT := prepareValidJWT(t, alg)
	router := main.NewRouter(
		new(mockAuthenticator),
		authentication.NewJWTAuthorizer(alg, tokenRetriever),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupRequest{
		Name:      "someName",
		CreatorID: 1,
	}
	expectedGroupResponse := expenses.GroupResponse{
		ID:   1,
		Name: groupRequest.Name,
		Users: []expenses.UserResponse{
			{
				ID:    1,
				Email: "some@mail.com",
			},
		},
	}
	groupService.On("Create", mock.Anything, groupRequest).Return(expectedGroupResponse, nil)
	tokenRetriever.On("Retrieve", tokenUUID).Return(authentication.UserContext{}, nil)

	body, err := json.Marshal(&groupRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+validJWT)
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var createdGroup expenses.GroupResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&createdGroup))
	assert.Equal(t, expectedGroupResponse, createdGroup)
}

func prepareValidJWT(t *testing.T, accessAlg *jwt.Algorithm) (string, string) {
	claims := jwt.NewClaims()
	accessUUID := "uuid-id"
	claims["access_uuid"] = accessUUID
	accessJWT, err := accessAlg.Encode(claims)
	require.NoError(t, err)
	return accessUUID, accessJWT
}
