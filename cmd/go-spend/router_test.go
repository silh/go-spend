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
	"time"
)

var (
	defaultUserContextForCreate = authentication.UserContext{
		UserID:  1,
		GroupID: 0,
	}
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

func (m *mockGroupService) Create(ctx context.Context, request expenses.CreateGroupContext) (expenses.GroupResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.GroupResponse), args.Error(1)
}

func (m *mockGroupService) FindByID(_ context.Context, _ uint) (expenses.GroupResponse, error) {
	panic("implement me")
}

func (m *mockGroupService) AddUserToGroup(ctx context.Context, addRequest expenses.AddToGroupRequest) error {
	args := m.Called(ctx, addRequest)
	return args.Error(0)
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

type mockExpensesService struct {
	mock.Mock
}

func (m *mockExpensesService) Create(
	ctx context.Context,
	newExpense expenses.CreateExpenseContext,
) (expenses.ExpenseResponse, error) {
	args := m.Called(ctx, newExpense)
	return args.Get(0).(expenses.ExpenseResponse), args.Error(1)
}

func TestNewRouter(t *testing.T) {
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
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
		new(mockExpensesService),
		new(mockGroupService),
		userService,
	)

	createUserRequest := expenses.CreateUserRequest{Email: "some@mail.com", Password: "1234"}
	jsonBody, err := json.Marshal(createUserRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()
	userResponse := expenses.UserResponse{ID: 1, Email: createUserRequest.Email}
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
		new(mockExpensesService),
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
		new(mockExpensesService),
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
		new(mockExpensesService),
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
				new(mockExpensesService),
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
				new(mockExpensesService),
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
		new(mockExpensesService),
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
				new(mockExpensesService),
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

func TestRouterCreateGroup(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupRequest{
		Name: "someName",
	}
	groupContext := expenses.CreateGroupContext{
		Name:      "someName",
		CreatorID: defaultUserContextForCreate.UserID,
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
	groupService.On("Create", mock.Anything, groupContext).Return(expectedGroupResponse, nil)

	body, err := json.Marshal(&groupRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), "user", defaultUserContextForCreate))
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
	groupRequest := expenses.CreateGroupContext{
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
				new(mockExpensesService),
				groupService,
				new(mockUserService),
			)

			test.prepareMock(groupService)

			body, err := json.Marshal(&groupRequest)
			require.NoError(t, err)
			req := httptest.NewRequest(test.method, "/groups", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), "user", defaultUserContextForCreate))
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
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupContext{
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

func TestCreateGroupWithAuthorizationWithJWTAuthorizerCreated(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	tokenRetriever := new(mockTokenRetriever)
	alg := jwt.HmacSha256("key")
	tokenUUID, validJWT := prepareValidJWT(t, alg)
	router := main.NewRouter(
		new(mockAuthenticator),
		authentication.NewJWTAuthorizer(alg, tokenRetriever),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)

	groupRequest := expenses.CreateGroupContext{
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
	tokenRetriever.On("Retrieve", tokenUUID).Return(defaultUserContextForCreate, nil)

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

func TestCreateExpense(t *testing.T) {
	// given
	expensesService := new(mockExpensesService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		expensesService,
		new(mockGroupService),
		new(mockUserService),
	)

	expenseRequest := expenses.CreateExpenseRequest{
		Amount: 100.10,
		Shares: expenses.ExpenseShares{
			1: 100,
		},
	}
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 1,
	}

	body, err := json.Marshal(&expenseRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()
	expectedResponse := expenses.ExpenseResponse{
		UserID:    1,
		Amount:    expenseRequest.Amount,
		Timestamp: time.Now(),
		Shares:    expenseRequest.Shares,
	}

	checkFunc := func(ctx expenses.CreateExpenseContext) bool {
		return ctx.UserID == userContext.UserID && ctx.GroupID == userContext.GroupID
	}
	expensesService.On("Create", mock.Anything, mock.MatchedBy(checkFunc)).Return(expectedResponse, nil)

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var response expenses.ExpenseResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	assert.InDelta(t, expenseRequest.Amount, response.Amount, 0.01)
	assert.Equal(t, userContext.UserID, expectedResponse.UserID)
	assert.NotZero(t, response.Timestamp)
}

func TestCreateExpenseIncorrectBody(t *testing.T) {
	// given
	expensesService := new(mockExpensesService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		expensesService,
		new(mockGroupService),
		new(mockUserService),
	)

	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 1,
	}

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer([]byte("body")))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// This one should not happen with proper setup
func TestCreateExpenseNoUserForbidden(t *testing.T) {
	// given
	expensesService := new(mockExpensesService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		expensesService,
		new(mockGroupService),
		new(mockUserService),
	)

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer([]byte("body")))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestCreateExpenseNonProperContextBadRequest(t *testing.T) {
	// given
	expensesService := new(mockExpensesService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		expensesService,
		new(mockGroupService),
		new(mockUserService),
	)

	expenseRequest := expenses.CreateExpenseRequest{
		Amount: 100.10,
		Shares: expenses.ExpenseShares{
			1: 95,
			2: 1,
		},
	}
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 1,
	}

	body, err := json.Marshal(&expenseRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestCreateExpenseServiceFailsServerError(t *testing.T) {
	// given
	expensesService := new(mockExpensesService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		expensesService,
		new(mockGroupService),
		new(mockUserService),
	)

	expenseRequest := expenses.CreateExpenseRequest{
		Amount: 100.10,
		Shares: expenses.ExpenseShares{
			1: 100,
		},
	}
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 1,
	}

	body, err := json.Marshal(&expenseRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBuffer(body))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	checkFunc := func(ctx expenses.CreateExpenseContext) bool {
		return ctx.UserID == userContext.UserID && ctx.GroupID == userContext.GroupID
	}
	expensesService.On("Create", mock.Anything, mock.MatchedBy(checkFunc)).
		Return(expenses.ExpenseResponse{}, errors.New("expected"))

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestAddToGroup(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 22,
	}
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	groupService.On("AddUserToGroup", reqWithContext.Context(), addRequest).Return(nil)

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestAddToGroupForbiddenWithoutUserContext(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAddToGroupForbiddenWithWrongStuffInContext(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	// this should never happen
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", "rrrr"))
	req = req.WithContext(context.WithValue(req.Context(), "user", defaultUserContextForCreate))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAddToGroupIncorrectBodyBadRequest(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 22,
	}
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer([]byte("data")))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAddToGroupDifferentGroupOfCallerForbidden(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 10,
	}
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAddToGroupErrUserOrGroupNotFoundBadRequest(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 22,
	}
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	groupService.On("AddUserToGroup", reqWithContext.Context(), addRequest).
		Return(expenses.ErrUserOrGroupNotFound)

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAddToGroupErrUserIsInAnotherGroupBadRequest(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 22,
	}
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	groupService.On("AddUserToGroup", reqWithContext.Context(), addRequest).
		Return(expenses.ErrUserIsInAnotherGroup)

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAddToGroupOtherErrServerError(t *testing.T) {
	// given
	groupService := new(mockGroupService)
	router := main.NewRouter(
		new(mockAuthenticator),
		new(mockAuthorizer),
		new(mockExpensesService),
		groupService,
		new(mockUserService),
	)
	// that is done by authorizer in real app
	userContext := authentication.UserContext{
		UserID:  1,
		GroupID: 22,
	}
	addRequest := expenses.AddToGroupRequest{
		UserID:  2,
		GroupID: 22,
	}
	data, err := json.Marshal(&addRequest)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/groups", bytes.NewBuffer(data))
	reqWithContext := req.WithContext(context.WithValue(req.Context(), "user", userContext))
	recorder := httptest.NewRecorder()

	groupService.On("AddUserToGroup", reqWithContext.Context(), addRequest).
		Return(errors.New("expected"))

	// when
	router.ServeHTTP(recorder, reqWithContext)

	// then
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func prepareValidJWT(t *testing.T, accessAlg *jwt.Algorithm) (string, string) {
	claims := jwt.NewClaims()
	accessUUID := "uuid-id"
	claims["access_uuid"] = accessUUID
	accessJWT, err := accessAlg.Encode(claims)
	require.NoError(t, err)
	return accessUUID, accessJWT
}
