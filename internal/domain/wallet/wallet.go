package wallet

import (
	"errors"
	"time"
)

type Wallet struct {
	UserID    int
	Balance   int
	Withdrawn int
}

type WalletOperation int

const (
	ADD WalletOperation = iota
	SUB
)

var (
	ErrNotEnoughBalance = errors.New("not enough balance")
	ErrWrongNumber      = errors.New("wrong number")
	ErrGetWallet        = errors.New("err get user wallet")
	ErrGetUserID        = errors.New("err get userID from number")
)

type WalletHistory struct {
	UserID    int64
	Number    string
	Operation WalletOperation
	Amount    int
	Date      time.Time
}
