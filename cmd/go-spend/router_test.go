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

type mockGroupService struct {
	mock.Mock
}

type mockAuthenticator struct {
	mock.Mock
}

func (m *mockAuthenticator) Authenticate(
	ctx context.Context,
	email expenses.Email,
	password expenses.Password,
) (authentication.TokenResponse, error) {
	panic("implement me")
}

func (m *mockGroupService) Create(ctx context.Context, request expenses.CreateGroupRequest) (expenses.GroupResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.GroupResponse), args.Error(1)
}

func (m *mockGroupService) FindByID(_ context.Context, _ uint) (expenses.GroupResponse, error) {
	panic("implement me")
}

func TestNewRouter(t *testing.T) {
	router := main.NewRouter(new(mockUserService), new(mockGroupService), new(mockAuthenticator))
	assert.NotNil(t, router)
}

func TestCreateUserWithProperParams(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))

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
	router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))

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
	router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))

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
	router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))

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
			router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))
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

func TestServiceError(t *testing.T) {
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
			router := main.NewRouter(userService, new(mockGroupService), new(mockAuthenticator))

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
