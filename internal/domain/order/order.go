package order

import (
	"errors"
	"time"
)

type Status string

const (
	// REGISTERED Status = "REGISTERED"
	// INVALID    Status = "INVALID"
	// PROCESSING Status = "PROCESSING"
	// PROCESSED  Status = "PROCESSED"
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
)

var (
	ErrEmptyOrderNumber = errors.New("empty order number")
	ErrConflict         = errors.New("order number already exists")
	ErrAlreadyAdded     = errors.New("user already create order")
	ErrWrongNumber      = errors.New("wrong order number")
)

type Order struct {
	Number    string
	UserID    int64
	Status    Status
	Accrual   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
