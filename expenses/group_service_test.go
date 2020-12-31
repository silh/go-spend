package expenses_test

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"testing"
)

const (
	validEmail = "email@mail.com"
)

type mockGroupRepository struct {
	mock.Mock
}

func (m *mockGroupRepository) Create(ctx context.Context, db pgxtype.Querier, group util.NonEmptyString) (expenses.Group, error) {
	args := m.Called(ctx, db, group)
	return args.Get(0).(expenses.Group), args.Error(1)
}

func (m *mockGroupRepository) FindByID(ctx context.Context, db pgxtype.Querier, id uint) (expenses.Group, error) {
	args := m.Called(ctx, db, id)
	return args.Get(0).(expenses.Group), args.Error(1)
}

func (m *mockGroupRepository) FindByIDWithUsers(ctx context.Context, db pgxtype.Querier, id uint) (expenses.GroupResponse, error) {
	args := m.Called(ctx, db, id)
	return args.Get(0).(expenses.GroupResponse), args.Error(1)
}

func (m *mockGroupRepository) FindByUserID(_ context.Context, _ pgxtype.Querier, _ uint) (expenses.Group, error) {
	panic("implement me")
}

func (m *mockGroupRepository) AddUserToGroup(ctx context.Context, db pgxtype.Querier, userID uint, groupID uint) error {
	args := m.Called(ctx, db, userID, groupID)
	return args.Error(0)
}

type mockTx struct {
	mock.Mock
}

func (m *mockTx) Begin(_ context.Context) (pgx.Tx, error) {
	panic("implement me")
}

func (m *mockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockTx) Rollback(_ context.Context) error {
	return nil
}

func (m *mockTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	panic("implement me")
}

func (m *mockTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults {
	panic("implement me")
}

func (m *mockTx) LargeObjects() pgx.LargeObjects {
	panic("implement me")
}

func (m *mockTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	panic("implement me")
}

func (m *mockTx) Exec(_ context.Context, _ string, _ ...interface{}) (commandTag pgconn.CommandTag, err error) {
	panic("implement me")
}

func (m *mockTx) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	panic("implement me")
}

func (m *mockTx) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	panic("implement me")
}

func (m *mockTx) QueryFunc(_ context.Context, _ string, _ []interface{}, _ []interface{}, _ func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	panic("implement me")
}

func (m *mockTx) Conn() *pgx.Conn {
	panic("implement me")
}

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

func (m *mockUserRepository) FindById(ctx context.Context, db pgxtype.Querier, id uint) (expenses.User, error) {
	args := m.Called(ctx, db, id)
	return args.Get(0).(expenses.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(
	ctx context.Context,
	db pgxtype.Querier,
	email expenses.Email,
) (expenses.User, error) {
	args := m.Called(ctx, db, email)
	return args.Get(0).(expenses.User), args.Error(1)
}

func TestNewDefaultGroupService(t *testing.T) {
	groupService := expenses.NewDefaultGroupService(pgdb, expenses.NewPgUserRepository(), expenses.NewPgGroupRepository())
	require.NotNil(t, groupService)
}

// This is an integration test as we need to check the storage of multiple things in transaction
func TestCreateGroupWithCreator(t *testing.T) {
	// given
	ctx := context.Background()

	userRepository := expenses.NewPgUserRepository()
	groupService := expenses.NewDefaultGroupService(pgdb, userRepository, expenses.NewPgGroupRepository())

	// Create a user so that it can create a group
	user, err := userRepository.Create(ctx, pgdb, expenses.CreateUserRequest{Email: validEmail, Password: "12314"})
	require.NoError(t, err)
	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: user.ID}

	// when
	createdGroup, err := groupService.Create(ctx, createGroupRequest)
	require.NoError(t, err)
	assert.NotZero(t, createdGroup)
	assert.NotZero(t, createdGroup.ID)
	assert.Equal(t, createGroupRequest.Name, createdGroup.Name)

	// then
	expectedUser := expenses.UserResponse{ID: user.ID, Email: user.Email}
	foundGroup, err := groupService.FindByID(ctx, createdGroup.ID)
	require.NoError(t, err)
	assert.NotZero(t, foundGroup)
	require.Equal(t, 1, len(foundGroup.Users))
	assert.Equal(t, expectedUser, foundGroup.Users[0])
}

func TestCreateGroupFailedToStartTx(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)

	db.On("Begin", ctx).Return(nil, errors.New("expected"))

	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: 1}

	// when
	_, err := groupService.Create(ctx, createGroupRequest)

	//then
	require.Error(t, err)
}

func TestCreateGroupFailedToFindUser(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	tx := new(mockTx)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	db.On("Begin", ctx).Return(tx, nil)
	userRepository.On("FindById", ctx, tx, uint(1)).Return(expenses.User{}, errors.New("expected"))

	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: 1}

	// when
	_, err := groupService.Create(ctx, createGroupRequest)
	// then
	require.Error(t, err)
}

func TestCreateGroupFailedToCreate(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	tx := new(mockTx)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	db.On("Begin", ctx).Return(tx, nil)
	user := expenses.User{ID: 1}
	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: 1}
	userRepository.On("FindById", ctx, tx, uint(1)).Return(user, nil)
	groupRepository.On("Create", ctx, tx, createGroupRequest.Name).
		Return(expenses.Group{}, errors.New("expected"))

	// when
	_, err := groupService.Create(ctx, createGroupRequest)
	// then
	require.Error(t, err)
}

func TestCreateGroupFailedAddUser(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	tx := new(mockTx)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	db.On("Begin", ctx).Return(tx, nil)
	user := expenses.User{ID: 1}
	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: 1}
	group := expenses.Group{ID: 1, Name: createGroupRequest.Name}
	userRepository.On("FindById", ctx, tx, uint(1)).Return(user, nil)
	groupRepository.On("Create", ctx, tx, createGroupRequest.Name).
		Return(group, nil)
	groupRepository.On("AddUserToGroup", ctx, tx, user.ID, group.ID).Return(errors.New("expected"))

	// when
	_, err := groupService.Create(ctx, createGroupRequest)
	// then
	require.Error(t, err)
}

func TestCreateGroupFailedToCommitTx(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	tx := new(mockTx)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	db.On("Begin", ctx).Return(tx, nil)
	user := expenses.User{ID: 1}
	createGroupRequest := expenses.CreateGroupRequest{Name: "name", CreatorID: 1}
	group := expenses.Group{ID: 1, Name: createGroupRequest.Name}
	userRepository.On("FindById", ctx, tx, uint(1)).Return(user, nil)
	groupRepository.On("Create", ctx, tx, createGroupRequest.Name).
		Return(group, nil)
	groupRepository.On("AddUserToGroup", ctx, tx, user.ID, group.ID).Return(nil)
	tx.On("Commit", ctx).Return(errors.New("expected"))

	// when
	_, err := groupService.Create(ctx, createGroupRequest)
	// then
	require.Error(t, err)
}

func TestGroupByID(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	id := uint(100)
	expectedGroup := expenses.GroupResponse{ID: id, Name: "some", Users: []expenses.UserResponse{}}
	groupRepository.On("FindByIDWithUsers", ctx, db, id).Return(expectedGroup, nil)

	// when
	groupResponse, err := groupService.FindByID(ctx, id)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedGroup, groupResponse)
}

func TestDefaultGroupServiceAddUserToGroup(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	addToGroupRequest := expenses.AddToGroupRequest{
		UserID:  11123,
		GroupID: 214,
	}
	groupRepository.On("AddUserToGroup", ctx, db, addToGroupRequest.UserID, addToGroupRequest.GroupID).
		Return(nil)

	// when
	err := groupService.AddUserToGroup(ctx, addToGroupRequest)

	// then
	require.NoError(t, err)
}

func TestDefaultGroupServiceAddUserToGroupErrorPropagated(t *testing.T) {
	// given
	ctx := context.Background()

	db := new(mockTxQuerier)
	userRepository := new(mockUserRepository)
	groupRepository := new(mockGroupRepository)
	groupService := expenses.NewDefaultGroupService(db, userRepository, groupRepository)
	addToGroupRequest := expenses.AddToGroupRequest{
		UserID:  11123,
		GroupID: 214,
	}
	groupRepository.On("AddUserToGroup", ctx, db, addToGroupRequest.UserID, addToGroupRequest.GroupID).
		Return(errors.New("expected"))

	// when
	err := groupService.AddUserToGroup(ctx, addToGroupRequest)

	// then
	require.Error(t, err)
}
