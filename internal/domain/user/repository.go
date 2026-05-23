package user

import (
	"context"
)

type repository interface {
	GetUserByLogin(context.Context, string) (*User, error)
	AddUser(context.Context, string, [32]byte) (int64, error)
}
