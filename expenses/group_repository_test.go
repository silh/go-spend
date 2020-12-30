package expenses_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("gggg")
	created, err := repository.Create(ctx, pgdb, groupName)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, groupName, created.Name)
}

func TestFindGroupByID(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("gggg")
	created, _ := repository.Create(ctx, pgdb, groupName)
	found, err := repository.FindByID(ctx, pgdb, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, found)
}

func TestFindGroupByIDNonExistentGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	found, err := repository.FindByID(ctx, pgdb, 1)
	assert.EqualError(t, err, expenses.ErrGroupNotFound.Error())
	assert.Zero(t, found)
}

func TestCantCreateTwoGroupsWithTheSameName(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("myGroup")
	_, _ = repository.Create(ctx, pgdb, groupName)
	created2, err := repository.Create(ctx, pgdb, groupName)
	assert.EqualError(t, err, expenses.ErrGroupNameAlreadyExists.Error())
	assert.Zero(t, created2)
}

func TestAddUserToGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	// create user and group
	user, err := userRepository.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "some@mail.ru", Password: "12xczc"})
	require.NoError(t, err)
	groupName := util.NonEmptyString("myGroup")
	group, err := groupRepository.Create(ctx, pgdb, groupName)
	require.NoError(t, err)

	err = groupRepository.AddUserToGroup(ctx, pgdb, user.ID, group.ID)
	require.NoError(t, err)
}

func TestFindGroupByUserID(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	// create user and group, add user to group
	user, err := userRepository.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "some@mail.ru", Password: "12xczc"})
	require.NoError(t, err)
	groupName := util.NonEmptyString("myGroup")
	group, err := groupRepository.Create(ctx, pgdb, groupName)
	require.NoError(t, err)
	err = groupRepository.AddUserToGroup(ctx, pgdb, user.ID, group.ID)
	require.NoError(t, err)

	found, err := groupRepository.FindByUserID(ctx, pgdb, user.ID)
	require.NoError(t, err)
	assert.Equal(t, group, found)
}

func TestFindWithUsersByID(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	// create user and group, add user to group
	user, err := userRepository.Create(ctx, pgdb, expenses.CreateUserRequest{Email: "some@mail.ru", Password: "12xczc"})
	require.NoError(t, err)
	groupName := util.NonEmptyString("myGroup")
	group, err := groupRepository.Create(ctx, pgdb, groupName)
	require.NoError(t, err)
	err = groupRepository.AddUserToGroup(ctx, pgdb, user.ID, group.ID)
	require.NoError(t, err)

	// Find group with users and check result
	expectedUser := expenses.UserResponse{ID: user.ID, Email: user.Email}
	found, err := groupRepository.FindByIDWithUsers(ctx, pgdb, group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, found.ID)
	assert.Equal(t, group.Name, found.Name)
	require.Equal(t, 1, len(found.Users))
	assert.Equal(t, expectedUser, found.Users[0])
}
