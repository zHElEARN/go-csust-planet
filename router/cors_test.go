package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSDisabledByDefault(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequestWithOrigin(t, r, http.MethodGet, "/v1/config/announcements", nil, "https://planet.zhelearn.com")
	assertStatus(t, resp, http.StatusOK)
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS header by default, got %q", got)
	}
}

func TestCORSAllowsWhitelistedOrigins(t *testing.T) {
	r := newCORSTestRouter([]string{"https://planet.zhelearn.com"})

	for _, path := range []string{"/v1/ping", "/v2/ping"} {
		resp := performRequestWithOrigin(t, r, http.MethodGet, path, nil, "https://planet.zhelearn.com")
		assertStatus(t, resp, http.StatusOK)
		if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://planet.zhelearn.com" {
			t.Fatalf("expected matching CORS origin for %s, got %q", path, got)
		}
		if got := resp.Header().Get("Vary"); got != "Origin" {
			t.Fatalf("expected Vary: Origin for %s, got %q", path, got)
		}

		resp = performRequestWithOrigin(t, r, http.MethodOptions, path, nil, "https://planet.zhelearn.com")
		assertStatus(t, resp, http.StatusNoContent)
		if got := resp.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Fatalf("expected allow methods header to be set for %s", path)
		}
		if got := resp.Header().Get("Access-Control-Allow-Headers"); got == "" {
			t.Fatalf("expected allow headers header to be set for %s", path)
		}
	}
}

func TestCORSRejectsNonWhitelistedOrigins(t *testing.T) {
	r := newCORSTestRouter([]string{"https://planet.zhelearn.com"})

	for _, path := range []string{"/v1/ping", "/v2/ping"} {
		resp := performRequestWithOrigin(t, r, http.MethodGet, path, nil, "https://evil.example.com")
		assertStatus(t, resp, http.StatusOK)
		if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no CORS header for non-whitelisted origin on %s, got %q", path, got)
		}
	}
}

func TestCORSAllowsNoOriginsWhenEmpty(t *testing.T) {
	r := newCORSTestRouter(nil)

	for _, path := range []string{"/v1/ping", "/v2/ping"} {
		resp := performRequestWithOrigin(t, r, http.MethodGet, path, nil, "https://planet.zhelearn.com")
		assertStatus(t, resp, http.StatusOK)
		if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no CORS header with an empty allow list on %s, got %q", path, got)
		}
	}
}

func newCORSTestRouter(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(corsMiddleware(origins))
	r.GET("/v1/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/v2/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func performRequestWithOrigin(t *testing.T, r *gin.Engine, method, path string, body any, origin string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, requestBody)
	req.Header.Set("Origin", origin)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}
