package expenses_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"go-spend/util"
	"testing"
)

const (
	deleteAllGroupsQuery = "DELETE FROM groups"
)

func TestCreateGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("gggg")
	created, err := repository.Create(ctx, pgDB, groupName)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, groupName, created.Name)
}

func TestFindGroupByID(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("gggg")
	created, _ := repository.Create(ctx, pgDB, groupName)
	found, err := repository.FindByID(ctx, pgDB, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, found)
}

func TestFindGroupByIDNonExistentGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	found, err := repository.FindByID(ctx, pgDB, 1)
	assert.EqualError(t, err, expenses.ErrGroupNotFound.Error())
	assert.Zero(t, found)
}

func TestCantCreateTwoGroupsWithTheSameName(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	repository := expenses.NewPgGroupRepository()
	groupName := util.NonEmptyString("myGroup")
	_, _ = repository.Create(ctx, pgDB, groupName)
	created2, err := repository.Create(ctx, pgDB, groupName)
	assert.EqualError(t, err, expenses.ErrNameAlreadyExists.Error())
	assert.Zero(t, created2)
}

func TestAddUserToGroup(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	// create user and group
	user, err := userRepository.Create(ctx, pgDB, expenses.CreateUserRequest{Email: "some@mail.ru", Password: "12xczc"})
	require.NoError(t, err)
	groupName := util.NonEmptyString("myGroup")
	group, err := groupRepository.Create(ctx, pgDB, groupName)
	require.NoError(t, err)

	err = groupRepository.AddUserToGroup(ctx, pgDB, user.ID, group.ID)
	require.NoError(t, err)
}

func TestFindGroupByUserID(t *testing.T) {
	ctx := context.Background()
	cleanUpDB(t, ctx)

	userRepository := expenses.NewPgUserRepository()
	groupRepository := expenses.NewPgGroupRepository()

	// create user and group, add user to group
	user, err := userRepository.Create(ctx, pgDB, expenses.CreateUserRequest{Email: "some@mail.ru", Password: "12xczc"})
	require.NoError(t, err)
	groupName := util.NonEmptyString("myGroup")
	group, err := groupRepository.Create(ctx, pgDB, groupName)
	require.NoError(t, err)
	err = groupRepository.AddUserToGroup(ctx, pgDB, user.ID, group.ID)
	require.NoError(t, err)

	found, err := groupRepository.FindByUserID(ctx, pgDB, user.ID)
	require.NoError(t, err)
	assert.Equal(t, group, found)
}
