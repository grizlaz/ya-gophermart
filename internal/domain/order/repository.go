package order

import (
	"context"
)

type repository interface {
	CreateOrder(context.Context, Order) error
	GetUserOrders(ctx context.Context, userID int64) ([]Order, error)
	GetUserOrder(ctx context.Context, userID int64, number string) (*Order, error)
	SetOrdersStatus(ctx context.Context, newStatus Status, numbers ...string) error
}
