package expenses

import (
	"context"
	"github.com/jackc/pgtype/pgxtype"
)

// UserRepository is an repository of users of the expenses system
type UserRepository interface {
	// Creates User in the storage
	Create(context.Context, User) (uint, error)
	// Find user by its ID
	FindById(ctx context.Context, id uint) (User, error)
}

const (
	createUserQuery = "INSERT INTO users (id, email, password) VALUES ($1, $2, $3) RETURNING ID"
)

// Implementation of UserRepository that works with postgresql
type PgUserRepository struct {
	db pgxtype.Querier
}

// NewPgRepository create new PgUserRepository. According to docs all pgxtype.Querier Query calls return rows so the
// returned exception is ignored as it is provided by rows.Error()
func NewPgRepository(db pgxtype.Querier) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Create(ctx context.Context, user User) (uint, error) {
	var id uint
	rows, _ := r.db.Query(
		ctx,
		createUserQuery,
		user.ID,
		user.Email,
		user.Password,
	)
	defer rows.Close()
	if rows.Err() != nil {
		return id, rows.Err()
	}
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return id, err
		}
	}

	return id, nil
}

// Looks up user in DB and returns it if it was found. Return zero value User if it was not found
func (r *PgUserRepository) FindById(ctx context.Context, id uint) (User, error) {
	var user User
	rows, _ := r.db.Query(
		ctx,
		"SELECT u.id, u.email, u.password FROM users as u WHERE u.id = $1",
		id,
	)
	defer rows.Close()
	if rows.Err() != nil {
		return user, rows.Err()
	}
	if !rows.Next() {
		return user, nil
	}
	return user, rows.Scan(&user.ID, &user.Email, &user.Password)
}
