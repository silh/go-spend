package expenses

import "context"

// Service for expenses
type Service interface {
	Create(ctx context.Context, newExpense CreateExpenseContext) (ExpenseResponse, error)
}

type DefaultService struct {
}

func (d *DefaultService) Create(ctx context.Context, newExpense CreateExpenseContext) (ExpenseResponse, error) {
	return ExpenseResponse{}, nil
}
