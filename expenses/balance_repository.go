package expenses

import (
	"context"
	"go-spend/db"
)

// BalanceRepository provides access to balance calculations in the storage
type BalanceRepository interface {
	Get(ctx context.Context, db db.TxQuerier, userID uint) (Balance, error)
}

const (
	getBalanceQuery = `WITH other_users as (
    SELECT ug_full.user_id
    FROM users_groups as ug_full
             LEFT JOIN users_groups as ug ON ug.group_id = ug_full.group_id
             JOIN users as u ON u.id = ug_full.user_id
    WHERE ug.user_id = $1
      AND ug_full.user_id <> $1
),
     who_i_owe as (
         SELECT sum((es.percent * e.amount) / 100) as received, e.user_id
         FROM expenses_shares as es
                  JOIN expenses as e ON es.expense_id = e.id
                  JOIN users as u ON u.id = es.user_id
                  JOIN other_users ON other_users.user_id = e.user_id
         WHERE es.user_id = $1
         GROUP BY e.user_id
     ),
     who_owes_me as (
         SELECT sum((es.percent * e.amount) / 100) as gave, es.user_id
         FROM expenses_shares as es
                  JOIN expenses as e ON es.expense_id = e.id
                  JOIN users as u ON u.id = es.user_id
                  JOIN other_users ON other_users.user_id = es.user_id
         WHERE e.user_id = $1
           and es.user_id = other_users.user_id
         GROUP BY es.user_id
     )
SELECT CASE
           WHEN who_owes_me.user_id IS NULL THEN who_i_owe.user_id
           ELSE who_owes_me.user_id END as user_id,
       CASE
           WHEN who_owes_me.gave is NULL THEN -who_i_owe.received
           WHEN who_i_owe.received IS NULL THEN who_owes_me.gave
           ELSE who_owes_me.gave -who_i_owe.received END as balance
FROM who_owes_me
         FULL JOIN who_i_owe ON who_owes_me.user_id = who_i_owe.user_id`
)

// PgBalanceRepository is BalanceRepository that works with PostgresDB
type PgBalanceRepository struct {
}

// NewPgBalanceRepository creates new PgBalanceRepository
func NewPgBalanceRepository() *PgBalanceRepository {
	return &PgBalanceRepository{}
}

func (*PgBalanceRepository) Get(ctx context.Context, db db.TxQuerier, userID uint) (Balance, error) {
	rows, err := db.Query(ctx, getBalanceQuery, userID)
	if err != nil {
		return Balance{}, err
	}
	totalBalance := make(Balance)
	for rows.Next() {
		var oneBalance balanceLine
		if err := rows.Scan(&oneBalance.UserID, &oneBalance.Balance); err != nil {
			return Balance{}, err
		}
		totalBalance[oneBalance.UserID] = oneBalance.Balance
	}
	return totalBalance, nil
}

type balanceLine struct {
	UserID  uint
	Balance float32
}
