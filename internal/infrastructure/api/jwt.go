package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/config"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func MakeJWT(ID int64) (string, error) {
	cfg := config.Get()
	claims := user.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.TokenExp)),
		},
		UserID: ID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.SecretKey)
}

func getUserID(c echo.Context) (int64, error) {
	token, err := echo.ContextGet[*jwt.Token](c, "user")
	if err != nil {
		return 0, echo.ErrUnauthorized
	}
	claims, ok := token.Claims.(*user.UserClaims)
	if !ok {
		return 0, errors.New("failed to cast claims as UserClaims")
	}
	return claims.UserID, nil
}

func MakeJWTConfig() echojwt.Config {
	return echojwt.Config{
		SigningKey:    []byte(config.Get().SecretKey),
		SigningMethod: jwt.SigningMethodHS256.Name,
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, err)
		},
		TokenLookup: "header:Authorization",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(user.UserClaims)
		},
	}
}
