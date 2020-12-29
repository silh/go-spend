package expenses_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, request expenses.CreateUserContext) (expenses.UserResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.UserResponse), args.Error(1)
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

func TestNewRouter(t *testing.T) {
	router := expenses.NewRouter(new(mockUserService), new(mockGroupService))
	assert.NotNil(t, router)
}

func TestCreateUserWithProperParams(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := expenses.NewRouter(userService, new(mockGroupService))

	createUserRequest := expenses.RawCreateUserRequest{Email: "some@mail.com", Password: "1234"}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()
	userResponse := expenses.UserResponse{ID: 1, Email: expenses.Email(createUserRequest.Email)}
	userService.On(
		"Create",
		mock.Anything,
		mock.AnythingOfType("expenses.CreateUserContext"),
	).Return(userResponse, nil)

	// when
	router.GetMux().ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var result expenses.UserResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&result))
	assert.Equal(t, userResponse, result)
}

func TestCreateUserWithIncorrectMethod(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := expenses.NewRouter(userService, new(mockGroupService))

	createUserRequest := expenses.CreateUserContext{Email: "some@mail.com", Password: "1234"}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.GetMux().ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestCreateUserWithIncorrectBody(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := expenses.NewRouter(userService, new(mockGroupService))

	createUserRequest := struct {
		something int64
	}{something: 10}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.GetMux().ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateUserWithEmptyFields(t *testing.T) {
	// given
	userService := new(mockUserService)
	router := expenses.NewRouter(userService, new(mockGroupService))

	createUserRequest := struct {
		something int64
	}{something: 10}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	// when
	router.GetMux().ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateUserWithSomeIncorrectFields(t *testing.T) {
	tests := []struct {
		name string
		body expenses.CreateUserContext
	}{
		{
			name: "no email",
			body: expenses.CreateUserContext{Password: "1234"},
		},
		{
			name: "no password",
			body: expenses.CreateUserContext{Email: "some@email.com"},
		},
		{
			name: "invalid email",
			body: expenses.CreateUserContext{Email: "someemail.com", Password: "12341"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			userService := new(mockUserService)
			router := expenses.NewRouter(userService, new(mockGroupService))
			jsonBody, err := json.Marshal(&test.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
			recorder := httptest.NewRecorder()

			// when
			router.GetMux().ServeHTTP(recorder, req)

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
					mock.AnythingOfType("expenses.CreateUserContext"),
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
					mock.AnythingOfType("expenses.CreateUserContext"),
				).Return(expenses.UserResponse{}, errors.New("expected"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			userService := new(mockUserService)
			router := expenses.NewRouter(userService, new(mockGroupService))

			createUserRequest := expenses.RawCreateUserRequest{Email: "some@mail.com", Password: "1234"}
			jsonBody, err := json.Marshal(createUserRequest)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
			recorder := httptest.NewRecorder()
			test.prepareMock(userService)
			// when
			router.GetMux().ServeHTTP(recorder, req)

			// then
			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}
