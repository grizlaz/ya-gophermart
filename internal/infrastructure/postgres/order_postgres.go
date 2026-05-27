package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type OrderPostgres struct {
	db *sql.DB
}

func NewOrderPostgresDB(db *sql.DB) (*OrderPostgres, error) {
	pg := &OrderPostgres{db}
	// миграции вызываются в user_postgres
	// cfg := config.Get()
	// if err := goose.Up(db, cfg.MigrationsDir); err != nil {
	// 	return nil, err
	// }
	return pg, nil
}

func (p *OrderPostgres) CreateOrder(ctx context.Context, orderToCreate order.Order) error {
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

func (p *OrderPostgres) SetOrdersStatus(ctx context.Context, newStatus order.Status, numbers ...string) error {
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

func (p *OrderPostgres) GetUserOrders(ctx context.Context, userID int64) ([]order.Order, error) {
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
	return orders, nil
}

func (p *OrderPostgres) GetUserOrder(ctx context.Context, userID int64, number string) (*order.Order, error) {
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

func (p *OrderPostgres) GetUserIDFromOrder(ctx context.Context, number string) (int64, error) {
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

func (p *OrderPostgres) BalanceWithdrawal(ctx context.Context, userID int64, number string, operation wallet.WalletOperation, amount int) error {
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
