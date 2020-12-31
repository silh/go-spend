package expenses

import "time"

// Expense represents a single expense created by user as it is stored in DB.
type Expense struct {
	ID        uint
	UserID    uint
	Amount    float32
	Timestamp time.Time
}

// CreateExpenseRequest a request for creation of a new expense in DB.
type CreateExpenseRequest struct {
	UserID uint
	Amount float32
}
