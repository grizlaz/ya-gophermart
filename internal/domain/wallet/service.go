package wallet

import (
	"context"
	"errors"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
)

type Service struct {
	db repository
}

func NewService(db repository) (*Service, error) {
	return &Service{db}, nil
}

func (w *Service) GetUserBalance(ctx context.Context, userID int64) (*Wallet, error) {
	return w.db.GetUserBalance(ctx, userID)
}

func (w *Service) GetUserWithdrawals(ctx context.Context, userID int64) (*[]WalletHistory, error) {
	return w.db.GetUserWithdrawals(ctx, userID)
}

func (w *Service) BalanceWithdrawal(ctx context.Context, userID int64, number string, amount int) error {
	correctNumber := order.CheckLuhn(number)
	if !correctNumber {
		return ErrWrongNumber
	}
	userOrder, err := w.db.GetUserOrder(ctx, userID, number)
	if err != nil || userOrder == nil {
		return ErrWrongNumber
	}
	// todo что надо делать, если сумма для списания больше цены заказа?
	// if userOrder.Accrual < amount {}
	userWallet, err := w.db.GetUserBalance(ctx, userID)
	if err != nil {
		return errors.Join(err, ErrGetWallet)
	}

	if userWallet == nil {
		return ErrGetWallet
	}

	if userWallet.Balance < amount {
		return ErrNotEnoughBalance
	}

	return w.db.BalanceWithdrawal(ctx, userID, number, SUB, amount)
}

func (w *Service) AddBalanceFromNumber(ctx context.Context, number string, amount int) error {
	userID, err := w.db.GetUserIDFromOrder(ctx, number)
	if err != nil {
		return ErrGetUserID
	}
	return w.db.BalanceWithdrawal(ctx, userID, number, ADD, amount)
}
