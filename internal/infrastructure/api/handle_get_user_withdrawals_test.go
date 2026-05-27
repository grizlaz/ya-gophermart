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
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetUserWithdrawals(t *testing.T) {
	path := "/api/user/withdrawals"
	userID := int64(1)
	token, err := MakeJWT(int64(userID))
	require.NoError(t, err)
	t.Run("success empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := http.StatusNoContent

		walletWithdrawalsService := mock_api.NewMockwalletWithdrawalsService(ctrl)
		walletWithdrawalsService.EXPECT().GetUserWithdrawals(context.Background(), userID).Return(&[]wallet.WalletHistory{}, nil)

		e := echo.New()
		e.Use(echojwt.WithConfig(MakeJWTConfig()))
		e.GET(path, HandleGetUserWithdrawals(walletWithdrawalsService))

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
		history := []wallet.WalletHistory{
			{
				UserID:    userID,
				Number:    "109",
				Operation: wallet.ADD,
				Amount:    109,
				Date:      now,
			},
		}
		response := []walletWithdrawal{
			{
				Order:       "109",
				Sum:         109,
				ProcessedAt: now.Format(time.RFC3339),
			},
		}

		walletWithdrawalsService := mock_api.NewMockwalletWithdrawalsService(ctrl)
		walletWithdrawalsService.EXPECT().GetUserWithdrawals(context.Background(), userID).Return(&history, nil)

		e := echo.New()
		e.Use(echojwt.WithConfig(MakeJWTConfig()))
		e.GET(path, HandleGetUserWithdrawals(walletWithdrawalsService))

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

		var result []walletWithdrawal
		err = json.Unmarshal(responseBody, &result)
		require.NoError(t, err)
		require.Equal(t, response, result)
	})
}
