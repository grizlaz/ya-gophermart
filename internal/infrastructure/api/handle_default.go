package api

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func Handle() echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println(c.Request().URL.Path)
		return nil
	}
}
