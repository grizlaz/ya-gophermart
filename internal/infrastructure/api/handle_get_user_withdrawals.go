package api

import (
	"context"
	"net/http"
	"time"

	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/labstack/echo/v4"
)

type walletWithdrawalsService interface {
	GetUserWithdrawals(context.Context, int64) (*[]wallet.WalletHistory, error)
}

type walletWithdrawal struct {
	Order       string `json:"order"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}

func HandleGetUserWithdrawals(service walletWithdrawalsService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		withdrawals, err := service.GetUserWithdrawals(c.Request().Context(), userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		if len(*withdrawals) == 0 {
			return c.NoContent(http.StatusNoContent)
		}

		response := make([]walletWithdrawal, 0, len(*withdrawals))
		for _, v := range *withdrawals {
			response = append(response, walletWithdrawal{
				Order:       v.Number,
				Sum:         v.Amount,
				ProcessedAt: v.Date.Format(time.RFC3339),
			})
		}

		return c.JSON(http.StatusOK, response)
	}
}
