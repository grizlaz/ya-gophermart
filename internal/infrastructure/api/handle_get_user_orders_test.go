package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetUserOrders(t *testing.T) {
	path := "/api/user/orders"
	userID := int64(1)
	token, err := MakeJWT(int64(userID))
	require.NoError(t, err)
	t.Run("success empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := http.StatusNoContent

		orderGetService := mock_api.NewMockorderGetService(ctrl)
		orderGetService.EXPECT().GetUserOrders(context.Background(), userID).Return(make([]order.Order, 0), nil)

		e := echo.New()
		e.Use(echojwt.WithConfig(MakeJWTConfig()))
		e.GET(path, HandleGetUserOrders(orderGetService))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Add("Authorization", token)

		c := e.NewContext(request, recorder)
		c.SetPath(path)

		e.ServeHTTP(recorder, request)

		assert.Equal(t, expectedStatus, recorder.Result().StatusCode, "wrong status code")
		defer recorder.Result().Body.Close()
		responseBody, err := io.ReadAll(recorder.Result().Body)
		require.NoError(t, err)

		require.Empty(t, responseBody)
	})
	t.Run("success orders", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := http.StatusOK

		now := time.Now()
		orders := []order.Order{
			{
				Number:    "109",
				UserID:    userID,
				Status:    order.NEW,
				Accrual:   0,
				UpdatedAt: now,
			},
		}
		response := []userOrder{
			{
				Number:    "109",
				Status:    order.NEW,
				Accrual:   0,
				UpdatedAt: now.Format(time.RFC3339),
			},
		}

		orderGetService := mock_api.NewMockorderGetService(ctrl)
		orderGetService.EXPECT().GetUserOrders(context.Background(), userID).Return(orders, nil)

		e := echo.New()
		e.Use(echojwt.WithConfig(MakeJWTConfig()))
		e.GET(path, HandleGetUserOrders(orderGetService))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Add("Authorization", token)

		c := e.NewContext(request, recorder)
		c.SetPath(path)

		e.ServeHTTP(recorder, request)

		assert.Equal(t, expectedStatus, recorder.Result().StatusCode, "wrong status code")
		defer recorder.Result().Body.Close()
		responseBody, err := io.ReadAll(recorder.Result().Body)
		require.NoError(t, err)

		var result []userOrder
		err = json.Unmarshal(responseBody, &result)
		require.NoError(t, err)
		require.Equal(t, response, result)
	})
}
