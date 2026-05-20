package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/labstack/echo/v4"
)

type userRegisterService interface {
	Register(ctx context.Context, newUser user.User) (int64, error)
}

// type userRegisterRequest struct {
// 	Login    string `json:"login"`
// 	Password string `json:"password"`
// }

func HandleUserRegister(userService userRegisterService) echo.HandlerFunc {
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

		userID, err := userService.Register(c.Request().Context(), request)
		if err != nil {
			if errors.Is(err, user.ErrLoginAlreadyExists) {
				return echo.NewHTTPError(http.StatusConflict, user.ErrLoginAlreadyExists)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		token, err := makeJWT(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		c.Response().Header().Set(user.AuthHeaderName, token)

		return c.NoContent(http.StatusOK)
	}
}
