package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/grizlaz/ya-gophermart/internal/app"
	"github.com/grizlaz/ya-gophermart/internal/domain/loyalty"
	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/config"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	repository "github.com/grizlaz/ya-gophermart/internal/infrastructure/postgres"
	"go.uber.org/zap"
)

func main() {
	config := config.Get()
	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	db, err := sql.Open("pgx", config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("error init db", zap.Error(err))
	}
	defer db.Close()
	ctx := context.Background()
	storage, err := repository.NewPostgresDB(db)
	if err != nil {
		logger.Log.Fatal("error init pg repository", zap.Error(err))
	}
	userService, err := user.NewService(storage)
	if err != nil {
		logger.Log.Fatal("error init user service", zap.Error(err))
	}
	loyaltyService, err := loyalty.NewService(ctx, storage)
	if err != nil {
		logger.Log.Fatal("error init loyalty service", zap.Error(err))
	}
	orderService, err := order.NewService(ctx, storage, loyaltyService)
	if err != nil {
		logger.Log.Fatal("error init order service", zap.Error(err))
	}
	walletService, err := wallet.NewService(storage)
	if err != nil {
		logger.Log.Fatal("error init wallet service", zap.Error(err))
	}
	srv := app.NewServer(userService, orderService, walletService, loyaltyService)
	if err := http.ListenAndServe(config.ServerAddress, srv); !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Fatal("error running server", zap.Error(err))
	}
}
