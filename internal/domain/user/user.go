package user

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       int64  `json:"-"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

var (
	ErrWrongUserData      = errors.New("wrong user/password")
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrUnauthorized       = errors.New("unauthorized")
)

var AuthHeaderName = "Authorization"
