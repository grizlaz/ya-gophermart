package loyalty

import (
	"context"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
)

type repository interface {
	SetOrdersStatus(ctx context.Context, newStatus order.Status, numbers ...string) error
	GetUserIDFromOrder(ctx context.Context, number string) (int64, error)
	BalanceWithdrawal(ctx context.Context, userID int64, number string, operation wallet.WalletOperation, amount int) error
}
