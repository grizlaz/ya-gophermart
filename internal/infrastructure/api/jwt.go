package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/config"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/logger"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func makeJWT(ID int64) (string, error) {
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
	tokenString := c.Request().Header.Get(user.AuthHeaderName)
	if len(tokenString) == 0 {
		return 0, user.ErrUnauthorized
	}
	cfg := config.Get()

	claims := &user.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return cfg.SecretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims.UserID == 0 {
		return 0, user.ErrUnauthorized
	}

	if !token.Valid {
		logger.Log.Info("invalid token")
		return 0, err
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
