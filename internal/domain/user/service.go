package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type Service struct {
	db repository
}

func NewService(db repository) (*Service, error) {
	return &Service{db}, nil
}

func (u *Service) Register(ctx context.Context, newUser User) (int64, error) {
	user, err := u.db.GetUserByLogin(ctx, newUser.Login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("err get user by login", zap.Error(err))
		return 0, err
	}
	if user != nil {
		return 0, ErrLoginAlreadyExists
	}
	return u.db.AddUser(ctx, newUser)
}

func (u *Service) Auth(ctx context.Context, login, password string) (int64, error) {
	user, err := u.db.GetUserByLogin(ctx, login)
	if err != nil {
		logger.Log.Error("err get user by login", zap.Error(err))
		return 0, ErrWrongUserData
	}
	if user.Password != password {
		return 0, ErrWrongUserData
	}
	return user.ID, nil
}
