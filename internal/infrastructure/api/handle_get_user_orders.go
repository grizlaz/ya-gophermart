package api

import (
	"context"
	"net/http"
	"time"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/labstack/echo/v4"
)

type orderGetService interface {
	GetUserOrders(ctx context.Context, userID int64) ([]order.Order, error)
}

type userOrder struct {
	Number    string       `json:"number"`
	Status    order.Status `json:"status"`
	Accrual   int          `json:"accrual,omitempty"`
	UpdatedAt string       `json:"uploaded_at"`
}

func HandleGetUserOrders(orderService orderGetService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		orders, err := orderService.GetUserOrders(c.Request().Context(), userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		if len(orders) == 0 {
			return c.NoContent(http.StatusNoContent)
		}

		response := make([]userOrder, 0, len(orders))
		for _, v := range orders {
			response = append(response, userOrder{
				Number:    v.Number,
				Status:    v.Status,
				Accrual:   v.Accrual,
				UpdatedAt: v.UpdatedAt.Format(time.RFC3339),
			})
		}

		return c.JSON(http.StatusOK, response)
	}
}
