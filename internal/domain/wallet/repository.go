package wallet

import (
	"context"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
)

type repository interface {
	GetUserBalance(context.Context, int64) (*Wallet, error)
	GetUserWithdrawals(context.Context, int64) (*[]WalletHistory, error)
	GetUserOrder(ctx context.Context, userID int64, number string) (*order.Order, error)
	BalanceWithdrawal(ctx context.Context, userID int64, number string, operation WalletOperation, amount int) error
}
