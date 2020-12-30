package expenses

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
)

const (
	// PG Error codes
	uniqueViolation = "23505"
)

// UserRepository is an repository of users of the expenses system
type UserRepository interface {
	// Creates User in the storage
	Create(context.Context, pgxtype.Querier, CreateUserRequest) (User, error)
	// Find User by its ID
	FindById(ctx context.Context, db pgxtype.Querier, id uint) (User, error)
	// Find User by email
	FindByEmail(ctx context.Context, db pgxtype.Querier, email Email) (User, error)
}

const (
	createUserQuery   = "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING ID"
	findUserByIdQuery = "SELECT u.id, u.email, u.password, COALESCE(ug.group_id, 0) " +
		"FROM users as u " +
		"LEFT JOIN users_groups as ug ON u.id = ug.user_id " +
		"WHERE u.id = $1"
	findUserByEmailQuery = "SELECT u.id, u.email, u.password, COALESCE(ug.group_id, 0) " +
		"FROM users as u " +
		"LEFT JOIN users_groups as ug ON u.id = ug.user_id " +
		"WHERE u.email = $1"
)

var (
	ErrEmailAlreadyExists = errors.New("user with such email already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// Implementation of UserRepository that works with postgresql
type PgUserRepository struct {
}

// NewPgUserRepository create new PgUserRepository
func NewPgUserRepository() *PgUserRepository {
	return &PgUserRepository{}
}

func (r *PgUserRepository) Create(ctx context.Context, db pgxtype.Querier, user CreateUserRequest) (User, error) {
	var id uint
	if err := db.QueryRow(ctx, createUserQuery, user.Email, user.Password).Scan(&id); err != nil {
		if pfError, ok := err.(*pgconn.PgError); ok && pfError.Code == uniqueViolation {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, err
	}
	return User{ID: id, Email: user.Email, Password: user.Password}, nil
}

// Looks up user in DB and returns it if it was found. Returns ErrUserNotFound if wasn't found.
func (r *PgUserRepository) FindById(ctx context.Context, db pgxtype.Querier, id uint) (User, error) {
	var user User
	row := db.QueryRow(ctx, findUserByIdQuery, id)
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.GroupID); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

// Looks up user in DB by email and returns it if it was found. Returns ErrUserNotFound if wasn't found.
func (r *PgUserRepository) FindByEmail(ctx context.Context, db pgxtype.Querier, email Email) (User, error) {
	var user User
	row := db.QueryRow(ctx, findUserByEmailQuery, email)
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.GroupID); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}
