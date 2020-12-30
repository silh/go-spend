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
	_ expenses.CreateUserRequest,
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

func (m *mockQuerier) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	panic("implement me")
}

func (m *mockQuerier) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	panic("implement me")
}

func (m *mockQuerier) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	panic("implement me")
}

func (m *mockQuerier) Begin(_ context.Context) (pgx.Tx, error) {
	panic("implement me")
}

type mockTokeSaver struct {
	mock.Mock
}

func (m *mockTokeSaver) Save(pair authentication.TokenPair, userContext authentication.UserContext) error {
	args := m.Called(pair, userContext)
	return args.Error(0)
}

var (
	testTokenCreator      = authentication.NewTokenCreator(jwt.HmacSha256("acc"), jwt.HmacSha256("ref"))
	simplePasswordChecker = authentication.NoAcPasswordEncoder{}
)

func TestNewAuthService(t *testing.T) {
	auth := authentication.NewAuthService(
		new(mockQuerier),
		testTokenCreator,
		new(mockTokeSaver),
		simplePasswordChecker,
		new(mockUserRepository),
	)
	require.NotNil(t, auth)
}

func TestAuthReturnsTokensWhenCredsAreCorrect(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	mockSaver := new(mockTokeSaver)
	auth := authentication.NewAuthService(
		mockDB,
		testTokenCreator,
		mockSaver,
		simplePasswordChecker,
		userRepository,
	)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	user := expenses.User{ID: 1, Email: email, Password: password}
	userRepository.On("FindByEmail", ctx, mockDB, email).Return(user, nil)
	mockSaver.On("Save", mock.Anything, authentication.UserContext{UserID: user.ID, GroupID: user.GroupID}).
		Return(nil)

	// when
	tokens, err := auth.Authenticate(ctx, email, password)

	// then
	require.NoError(t, err)
	require.NotZero(t, tokens)
}

func TestAuthReturnsErrorIfFailedToFindTheUser(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	mockSaver := new(mockTokeSaver)
	auth := authentication.NewAuthService(
		mockDB,
		testTokenCreator,
		mockSaver,
		simplePasswordChecker,
		userRepository,
	)

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
	mockSaver := new(mockTokeSaver)
	auth := authentication.NewAuthService(
		mockDB,
		testTokenCreator,
		mockSaver,
		simplePasswordChecker,
		userRepository,
	)

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
	mockSaver := new(mockTokeSaver)
	auth := authentication.NewAuthService(
		mockDB,
		testTokenCreator,
		mockSaver,
		simplePasswordChecker,
		userRepository,
	)

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

func TestAuthReturnsErrWhenTokenSavingFails(t *testing.T) {
	ctx := context.Background()
	userRepository := new(mockUserRepository)
	mockDB := new(mockQuerier)
	mockSaver := new(mockTokeSaver)
	auth := authentication.NewAuthService(
		mockDB,
		testTokenCreator,
		mockSaver,
		simplePasswordChecker,
		userRepository,
	)

	// given
	email := expenses.Email("some@mail.com")
	password := expenses.Password("password")
	user := expenses.User{ID: 1, Email: email, Password: password}
	userRepository.On("FindByEmail", ctx, mockDB, email).Return(user, nil)
	mockSaver.On("Save", mock.Anything, authentication.UserContext{UserID: user.ID, GroupID: user.GroupID}).
		Return(errors.New("expected"))

	// when
	_, err := auth.Authenticate(ctx, email, password)

	// then
	require.Error(t, err)
}
