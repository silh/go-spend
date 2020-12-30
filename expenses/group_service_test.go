package expenses_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-spend/expenses"
	"testing"
)

func TestNewDefaultGroupService(t *testing.T) {
	groupService := expenses.NewDefaultGroupService(PGDB, expenses.NewPgUserRepository(), expenses.NewPgGroupRepository())
	require.NotNil(t, groupService)
}

// This is an integration test as we need to check the storage of multiple things in transaction
func TestCreateGroupWithCreator(t *testing.T) {
	// given
	ctx := context.Background()

	userRepository := expenses.NewPgUserRepository()
	groupService := expenses.NewDefaultGroupService(PGDB, userRepository, expenses.NewPgGroupRepository())

	// Create a user so that it can create a group
	user, err := userRepository.Create(ctx, PGDB, expenses.CreateUserRequest{Email: validEmail, Password: "12314"})
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
