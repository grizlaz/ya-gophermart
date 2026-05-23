package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgresDB(db *sql.DB) (*Postgres, error) {
	pg := &Postgres{db}
	migrationsDir := "migrations"
	if err := goose.Up(db, migrationsDir); err != nil {
		return nil, err
	}
	return pg, nil
}

func (p *Postgres) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
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

func (p *Postgres) AddUser(ctx context.Context, login string, password [32]byte) (int64, error) {
	userQuery := `INSERT INTO public.user (login, password) VALUES ($1, $2) RETURNING id`
	walletQuery := `INSERT INTO public.wallet (user_id, balance, withdrawn) VALUES ($1, $2, $3)`
	startWalletBalance := 0
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	row := tx.QueryRowContext(ctx, userQuery, login, password[:])
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

func (p *Postgres) CreateOrder(ctx context.Context, orderToCreate order.Order) error {
	query := `INSERT INTO "order" ("number", user_id, status, accrual) VALUES ($1, $2, $3, $4)`
	_, err := p.db.ExecContext(ctx, query, orderToCreate.Number, orderToCreate.UserID, orderToCreate.Status, orderToCreate.Accrual)
	if err != nil {
		logger.Log.Error("error create order", zap.Error(err))
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = order.ErrConflict
			return err
		}
		return err
	}
	return nil
}

func (p *Postgres) SetOrdersStatus(ctx context.Context, newStatus order.Status, numbers ...string) error {
	argNumberDiff := 2
	var values []string
	args := make([]any, 0, len(numbers)+1)
	args = append(args, newStatus)
	for i, num := range numbers {
		params := fmt.Sprintf("(\"number\" = $%d)", i+argNumberDiff)
		values = append(values, params)
		args = append(args, num)
	}
	query := `UPDATE "order" o SET status = $1, updated_at = now() WHERE ` + strings.Join(values, " or ") + `;`
	logger.Log.Debug("SetOrdersStatus", zap.String("query", query), zap.String("status", string(newStatus)), zap.Strings("number", numbers))
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

func (p *Postgres) GetUserOrders(ctx context.Context, userID int64) (*[]order.Order, error) {
	query := `SELECT o."number", o.user_id, o.status, o.accrual, o.created_at, o.updated_at
			  FROM "order" o
			  WHERE o.user_id = $1
			  ORDER BY o.created_at DESC`
	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Log.Error("error get user orders", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	orders := make([]order.Order, 0)
	for rows.Next() {
		curOrder := order.Order{}
		err = rows.Scan(&curOrder.Number, &curOrder.UserID, &curOrder.Status, &curOrder.Accrual, &curOrder.CreatedAt, &curOrder.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, curOrder)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &orders, nil
}

func (p *Postgres) GetUserOrder(ctx context.Context, userID int64, number string) (*order.Order, error) {
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

func (p *Postgres) GetUserIDFromOrder(ctx context.Context, number string) (int64, error) {
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

func (p *Postgres) GetUserBalance(ctx context.Context, userID int64) (*wallet.Wallet, error) {
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

func (p *Postgres) GetUserWithdrawals(ctx context.Context, userID int64) (*[]wallet.WalletHistory, error) {
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

func (p *Postgres) BalanceWithdrawal(ctx context.Context, userID int64, number string, operation wallet.WalletOperation, amount int) error {
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
