package user

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
)

type Repository interface {
	Create(user User) error
	FindById(id string) (User, error)
}

type RepositoryImpl struct {
	db pgxtype.Querier
}

func NewRepositoryImpl(db pgxtype.Querier) *RepositoryImpl {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) Create(user User) (User, error) {
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

func (r *RepositoryImpl) FindById(id string) (User, error) {
	panic("implement me")
}
