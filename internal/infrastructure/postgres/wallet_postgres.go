package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type WalletPostgres struct {
	db *sql.DB
}

func NewWalletPostgresDB(db *sql.DB) (*WalletPostgres, error) {
	pg := &WalletPostgres{db}
	// миграции вызываются в user_postgres
	// cfg := config.Get()
	// if err := goose.Up(db, cfg.MigrationsDir); err != nil {
	// 	return nil, err
	// }
	return pg, nil
}

func (p *WalletPostgres) GetUserIDFromOrder(ctx context.Context, number string) (int64, error) {
	query := `SELECT o.user_id
			  FROM "order" o
			  WHERE o."number" = $1`
	var userID int64
	row := p.db.QueryRowContext(ctx, query, number)
	err := row.Scan(&userID)
	if err != nil {
		logger.Log.Error("error get user userID from order", zap.Error(err))
		return 0, err
	}
	return userID, nil
}

func (p *WalletPostgres) GetUserOrder(ctx context.Context, userID int64, number string) (*order.Order, error) {
	query := `SELECT o."number", o.user_id, o.status, o.accrual, o.created_at, o.updated_at
			  FROM "order" o
			  WHERE o.user_id = $1 and o."number" = $2`
	var order order.Order
	row := p.db.QueryRowContext(ctx, query, userID, number)
	err := row.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Log.Error("error get user order", zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (p *WalletPostgres) GetUserBalance(ctx context.Context, userID int64) (*wallet.Wallet, error) {
	query := `SELECT w.user_id, w.balance, w.withdrawn 
			  FROM "wallet" w 
			  WHERE w.user_id = $1`
	var wallet wallet.Wallet
	row := p.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(&wallet.UserID, &wallet.Balance, &wallet.Withdrawn)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (p *WalletPostgres) GetUserWithdrawals(ctx context.Context, userID int64) (*[]wallet.WalletHistory, error) {
	query := `SELECT h.user_id, h."number", h.operation, h.amount, h.date
			  FROM "wallet_history" h
			  WHERE h.user_id = $1 and h.operation = $2`
	var walletHistory wallet.WalletHistory
	rows, err := p.db.QueryContext(ctx, query, userID, wallet.SUB)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	histories := make([]wallet.WalletHistory, 0)
	for rows.Next() {
		curHistory := walletHistory
		err = rows.Scan(&curHistory.UserID, &curHistory.Number, &curHistory.Operation, &curHistory.Amount, &curHistory.Date)
		if err != nil {
			return nil, err
		}
		histories = append(histories, curHistory)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &histories, nil
}

func (p *WalletPostgres) BalanceWithdrawal(ctx context.Context, userID int64, number string, operation wallet.WalletOperation, amount int) error {
	changeBalanceQuery := `UPDATE "wallet"
						   SET balance = balance - $2, 
						   	   withdrawn = withdrawn + $2
						   WHERE user_id = $1`
	if operation == wallet.ADD {
		changeBalanceQuery = `UPDATE "wallet"
							   SET balance = balance + $2
							   WHERE user_id = $1`
	}
	addHistoryQuery := `INSERT INTO "wallet_history" (user_id, "number", operation, amount)
						VALUES ($1, $2, $3, $4)`

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("error init tx", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, changeBalanceQuery, userID, amount)
	if err != nil {
		logger.Log.Error("error change balance", zap.Error(err))
		return err
	}

	_, err = tx.ExecContext(ctx, addHistoryQuery, userID, number, operation, amount)
	if err != nil {
		logger.Log.Error("error add withdrawal history", zap.Error(err))
		return err
	}

	return tx.Commit()
}
