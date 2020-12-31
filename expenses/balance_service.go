package expenses

import (
	"context"
	"go-spend/db"
)

// BalanceService provides means to fetch balance for current user. At the moment just delegates to repository but can
// be used, for example, when we need to add cache.
type BalanceService interface {
	// Get balance for User with userID
	Get(ctx context.Context, userID uint) (Balance, error)
}

// DefaultBalanceService is default implementation of BalanceService
type DefaultBalanceService struct {
	db                db.TxQuerier
	balanceRepository BalanceRepository
}

// NewDefaultBalanceService creates new instance of DefaultBalanceService
func NewDefaultBalanceService(db db.TxQuerier, balanceRepository BalanceRepository) *DefaultBalanceService {
	return &DefaultBalanceService{db: db, balanceRepository: balanceRepository}
}

// Get Balance from a DB for provided user
func (d *DefaultBalanceService) Get(ctx context.Context, userID uint) (Balance, error) {
	return d.balanceRepository.Get(ctx, d.db, userID)
}
