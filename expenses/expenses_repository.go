package expenses

import (
	"context"
	"github.com/jackc/pgtype/pgxtype"
)

// Repository for User and Group expenses
type Repository interface {
	// Create stores new Expense
	Create(ctx context.Context, db pgxtype.Querier, req NewExpense) (Expense, error)
}

const (
	createExpenseQuery = "INSERT INTO expenses (user_id, amount) VALUES ($1, $2) RETURNING id, timestamp"
)

type PgRepository struct {
}

func (p *PgRepository) Create(ctx context.Context, db pgxtype.Querier, req NewExpense) (Expense, error) {
	result := Expense{
		UserID: req.UserID,
		Amount: req.Amount,
	}
	if err := db.QueryRow(ctx, createExpenseQuery, req.UserID, req.Amount).Scan(&result.ID, &result.Timestamp); err != nil {
		return Expense{}, err
	}
	return result, nil
}
