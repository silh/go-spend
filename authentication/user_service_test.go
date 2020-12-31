package authentication_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/expenses"
	"testing"
)

const (
	validEmail = "email@mail.com"
)

func TestNewDefaultUserService(t *testing.T) {
	service := authentication.NewDefaultUserService(
		new(mockQuerier),
		&authentication.NoAcPasswordEncoder{},
		new(mockUserRepository),
	)
	assert.NotNil(t, service)
}

func TestDefaultUserServiceCreate(t *testing.T) {
	mockRepo := new(mockUserRepository)
	db := new(mockQuerier)
	service := authentication.NewDefaultUserService(db, simplePasswordChecker, mockRepo)

	ctx := context.Background()
	request := expenses.CreateUserRequest{Email: validEmail, Password: "123"}
	createdUser := expenses.User{ID: 1, Email: validEmail, Password: "123"}
	mockRepo.On("Create", ctx, db, request).Return(createdUser, nil)

	expected := expenses.UserResponse{ID: createdUser.ID, Email: createdUser.Email}

	actual, err := service.Create(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDefaultUserServiceCreateError(t *testing.T) {
	mockRepo := new(mockUserRepository)
	db := new(mockQuerier)
	service := authentication.NewDefaultUserService(db, simplePasswordChecker, mockRepo)

	ctx := context.Background()
	request := expenses.CreateUserRequest{Email: validEmail, Password: "123"}
	expectedError := errors.New("db is not accessible")
	mockRepo.On("Create", ctx, db, request).Return(expenses.User{}, expectedError)

	actual, err := service.Create(ctx, request)
	assert.Zero(t, actual)
	assert.EqualError(t, err, expectedError.Error())
}

func TestDefaultUserServiceCreateWithBCrypt(t *testing.T) {
	mockRepo := new(mockUserRepository)
	db := new(mockQuerier)
	passwordEncoder := &authentication.BCryptPasswordEncoder{}
	service := authentication.NewDefaultUserService(db, passwordEncoder, mockRepo)

	ctx := context.Background()
	request := expenses.CreateUserRequest{Email: validEmail, Password: "123"}
	createdUser := expenses.User{ID: 1, Email: validEmail}
	checkFunc := func(userReq expenses.CreateUserRequest) bool {
		createdUser.Password = userReq.Password // set it here after encoding
		return passwordEncoder.Check(string(userReq.Password), string(request.Password))
	}
	mockRepo.On("Create", ctx, db, mock.MatchedBy(checkFunc)).Return(createdUser, nil)

	expected := expenses.UserResponse{ID: createdUser.ID, Email: createdUser.Email}

	actual, err := service.Create(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, actual, expected)
}
