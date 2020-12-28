package expenses

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
)

const (
	// PG Error codes
	uniqueViolation = "23505"
)

var (
	ErrEmailAlreadyExists = errors.New("user with such email already exists")
)

// UserRepository is an repository of users of the expenses system
type UserRepository interface {
	// Creates User in the storage
	Create(context.Context, CreateUserRequest) (User, error)
	// Find user by its ID
	FindById(ctx context.Context, id uint) (User, error)
}

const (
	createUserQuery   = "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING ID"
	findUserByIdQuery = "SELECT u.id, u.email, u.password FROM users as u WHERE u.id = $1"
)

// Implementation of UserRepository that works with postgresql
type PgUserRepository struct {
	db pgxtype.Querier
}

// NewPgRepository create new PgUserRepository
func NewPgRepository(db pgxtype.Querier) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Create(ctx context.Context, user CreateUserRequest) (User, error) {
	var id uint
	if err := r.db.QueryRow(ctx, createUserQuery, user.Email, user.Password).Scan(&id); err != nil {
		if pfError, ok := err.(*pgconn.PgError); ok && pfError.Code == uniqueViolation {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, err
	}
	return User{ID: id, Email: user.Email, Password: user.Password}, nil
}

// Looks up user in DB and returns it if it was found. Return zero value User if it was not found
func (r *PgUserRepository) FindById(ctx context.Context, id uint) (User, error) {
	var user User
	if err := r.db.QueryRow(ctx, findUserByIdQuery, id).Scan(&user.ID, &user.Email, &user.Password); err != nil {
		return User{}, err
	}
	return user, nil
}
