package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/labstack/echo/v4"
)

type walletWithdrawalService interface {
	BalanceWithdrawal(ctx context.Context, userID int64, number string, amount int) error
}

type withdrawalRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

func HandleUserWithdrawal(service walletWithdrawalService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		contentType := c.Request().Header.Get("Content-Type")
		if contentType != "application/json" {
			return echo.NewHTTPError(http.StatusBadRequest, "wrong content-type")
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		defer c.Request().Body.Close()

		var request withdrawalRequest
		err = json.Unmarshal(body, &request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		err = service.BalanceWithdrawal(c.Request().Context(), userID, request.Order, request.Sum)

		if err != nil {
			if errors.Is(err, wallet.ErrNotEnoughBalance) {
				return echo.NewHTTPError(http.StatusPaymentRequired, err)
			}
			if errors.Is(err, wallet.ErrWrongNumber) {
				return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		return c.NoContent(http.StatusOK)
	}
}
