package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grizlaz/ya-gophermart/internal/domain/user"
	mock_api "github.com/grizlaz/ya-gophermart/internal/infrastructure/api/mock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserLogin(t *testing.T) {
	path := "/api/user/login"
	userID := int64(1)
	login := "test"
	pwd := "test"
	requestBody, err := json.Marshal(user.User{
		Login:    login,
		Password: pwd,
	})
	require.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := http.StatusOK

		authService := mock_api.NewMockuserAuthService(ctrl)
		authService.EXPECT().Auth(context.Background(), login, pwd).Return(userID, nil)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(requestBody))

		c := echo.New().NewContext(request, recorder)
		c.SetPath(path)
		handler := HandleUserLogin(authService)

		require.NoError(t, handler(c))

		assert.Equal(t, expectedStatus, recorder.Result().StatusCode, "wrong status code")

		token := recorder.Result().Header.Get(user.AuthHeaderName)
		assert.NotEmpty(t, token)
	})
}
