package user

import (
	"bytes"
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

func (u *Service) Register(ctx context.Context, login string, password [32]byte) (int64, error) {
	user, err := u.db.GetUserByLogin(ctx, login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("err get user by login", zap.Error(err))
		return 0, err
	}
	if user != nil {
		return 0, ErrLoginAlreadyExists
	}
	return u.db.AddUser(ctx, login, password)
}

func (u *Service) Auth(ctx context.Context, login string, password [32]byte) (int64, error) {
	user, err := u.db.GetUserByLogin(ctx, login)
	if err != nil {
		logger.Log.Error("err get user by login", zap.Error(err))
		return 0, ErrWrongUserData
	}
	if !bytes.Equal([]byte(user.Password), password[:]) {
		return 0, ErrWrongUserData
	}
	return user.ID, nil
}
