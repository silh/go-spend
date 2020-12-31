package expenses

import (
	"errors"
	"time"
)

// Expense represents a single expense created by user as it is stored in DB.
type Expense struct {
	ID        uint
	UserID    uint
	Amount    float32
	Timestamp time.Time
}

// ExpenseResponse provides info about stored expense.
type ExpenseResponse struct {
	UserID    uint          `json:"userId"`
	Amount    float32       `json:"amount"`
	Timestamp time.Time     `json:"timestamp"`
	Shares    ExpenseShares `json:"shares"`
}

// NewExpense a context for creation of a new expense in DB.
type NewExpense struct {
	UserID uint
	Amount float32
}

// CreateExpenseContext contains all information for expense creation
type CreateExpenseContext struct {
	UserID  uint
	GroupID uint
	Amount  float32
	Shares  ExpenseShares
}

// CreateExpenseRequest represents an incoming JSON for creation of a new expense
type CreateExpenseRequest struct {
	Amount float32       `json:"amount"`
	Shares ExpenseShares `json:"shares"`
}

// ExpenseShares states how much each participant should have paid. Key - userID, value - Amount
type ExpenseShares map[uint]Percent

// Percent is uint between 0 and 100 for the particular context
type Percent uint

// ValidateCreateExpenseContext checks CreateExpenseContext to contain proper information. Doesn't check if specified
// participants are actually in the required group.
func ValidateCreateExpenseContext(req CreateExpenseContext) error {
	if req.Amount <= 0 {
		return errors.New("amount should be positive number")
	}
	if req.UserID == 0 {
		return errors.New("incorrect user")
	}
	if req.GroupID == 0 {
		return errors.New("incorrect group")
	}
	if len(req.Shares) == 0 {
		return errors.New("shares should contain at least one share")
	}
	totalPercent := Percent(0)
	for _, share := range req.Shares {
		totalPercent += share
	}
	if totalPercent != 100 {
		return errors.New("total percent for shares incorrect")
	}
	return nil
}
