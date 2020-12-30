package authentication_test

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/authentication"
	"go-spend/authentication/jwt"
	"go-spend/expenses"
	"testing"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(
	_ context.Context,
	_ pgxtype.Querier,
	_ expenses.CreateUserContext,
) (expenses.User, error) {
	panic("implement me")
}

func (m *mockUserRepository) FindById(_ context.Context, _ pgxtype.Querier, _ uint) (expenses.User, error) {
	panic("implement me")
}

func (m *mockUserRepository) FindByEmail(
	ctx context.Context,
	db pgxtype.Querier,
	email expenses.Email,
) (expenses.User, error) {
	args := m.Called(ctx, db, email)
	return args.Get(0).(expenses.User), args.Error(1)
}

type mockQuerier struct {
	mock.Mock
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	panic("implement me")
}

func (m *mockQuerier) Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error) {
	panic("implement me")
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row {
	panic("implement me")
}

func (m *mockQuerier) Begin(ctx context.Context) (pgx.Tx, error) {
	panic("implement me")
}

var (
	testTokenCreator      = authentication.NewTokenCreator(jwt.HmacSha256("acc"), jwt.HmacSha256("ref"))
	simplePasswordChecker = authentication.NoAcPasswordEncoder{}
)

func TestNewAuthService(t *testing.T) {
	auth := authentication.NewAuthService(testTokenCreator, new(mockUserRepository), new(mockQuerier), simplePasswordChecker)
	require.NotNil(t, auth)
}

func TestAuthReturnsTokensWhenCredAreCorrect(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	auth := authentication.NewAuthService(testTokenCreator, userRepository, mockDB, simplePasswordChecker)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	userRepository.On("FindByEmail", ctx, mockDB, email).
		Return(expenses.User{ID: 1, Email: email, Password: password}, nil)

	// when
	tokens, err := auth.Authenticate(ctx, email, password)
	require.NoError(t, err)
	require.NotZero(t, tokens)
}

func TestAuthReturnsErrorIfFailedToFindTheUser(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	auth := authentication.NewAuthService(testTokenCreator, userRepository, mockDB, simplePasswordChecker)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	userRepository.On("FindByEmail", ctx, mockDB, email).
		Return(expenses.User{}, errors.New("something went wrong"))

	// when
	tokens, err := auth.Authenticate(ctx, email, password)
	require.Error(t, err)
	require.Zero(t, tokens)
}

func TestAuthReturnsErrEmailOrPasswordIncorrectIfUserNotFound(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	auth := authentication.NewAuthService(testTokenCreator, userRepository, mockDB, simplePasswordChecker)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	userRepository.On("FindByEmail", ctx, mockDB, email).
		Return(expenses.User{}, expenses.ErrUserNotFound)

	// when
	tokens, err := auth.Authenticate(ctx, email, password)
	require.EqualError(t, err, authentication.ErrEmailOrPasswordIncorrect.Error())
	require.Zero(t, tokens)
}

func TestAuthReturnsErrEmailOrPasswordIncorrectIfPasswordIsIncorrect(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	auth := authentication.NewAuthService(testTokenCreator, userRepository, mockDB, simplePasswordChecker)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	userRepository.On("FindByEmail", ctx, mockDB, email).
		Return(expenses.User{ID: 1, Email: email, Password: "anotherPassword"}, nil)

	// when
	tokens, err := auth.Authenticate(ctx, email, password)
	require.EqualError(t, err, authentication.ErrEmailOrPasswordIncorrect.Error())
	require.Zero(t, tokens)
}
