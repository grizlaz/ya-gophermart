package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/grizlaz/ya-gophermart/internal/app"
	"github.com/grizlaz/ya-gophermart/internal/domain/loyalty"
	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/domain/worker"
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

	userService, loyaltyService, orderService, walletService := initServices(ctx, storage)
	srv := app.NewServer(userService, orderService, walletService, loyaltyService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(config.ServerAddress, srv); !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("error running server", zap.Error(err))
		}
	}()

	logger.Log.Info("server started")
	<-quit

	shutdowCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdowCtx); err != nil {
		logger.Log.Fatal("error closing server", zap.Error(err))
	}

	logger.Log.Info("server stopped")
}

func initServices(ctx context.Context, storage *repository.Postgres) (*user.Service, *loyalty.Service, *order.Service, *wallet.Service) {
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
	_, err = worker.NewService(ctx, orderService, walletService, loyaltyService.OrderProcessed)
	if err != nil {
		logger.Log.Fatal("error init worker service", zap.Error(err))
	}
	return userService, loyaltyService, orderService, walletService
}
