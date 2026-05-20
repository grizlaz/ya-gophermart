package app

import (
	"context"
	"net/http"

	"github.com/grizlaz/ya-gophermart/internal/domain/loyalty"
	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	"github.com/grizlaz/ya-gophermart/internal/infrastructure/api"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e              *echo.Echo
	userService    *user.Service
	orderService   *order.Service
	walletService  *wallet.Service
	loyaltyService *loyalty.Service
}

func NewServer(userService *user.Service, orderService *order.Service, walletService *wallet.Service, loyaltyService *loyalty.Service) *Server {
	s := &Server{
		userService:    userService,
		orderService:   orderService,
		walletService:  walletService,
		loyaltyService: loyaltyService,
	}
	s.setupRouter()
	return s
}

func (s *Server) setupRouter() {
	s.e = echo.New()
	s.e.HideBanner = true

	s.e.Pre(middleware.RemoveTrailingSlash())
	s.e.Use(middleware.Gzip())
	s.e.Use(middleware.Decompress())

	s.e.POST("/api/user/register", api.HandleUserRegister(s.userService))
	s.e.POST("/api/user/login", api.HandleUserLogin(s.userService))
	restricted := s.e.Group("/api/user")
	{
		restricted.Use(echojwt.WithConfig(api.MakeJWTConfig()))
		restricted.POST("/orders", api.HandleCreateUserOrder(s.orderService))
		restricted.GET("/orders", api.HandleGetUserOrders(s.orderService))
		restricted.GET("/balance", api.HandleGetUserBalance(s.walletService))
		restricted.POST("/balance/withdraw", api.HandleUserWithdrawal(s.walletService))
		restricted.GET("/withdrawals", api.HandleGetUserWithdrawals(s.walletService))
		restricted.Any("/*", wrongURL)
	}

	s.e.Any("/*", wrongURL)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.e.ServeHTTP(w, r)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.e.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func wrongURL(c echo.Context) error {
	return c.String(http.StatusBadRequest, "wrong url")
}
