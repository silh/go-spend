package expenses

import (
	"context"
	"github.com/jackc/pgx/v4"
	"go-spend/db"
	"go-spend/log"
)

// Perform operations with groups of Users
type GroupService interface {
	// Create a new group and add its creator to the group
	Create(ctx context.Context, request CreateGroupContext) (GroupResponse, error)
	// Find Group by its ID
	FindByID(ctx context.Context, id uint) (GroupResponse, error)
	// AddUserToGroup adds user to an existing group
	AddUserToGroup(ctx context.Context, addRequest AddToGroupRequest) error
}

// DefaultGroupService is default implementation of GroupService. If fetches data through UserRepository and
// GroupRepository
type DefaultGroupService struct {
	db              db.TxQuerier
	userRepository  UserRepository
	groupRepository GroupRepository
}

// NewDefaultGroupService creates new instance of DefaultGroupService
func NewDefaultGroupService(
	db db.TxQuerier,
	userRepository UserRepository,
	groupRepository GroupRepository,
) *DefaultGroupService {
	return &DefaultGroupService{db: db, userRepository: userRepository, groupRepository: groupRepository}
}

// Create creates a group and assigns group creator to that group.
// If creator doesn't exist - returns ErrUserNotFound
// If creator is in another group - returns ErrUserIsInAnotherGroup
// If group with such name exists - returns ErrGroupNameAlreadyExists
func (d *DefaultGroupService) Create(ctx context.Context, request CreateGroupContext) (GroupResponse, error) {
	id := request.CreatorID
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return GroupResponse{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Error("failed to rollback transaction - %s", err.Error())
		}
	}() // safe to do so even after commit according to docs
	creator, err := d.userRepository.FindById(ctx, tx, id)
	if err != nil {
		return GroupResponse{}, err
	}
	group, err := d.groupRepository.Create(ctx, tx, request.Name)
	if err != nil {
		return GroupResponse{}, err
	}
	if err = d.groupRepository.AddUserToGroup(ctx, tx, creator.ID, group.ID); err != nil {
		return GroupResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return GroupResponse{}, err
	}
	return GroupResponse{
		ID:   group.ID,
		Name: group.Name,
		Users: []UserResponse{
			{
				ID:    creator.ID,
				Email: creator.Email,
			},
		},
	}, nil
}

func (d *DefaultGroupService) FindByID(ctx context.Context, id uint) (GroupResponse, error) {
	return d.groupRepository.FindByIDWithUsers(ctx, d.db, id)
}

func (d *DefaultGroupService) AddUserToGroup(ctx context.Context, addRequest AddToGroupRequest) error {
	return d.groupRepository.AddUserToGroup(ctx, d.db, addRequest.UserID, addRequest.GroupID)
}
