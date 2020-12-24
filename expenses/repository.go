package expenses

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
)

type UserRepository interface {
	Create(user User) error
	FindById(id string) (User, error)
}

type PgUserRepository struct {
	db pgxtype.Querier
}

func NewRepositoryImpl(db pgxtype.Querier) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Create(user User) (User, error) {
	rows, err := r.db.Query(
		context.Background(),
		"INSERT INTO users (id, email, password) VALUES ($1, $2, $3)",
		user.ID,
		user.Email,
		user.Password,
	)
	if err != nil {
		return User{}, err
	}
	for rows.Next() {
		fmt.Printf("%++v\n", rows.RawValues())
	}
	return user, nil
}

func (r *PgUserRepository) FindById(id string) (User, error) {
	panic("implement me")
}
