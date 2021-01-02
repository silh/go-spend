package expenses

import (
	"context"
	"go-spend/db"
	"go-spend/log"
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
	balanceCache      BalanceCacheGetterSetter
	balanceRepository BalanceRepository
}

// NewDefaultBalanceService creates new instance of DefaultBalanceService
func NewDefaultBalanceService(
	db db.TxQuerier,
	// it could have been done using a decorator pattern as well, but this service already does nothing else
	balanceCache BalanceCacheGetterSetter,
	balanceRepository BalanceRepository,
) *DefaultBalanceService {
	return &DefaultBalanceService{db: db, balanceCache: balanceCache, balanceRepository: balanceRepository}
}

// Get Balance from a DB for provided user
func (d *DefaultBalanceService) Get(ctx context.Context, userID uint) (Balance, error) {
	cacheKey := BalanceCacheKey(userID)
	balance, err := d.balanceCache.Get(cacheKey)
	if err == nil {
		return balance, nil // return what found if there was no error
	}
	balance, err = d.balanceRepository.Get(ctx, d.db, userID)
	if err != nil {
		return nil, err
	}
	if err = d.balanceCache.Set(cacheKey, balance); err != nil {
		log.Warn("could not set key to cache - %s", err)
	}
	return balance, nil
}
