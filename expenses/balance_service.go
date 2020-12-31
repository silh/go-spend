package expenses

import "context"

type BalanceService interface {
	Get(ctx context.Context, userID uint) (Balance, error)
}
