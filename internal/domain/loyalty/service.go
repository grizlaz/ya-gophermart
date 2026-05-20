package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/config"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	db        repository
	client    http.Client
	address   string
	ratelimit int
	orders    chan string
}

// не нравится что тут идет управление статусом заказа и кошельком пользователя, но пока не придумал как сделать красивее ы
func NewService(ctx context.Context, db repository) (*Service, error) {
	cfg := config.Get()
	service := &Service{
		db:        db,
		client:    http.Client{},
		address:   cfg.AccrualAddress,
		ratelimit: cfg.RateLimit,
		orders:    make(chan string, 100),
	}
	go service.checkOrders(ctx)
	return service, nil
}

func (l *Service) checkOrders(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var queue []string
	for {
		select {
		case newOrder := <-l.orders:
			logger.Log.Debug("add new order to queue", zap.String("number", newOrder))
			queue = append(queue, newOrder)
		case <-ticker.C:
			if len(queue) == 0 {
				continue
			}
			err := l.db.SetOrdersStatus(ctx, order.PROCESSING, queue...)
			if err != nil {
				logger.Log.Error("cannot update status for new orders", zap.Error(err))
				continue
			}
			g, ctx := errgroup.WithContext(ctx)
			g.SetLimit(l.ratelimit)
			for _, number := range queue {
				goNum := number
				g.Go(func() error {
					return l.validateResponse(ctx, goNum)
				})
			}
			queue = nil
			if err = g.Wait(); err != nil {
				logger.Log.Error("error while check order result", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (l *Service) validateResponse(ctx context.Context, number string) error {
	result, err := l.GetOrderStatus(ctx, number)
	if err != nil {
		return err
	}
	if slices.Contains(CheckLaterStatuses, result.Status) {
		//может висеть вечно
		l.AddOrderToQueue(number)
		return nil
	}
	userID, err := l.db.GetUserIDFromOrder(ctx, number)
	if err != nil {
		return err
	}
	err = l.db.SetOrdersStatus(ctx, order.Status(result.Status), number)
	if err != nil {
		return err
	}
	if result.Status == INVALID || result.Accrual == 0 {
		return nil
	}
	err = l.db.BalanceWithdrawal(ctx, userID, number, wallet.ADD, result.Accrual)
	if err != nil {
		return err
	}
	return nil
}

func (l *Service) AddOrderToQueue(number string) {
	l.orders <- number
}

func (l *Service) GetOrderStatus(ctx context.Context, number string) (*AccrualResponse, error) {
	if l.address == "" {
		return l.fakeRequest(number)
	}
	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", l.address, number), http.NoBody,
	)
	if err != nil {
		return nil, err
	}
	response, err := l.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, ErrWrongResponseStatus
	}
	if response.Header.Get("Content-Type") != "application/json" {
		return nil, ErrWrongContentType
	}

	var result AccrualResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (l *Service) fakeRequest(number string) (*AccrualResponse, error) {
	max := 1000
	sec1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	sec2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	sec3 := rand.New(rand.NewSource(time.Now().UnixNano()))
	//эмуляция недоступности сервиса
	if sec1.Int()%10 == 0 {
		return nil, ErrUnavailable
	}
	response := AccrualResponse{
		Order:   number,
		Accrual: sec3.Intn(max),
	}
	//эмуляция разной обработки заказов
	switch sec2.Int() % 4 {
	case 0:
		response.Status = REGISTERED
	case 1:
		response.Status = INVALID
	case 2:
		response.Status = PROCESSING
	case 3:
		response.Status = PROCESSED
	}
	return &response, nil
}
