package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/dto"
	jwtutil "github.com/zHElEARN/go-csust-planet/utils/jwt"
)

func TestAuthMiddlewareAcceptsCaseInsensitiveBearerPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	prevConfig := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: "test-jwt-secret"}
	t.Cleanup(func() {
		config.AppConfig = prevConfig
	})

	token, err := jwtutil.GenerateToken(uuid.New(), "20230001", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "bearer "+token)

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.Code)
	}
}

func TestAuthMiddlewarePreservesErrorMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		header        string
		expectedCode  int
		expectedError string
	}{
		{
			name:          "missing header",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "未提供身份验证令牌",
		},
		{
			name:          "invalid format",
			header:        "token demo",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "身份验证令牌格式不正确",
		},
		{
			name:          "invalid token",
			header:        "Bearer invalid",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "无效或过期的令牌",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevConfig := config.AppConfig
			config.AppConfig = &config.Config{JWTSecret: "test-jwt-secret"}
			t.Cleanup(func() {
				config.AppConfig = prevConfig
			})

			r := gin.New()
			r.Use(AuthMiddleware())
			r.GET("/protected", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			if resp.Code != tt.expectedCode {
				t.Fatalf("expected status %d, got %d", tt.expectedCode, resp.Code)
			}

			var body dto.ErrorResponse
			if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if body.Error != tt.expectedError {
				t.Fatalf("expected error %q, got %q", tt.expectedError, body.Error)
			}
		})
	}
}
