package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserWithdrawal(t *testing.T) {
	path := "/api/user/balance/withdraw"
	userID := int64(1)
	number := "109"
	amount := 100
	token, err := MakeJWT(int64(userID))
	require.NoError(t, err)
	requestBody, err := json.Marshal(withdrawalRequest{
		Order: number,
		Sum:   amount,
	})
	require.NoError(t, err)
	t.Run("success empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := http.StatusOK

		walletWithdrawalService := mock_api.NewMockwalletWithdrawalService(ctrl)
		walletWithdrawalService.EXPECT().BalanceWithdrawal(context.Background(), userID, number, amount)

		e := echo.New()
		e.Use(echojwt.WithConfig(MakeJWTConfig()))
		e.POST(path, HandleUserWithdrawal(walletWithdrawalService))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(requestBody))
		request.Header.Add("Authorization", token)
		request.Header.Add("Content-Type", "application/json")

		c := e.NewContext(request, recorder)
		c.SetPath(path)

		e.ServeHTTP(recorder, request)

		assert.Equal(t, expectedStatus, recorder.Result().StatusCode, "wrong status code")
	})
}
