package expenses

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"go-spend/util"
)

// Operation related to Group storage
type GroupRepository interface {
	// Store a new Group
	Create(ctx context.Context, db pgxtype.Querier, group util.NonEmptyString) (Group, error)
	// Find Group by its ID
	FindByID(ctx context.Context, db pgxtype.Querier, id uint) (Group, error)
	// Find Group by its ID with Users in this group
	FindByIDWithUsers(ctx context.Context, db pgxtype.Querier, id uint) (GroupResponse, error)
	// Find Group by one of its members ids. As we allow only one group per user that should return 0-1 results.
	FindByUserID(ctx context.Context, db pgxtype.Querier, userID uint) (Group, error)
	// Add User to an existing group. If User with such provided ID doesn't exists or Group with such ID doesn't exist an
	// error will be returned
	AddUserToGroup(ctx context.Context, db pgxtype.Querier, userID uint, groupID uint) error
}

const (
	createGroupQuery            = "INSERT INTO groups (name) VALUES ($1) RETURNING id"
	addUserToGroup              = "INSERT INTO users_groups (user_id, group_id) VALUES ($1, $2)"
	findGroupByIDQuery          = "SELECT g.id, g.name FROM groups as g WHERE g.id = $1"
	findGroupByIDWithUsersQuery = "SELECT g.id, g.name, u.id, u.email " +
		"FROM groups as g " +
		"JOIN users_groups as ug on g.id = ug.group_id " +
		"JOIN users as u on ug.user_id = u.id " +
		"WHERE g.id = $1"
	findGroupByUserIDQuery = "SELECT g.id, g.name " +
		"FROM groups as g " +
		"JOIN users_groups as ug ON g.id = ug.group_id " +
		"WHERE ug.user_id = $1"
)

var (
	ErrGroupNameAlreadyExists = errors.New("group with such name already exists")
	ErrGroupNotFound          = errors.New("group not found")
	ErrUserIsInAnotherGroup   = errors.New("user is already in another group")
)

// Group repository that stores info in Postgres DB
type PgGroupRepository struct {
}

// NewPgGroupRepository creates new PG repository
func NewPgGroupRepository() *PgGroupRepository {
	return &PgGroupRepository{}
}

func (p *PgGroupRepository) Create(ctx context.Context, db pgxtype.Querier, groupName util.NonEmptyString) (Group, error) {
	createdGroup := Group{Name: groupName}
	if err := db.QueryRow(ctx, createGroupQuery, groupName).Scan(&createdGroup.ID); err != nil {
		if pfError, ok := err.(*pgconn.PgError); ok && pfError.Code == uniqueViolation {
			return Group{}, ErrGroupNameAlreadyExists
		}
		return Group{}, err
	}
	return createdGroup, nil
}

func (p *PgGroupRepository) AddUserToGroup(ctx context.Context, db pgxtype.Querier, userID uint, groupID uint) error {
	if _, err := db.Exec(ctx, addUserToGroup, userID, groupID); err != nil {
		if pfError, ok := err.(*pgconn.PgError); ok && pfError.Code == uniqueViolation {
			return ErrUserIsInAnotherGroup
		}
		return err
	}
	return nil
}

func (p *PgGroupRepository) FindByID(ctx context.Context, db pgxtype.Querier, id uint) (Group, error) {
	var group Group
	if err := db.QueryRow(ctx, findGroupByIDQuery, id).Scan(&group.ID, &group.Name); err != nil {
		if err == pgx.ErrNoRows {
			return Group{}, ErrGroupNotFound
		}
		return Group{}, err
	}
	return group, nil
}

func (p *PgGroupRepository) FindByIDWithUsers(ctx context.Context, db pgxtype.Querier, id uint) (GroupResponse, error) {
	var group GroupResponse
	rows, err := db.Query(ctx, findGroupByIDWithUsersQuery, id)
	if err != nil {
		return GroupResponse{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var user UserResponse
		if err := rows.Scan(&group.ID, &group.Name, &user.ID, &user.Email); err != nil {
			return GroupResponse{}, err
		}
		group.Users = append(group.Users, user)
	}
	if rows.Err() != nil {
		return GroupResponse{}, rows.Err()
	}
	return group, nil
}

func (p *PgGroupRepository) FindByUserID(ctx context.Context, db pgxtype.Querier, userID uint) (Group, error) {
	var group Group
	if err := db.QueryRow(ctx, findGroupByUserIDQuery, userID).Scan(&group.ID, &group.Name); err != nil {
		if err == pgx.ErrNoRows {
			return Group{}, ErrGroupNotFound
		}
		return Group{}, err
	}
	return group, nil
}
