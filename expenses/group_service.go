package expenses

import (
	"context"
	"go-spend/db"
	"go-spend/log"
)

// Perform operations with groups of Users
type GroupService interface {
	// Create a new group
	Create(ctx context.Context, request CreateGroupRequest) (Group, error)
}

type DefaultGroupService struct {
	userRepository  UserRepository
	groupRepository GroupRepository
	txProvider      db.TxQuerier
}

func NewDefaultGroupService(userRepository UserRepository, groupRepository GroupRepository) *DefaultGroupService {
	return &DefaultGroupService{userRepository: userRepository, groupRepository: groupRepository}
}

func (d *DefaultGroupService) Create(ctx context.Context, request CreateGroupRequest) (Group, error) {
	id := request.CreatorID
	tx, err := d.txProvider.Begin(ctx)
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error("failed to rollback transaction - %s", err.Error()) // TODO check for already committed transaction.
		}
	}() // safe to do so even after commit according to docs
	if err != nil {
		return Group{}, err
	}
	creator, err := d.userRepository.FindById(ctx, tx, id)
	if err != nil {
		return Group{}, err
	}
	group, err := d.groupRepository.Create(ctx, tx, request.Name)
	if err != nil {
		return Group{}, err
	}
	if err = d.groupRepository.AddUserToGroup(ctx, tx, creator.ID, group.ID); err != nil {
		return Group{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Group{}, err
	}
	return group, nil
}
