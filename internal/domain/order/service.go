package order

import (
	"context"
)

type loyaltyService interface {
	// GetOrderStatus(ctx context.Context, number string) (*loyalty.AccrualResponse, error)
	AddOrderToQueue(number string)
}

type Service struct {
	db      repository
	loyalty loyaltyService
}

func NewService(ctx context.Context, db repository, loaloyaltyService loyaltyService) (*Service, error) {
	service := &Service{
		db:      db,
		loyalty: loaloyaltyService,
	}
	return service, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID int64, number string) error {
	if len(number) == 0 {
		return ErrEmptyOrderNumber
	}
	ok := CheckLuhn(number)
	if !ok {
		return ErrWrongNumber
	}
	newOrder := Order{
		Number:  number,
		UserID:  userID,
		Status:  NEW,
		Accrual: 0,
	}
	oldOrder, err := s.db.GetUserOrder(ctx, userID, number)
	if err != nil {
		return err
	}
	if oldOrder != nil {
		return ErrAlreadyAdded
	}
	err = s.db.CreateOrder(ctx, newOrder)
	if err != nil {
		return err
	}
	s.loyalty.AddOrderToQueue(number)
	return nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int64) (*[]Order, error) {
	return s.db.GetUserOrders(ctx, userID)
}

func CheckLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}
	sum := 0
	nDigits := len(number)
	parity := nDigits % 2
	for i := range nDigits {
		digit := int(number[i]) - '0'
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return (sum % 10) == 0
}
