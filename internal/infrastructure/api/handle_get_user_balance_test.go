package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grizlaz/ya-gophermart/internal/domain/wallet"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetUserBalance(t *testing.T) {
	path := "/api/user/balance"
	userID := int64(1)
	balance := 100
	withdrawn := 99
	successResponse := &wallet.Wallet{
		UserID:    int(userID),
		Balance:   balance,
		Withdrawn: withdrawn,
	}

	token, err := MakeJWT(int64(userID))
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	e.Use(echojwt.WithConfig(MakeJWTConfig()))

	walletService := mock_api.NewMockwalletGetService(ctrl)
	walletService.EXPECT().GetUserBalance(context.Background(), userID).Return(successResponse, nil)
	walletService.EXPECT().GetUserBalance(context.Background(), userID).Return(nil, errors.New("something gone wrong"))
	e.GET(path, HandleGetUserBalance(walletService))

	type testCase struct {
		name           string
		expectedStatus int
		response       *walletGetBalance
	}

	testCases := []testCase{
		{
			name:           "success",
			expectedStatus: http.StatusOK,
			response: &walletGetBalance{
				Current:   balance,
				Withdrawn: withdrawn,
			},
		},
		{
			name:           "service return err",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Add("Authorization", token)
			c := e.NewContext(request, recorder)
			c.SetPath(path)

			e.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedStatus, recorder.Result().StatusCode, "wrong status code")
			if tc.response != nil {
				defer recorder.Result().Body.Close()
				responseBody, err := io.ReadAll(recorder.Result().Body)
				require.NoError(t, err)

				var response walletGetBalance
				err = json.Unmarshal(responseBody, &response)
				require.NoError(t, err)
				require.Equal(t, *tc.response, response)
			}

		})
	}
}
