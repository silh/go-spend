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
	Create(ctx context.Context, request CreateGroupRequest) (GroupResponse, error)
	// Find Group by its ID
	FindByID(ctx context.Context, id uint) (GroupResponse, error)
}

type DefaultGroupService struct {
	db              db.TxQuerier
	userRepository  UserRepository
	groupRepository GroupRepository
}

func NewDefaultGroupService(
	db db.TxQuerier,
	userRepository UserRepository,
	groupRepository GroupRepository,
) *DefaultGroupService {
	return &DefaultGroupService{db: db, userRepository: userRepository, groupRepository: groupRepository}
}

// Create creates a group and assigns group creator to that group.
func (d *DefaultGroupService) Create(ctx context.Context, request CreateGroupRequest) (GroupResponse, error) {
	id := request.CreatorID
	tx, err := d.db.Begin(ctx)
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Error("failed to rollback transaction - %s", err.Error())
		}
	}() // safe to do so even after commit according to docs
	if err != nil {
		return GroupResponse{}, err
	}
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
	group, err := d.groupRepository.FindByIDWithUsers(ctx, d.db, id)
	if err != nil {
		return GroupResponse{}, err
	}
	return group, nil
}
