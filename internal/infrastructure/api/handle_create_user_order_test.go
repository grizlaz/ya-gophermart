package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateUserOrder(t *testing.T) {
	path := "/api/user/orders"
	userID := int64(1)
	number := "109"

	token, err := MakeJWT(int64(userID))
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	e := echo.New()
	e.Use(echojwt.WithConfig(MakeJWTConfig()))

	orderService := mock_api.NewMockorderService(ctrl)
	orderService.EXPECT().CreateOrder(context.Background(), userID, number).Return(nil)
	orderService.EXPECT().CreateOrder(context.Background(), userID, "added").Return(order.ErrAlreadyAdded)
	orderService.EXPECT().CreateOrder(context.Background(), userID, "conflict").Return(order.ErrConflict)
	orderService.EXPECT().CreateOrder(context.Background(), userID, "1").Return(order.ErrWrongNumber)
	e.POST(path, HandleCreateUserOrder(orderService))

	type testCase struct {
		name           string
		expectedStatus int
		body           *bytes.Reader
		token          string
		contentType    string
		response       string
	}
	testCases := []testCase{
		{
			name:           "success",
			expectedStatus: http.StatusAccepted,
			body:           bytes.NewReader([]byte(number)),
			token:          token,
			contentType:    "text/plain",
		},
		{
			name:           "unauthorized",
			expectedStatus: http.StatusUnauthorized,
			body:           bytes.NewReader([]byte(number)),
			token:          "",
			contentType:    "text/plain",
		},
		{
			name:           "wrong content type",
			expectedStatus: http.StatusBadRequest,
			body:           bytes.NewReader([]byte(number)),
			token:          token,
			contentType:    "application/json",
		},
		{
			name:           "empty body",
			expectedStatus: http.StatusBadRequest,
			body:           bytes.NewReader([]byte("")),
			token:          token,
			contentType:    "text/plain",
			response:       "empty body",
		},
		{
			name:           "already added",
			expectedStatus: http.StatusOK,
			body:           bytes.NewReader([]byte("added")),
			token:          token,
			contentType:    "text/plain",
		},
		{
			name:           "conflict",
			expectedStatus: http.StatusConflict,
			body:           bytes.NewReader([]byte("conflict")),
			token:          token,
			contentType:    "text/plain",
		},
		{
			name:           "wrong number",
			expectedStatus: http.StatusUnprocessableEntity,
			body:           bytes.NewReader([]byte("1")),
			token:          token,
			contentType:    "text/plain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, path, tc.body)
			request.Header.Add("Authorization", tc.token)
			request.Header.Add("Content-Type", tc.contentType)
			c := e.NewContext(request, recorder)
			c.SetPath(path)

			e.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedStatus, recorder.Result().StatusCode, "wrong status code")

		})
	}
}
