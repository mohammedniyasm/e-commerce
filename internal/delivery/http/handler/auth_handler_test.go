package handler

import (
	"ecommerce/config"
	"ecommerce/internal/delivery/http/middleware"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutWithoutRefreshCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authUseCase := mocks.NewMockAuthUseCase(t)
	claims := &interfaces.AccessClaims{
		UserID:    "1",
		JTI:       "access-token-id",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	authUseCase.EXPECT().Logout(mock.Anything, "", claims).Return(nil).Once()

	handler := NewAuthHandler(authUseCase, config.CookieConfig{
		RefreshTokenName: "refresh_token",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	context.Set(middleware.AccessClaimsKey, claims)

	handler.Logout(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "logout successful")
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "refresh_token=")
	assert.Contains(t, recorder.Header().Get("Set-Cookie"), "Max-Age=0")
}
