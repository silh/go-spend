package expenses

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype/pgxtype"
	pg "go-spend/db"
	"strings"
)

// Repository for User and Group expenses
type Repository interface {
	// Create stores new Expense
	Create(ctx context.Context, db pgxtype.Querier, req NewExpense) (Expense, error)
	// CreateShares stores shares of already existing Expense
	CreateShares(ctx context.Context, db pgxtype.Querier, req CreateExpenseShares) error
}

const (
	createExpenseQuery        = "INSERT INTO expenses (user_id, amount) VALUES ($1, $2) RETURNING id, timestamp"
	createExpensesSharesQuery = "INSERT INTO expenses_shares (expense_id, user_id, percent) VALUES "
)

var (
	ErrNotAllInserted           = errors.New("not all inserted")
	ErrUserOrExpenseDoesntExist = errors.New("user or expense doesn't exist")
)

type PgRepository struct {
}

func NewPgRepository() *PgRepository {
	return &PgRepository{}
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

func (p *PgRepository) CreateShares(ctx context.Context, db pgxtype.Querier, req CreateExpenseShares) error {
	query := createExpensesSharesQuery
	counter := 1
	var params []interface{}
	for user, percent := range req.Shares {
		query += fmt.Sprintf("($%d, $%d, $%d) ,", counter, counter+1, counter+2)
		counter += 3
		params = append(params, req.ExpenseID, user, percent)
	}
	query = strings.TrimSuffix(query, ",")
	commandTag, err := db.Exec(ctx, query, params...)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == pg.ForeignKeyViolation {
			return ErrUserOrExpenseDoesntExist
		}
		return err
	}
	if commandTag.RowsAffected() != int64(len(req.Shares)) {
		return ErrNotAllInserted
	}
	return nil
}
