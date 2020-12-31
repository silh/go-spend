package expenses

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"go-spend/db"
	"go-spend/log"
)

// Service for storing and retrieving expenses.
type Service interface {
	Create(ctx context.Context, newExpense CreateExpenseContext) (ExpenseResponse, error)
}

var (
	ErrCreatorNotInGroup     = errors.New("expense creator not in a group")
	ErrParticipantNotInGroup = errors.New("user in shares is not in a group")
)

// DefaultService is a default implementation of Service
type DefaultService struct {
	db                 db.TxQuerier
	groupRepository    GroupRepository
	expensesRepository Repository
}

func NewDefaultService(
	db db.TxQuerier,
	groupRepository GroupRepository,
	expensesRepository Repository,
) *DefaultService {
	return &DefaultService{db: db, groupRepository: groupRepository, expensesRepository: expensesRepository}
}

func (d *DefaultService) Create(ctx context.Context, createExpenseContext CreateExpenseContext) (ExpenseResponse, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return ExpenseResponse{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Error("failed to rollback transaction - %s", err.Error())
		}
	}()
	newExpense := NewExpense{
		UserID: createExpenseContext.UserID,
		Amount: createExpenseContext.Amount,
	}
	createdExpense, err := d.expensesRepository.Create(ctx, tx, newExpense)
	if err != nil {
		return ExpenseResponse{}, err
	}
	err = d.validateUsersInContext(ctx, createExpenseContext, tx)
	if err != nil {
		return ExpenseResponse{}, err
	}
	createExpenseShares := CreateExpenseShares{
		ExpenseID: createdExpense.ID,
		Shares:    createExpenseContext.Shares,
	}
	if err = d.expensesRepository.CreateShares(ctx, tx, createExpenseShares); err != nil {
		return ExpenseResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return ExpenseResponse{}, err
	}
	return ExpenseResponse{
		UserID:    createExpenseContext.UserID,
		Amount:    createExpenseContext.Amount,
		Timestamp: createdExpense.Timestamp,
		Shares:    createExpenseContext.Shares,
	}, nil
}

func (d *DefaultService) validateUsersInContext(
	ctx context.Context,
	createExpenseContext CreateExpenseContext,
	tx pgx.Tx,
) error {
	group, err := d.groupRepository.FindByIDWithUsers(ctx, tx, createExpenseContext.GroupID)
	if err != nil {
		return err
	}
	allUserIDs := map[uint]struct{}{}
	for _, user := range group.Users {
		allUserIDs[user.ID] = struct{}{}
	}
	// Creator in the group
	if _, ok := allUserIDs[createExpenseContext.UserID]; !ok {
		return ErrCreatorNotInGroup
	}
	// Mentioned in shares are in the group
	for userID := range createExpenseContext.Shares {
		if _, ok := allUserIDs[userID]; !ok {
			return ErrParticipantNotInGroup
		}
	}
	return nil
}
