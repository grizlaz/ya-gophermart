package worker

import (
	"context"

	"github.com/grizlaz/ya-gophermart/internal/domain/loyalty"
	orderLib "github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type Service struct {
	order  orderService
	wallet walletService
}

func NewService(ctx context.Context, orderService orderService, walletService walletService, processedOrders <-chan loyalty.AccrualResponse) (*Service, error) {
	service := &Service{
		order:  orderService,
		wallet: walletService,
	}
	go service.checkProcessedOrders(ctx, processedOrders)
	return service, nil
}

func (w *Service) checkProcessedOrders(ctx context.Context, processedOrders <-chan loyalty.AccrualResponse) {
	for {
		select {
		case order := <-processedOrders:
			err := w.order.SetOrdersStatus(ctx, orderLib.Status(order.Status), order.Order)
			if err != nil {
				logger.Log.Error("error process order. update status", zap.Error(err))
				break
			}
			if order.Status == loyalty.INVALID || order.Accrual == 0 {
				logger.Log.Debug("skip update wallet for order", zap.String("status", string(order.Status)), zap.String("number", order.Order), zap.Int("accrual", order.Accrual))
				break
			}
			err = w.wallet.AddBalanceFromNumber(ctx, order.Order, order.Accrual)
			if err != nil {
				logger.Log.Error("error process order. update wallet", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}
