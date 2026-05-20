package user

import (
	"context"
)

type repository interface {
	GetUserByLogin(context.Context, string) (*User, error)
	AddUser(context.Context, User) (int64, error)
}
