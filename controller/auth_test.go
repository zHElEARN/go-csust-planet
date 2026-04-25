package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/service"
)

type authServiceStub struct {
	resp dto.LoginResponse
	err  error
}

func (s authServiceStub) Login(token string) (dto.LoginResponse, error) {
	return s.resp, s.err
}

func TestLoginPreservesLegacyErrorMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "unauthorized",
			serviceErr:     service.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "获取用户信息失败或 Token 已过期",
		},
		{
			name:           "user query failed",
			serviceErr:     service.ErrUserQueryFailed,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "数据库查询出错",
		},
		{
			name:           "user create failed",
			serviceErr:     service.ErrUserCreateFailed,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "创建用户失败",
		},
		{
			name:           "token generate failed",
			serviceErr:     service.ErrTokenGenerateFailed,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "生成令牌失败",
		},
		{
			name:           "unexpected failure",
			serviceErr:     errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "登录失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"token":"demo"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")

			handler := NewHandler(Dependencies{
				AuthService: authServiceStub{err: tt.serviceErr},
			})

			handler.Login(ctx)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var resp dto.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Error != tt.expectedError {
				t.Fatalf("expected error %q, got %q", tt.expectedError, resp.Error)
			}
		})
	}
}
