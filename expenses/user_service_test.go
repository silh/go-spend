package expenses_test

import (
	"context"
	"errors"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

const (
	validEmail = "email@mail.com"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, _ pgxtype.Querier, request expenses.CreateUserContext) (expenses.User, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(expenses.User), args.Error(1)
}

func (m *MockUserRepository) FindById(_ context.Context, _ pgxtype.Querier, _ uint) (expenses.User, error) {
	panic("implement me")
}

func TestNewDefaultUserService(t *testing.T) {
	service := expenses.NewDefaultUserService(new(MockTxQuerier), new(MockUserRepository))
	assert.NotNil(t, service)
}

func TestDefaultUserServiceCreate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := expenses.NewDefaultUserService(new(MockTxQuerier), mockRepo)

	ctx := context.Background()
	request := expenses.CreateUserContext{Email: validEmail, Password: "123"}
	createdUser := expenses.User{ID: 1, Email: validEmail, Password: "123"}
	mockRepo.On("Create", ctx, request).Return(createdUser, nil)

	expected := expenses.UserResponse{ID: createdUser.ID, Email: createdUser.Email}

	actual, err := service.Create(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, actual, expected)
}

func TestDefaultUserServiceCreateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := expenses.NewDefaultUserService(new(MockTxQuerier), mockRepo)

	ctx := context.Background()
	request := expenses.CreateUserContext{Email: validEmail, Password: "123"}
	expectedError := errors.New("db is not accessible")
	mockRepo.On("Create", ctx, request).Return(expenses.User{}, expectedError)

	actual, err := service.Create(ctx, request)
	assert.Zero(t, actual)
	assert.EqualError(t, err, expectedError.Error())
}
