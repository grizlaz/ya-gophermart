package worker

import (
	"context"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
)

type orderService interface {
	SetOrdersStatus(ctx context.Context, newStatus order.Status, numbers ...string) error
}

type walletService interface {
	AddBalanceFromNumber(ctx context.Context, number string, amount int) error
}
