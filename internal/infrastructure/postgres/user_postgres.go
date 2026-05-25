package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/config"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type UserPostgres struct {
	db *sql.DB
}

func NewUserPostgresDB(db *sql.DB) (*UserPostgres, error) {
	pg := &UserPostgres{db}
	cfg := config.Get()

	if err := goose.Up(db, cfg.MigrationsDir); err != nil {
		return nil, err
	}
	return pg, nil
}

func (p *UserPostgres) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	query := `SELECT u."id", u."login", u."password" FROM public.user u WHERE u."login" = $1`
	var user user.User
	row := p.db.QueryRowContext(ctx, query, login)
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("error get user", zap.Error(err))
		}
		return nil, err
	}
	return &user, nil
}

func (p *UserPostgres) AddUser(ctx context.Context, login string, password string) (int64, error) {
	userQuery := `INSERT INTO public.user (login, password) VALUES ($1, $2) RETURNING id`
	walletQuery := `INSERT INTO public.wallet (user_id, balance, withdrawn) VALUES ($1, $2, $3)`
	startWalletBalance := 0
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	row := tx.QueryRowContext(ctx, userQuery, login, password)
	err = row.Scan(&id)
	if err != nil {
		logger.Log.Error("error add user", zap.Error(err))
		return 0, err
	}

	_, err = tx.ExecContext(ctx, walletQuery, id, startWalletBalance, startWalletBalance)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return id, nil
}
