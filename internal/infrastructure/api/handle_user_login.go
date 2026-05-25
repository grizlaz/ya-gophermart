package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/labstack/echo/v4"
)

type userAuthService interface {
	Auth(ctx context.Context, login string, password string) (int64, error)
}

func HandleUserLogin(userService userAuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer c.Request().Body.Close()
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		var request user.User
		err = json.Unmarshal(body, &request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		// shaPassword := sha256.Sum256([]byte(request.Password))
		// userID, err := userService.Auth(c.Request().Context(), request.Login, shaPassword)
		userID, err := userService.Auth(c.Request().Context(), request.Login, request.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		}

		token, err := MakeJWT(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		c.Response().Header().Set(user.AuthHeaderName, token)

		return c.NoContent(http.StatusOK)
	}
}
