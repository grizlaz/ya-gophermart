package api

import (
	"context"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/labstack/echo/v4"
)

type walletGetService interface {
	GetUserBalance(context.Context, int64) (*wallet.Wallet, error)
}

type walletGetBalance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

func HandleGetUserBalance(service walletGetService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		wallet, err := service.GetUserBalance(c.Request().Context(), userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		response := walletGetBalance{
			Current:   wallet.Balance,
			Withdrawn: wallet.Withdrawn,
		}
		return c.JSON(http.StatusOK, response)
	}
}
