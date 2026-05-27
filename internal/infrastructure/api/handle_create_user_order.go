package api

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/labstack/echo/v4"
)

type orderService interface {
	CreateOrder(ctx context.Context, userID int64, number string) error
}

func HandleCreateUserOrder(service orderService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := getUserID(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		contentType := c.Request().Header.Get("Content-Type")
		if contentType != "text/plain" {
			return echo.NewHTTPError(http.StatusBadRequest, "wrong content-type")
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		defer c.Request().Body.Close()

		number := string(body)
		if number == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "empty body")
		}

		err = service.CreateOrder(c.Request().Context(), userID, number)
		if err != nil {
			if errors.Is(err, order.ErrAlreadyAdded) {
				return c.NoContent(http.StatusOK)
			}
			if errors.Is(err, order.ErrConflict) {
				return echo.NewHTTPError(http.StatusConflict, err)
			}
			if errors.Is(err, order.ErrWrongNumber) {
				return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
			}
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		return c.NoContent(http.StatusAccepted)
	}
}
