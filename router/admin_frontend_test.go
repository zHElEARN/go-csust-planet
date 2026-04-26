package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/zHElEARN/go-csust-planet/middleware"
)

func TestAdminFrontendServesStaticFilesAndSpaFallback(t *testing.T) {
	distDir := t.TempDir()
	writeFrontendTestFile(t, filepath.Join(distDir, "index.html"), "INDEX")
	writeFrontendTestFile(t, filepath.Join(distDir, "_app", "immutable", "app.js"), "console.log('admin');")

	r := newFrontendTestRouter(t, AdminFrontendConfig{
		DistDir: distDir,
	})

	resp := performFrontendRequest(t, r, http.MethodGet, "/admin")
	assertStatus(t, resp, http.StatusOK)
	if !strings.Contains(resp.Body.String(), "INDEX") {
		t.Fatalf("expected admin index body, got %q", resp.Body.String())
	}

	resp = performFrontendRequest(t, r, http.MethodGet, "/admin/dashboard")
	assertStatus(t, resp, http.StatusOK)
	if !strings.Contains(resp.Body.String(), "INDEX") {
		t.Fatalf("expected spa fallback body, got %q", resp.Body.String())
	}

	resp = performFrontendRequest(t, r, http.MethodGet, "/admin/_app/immutable/app.js")
	assertStatus(t, resp, http.StatusOK)
	if !strings.Contains(resp.Body.String(), "console.log('admin');") {
		t.Fatalf("expected admin asset body, got %q", resp.Body.String())
	}

	resp = performFrontendRequest(t, r, http.MethodGet, "/admin/missing.js")
	assertStatus(t, resp, http.StatusNotFound)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestAdminFrontendProxiesToDevServer(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(req.URL.Path + "?" + req.URL.RawQuery))
	}))
	t.Cleanup(upstream.Close)

	r := newFrontendTestRouter(t, AdminFrontendConfig{
		DevServerURL: upstream.URL,
	})

	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/admin/settings?tab=general")
	if err != nil {
		t.Fatalf("expected proxy request to succeed, got error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("expected proxy response body to be readable, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, resp.StatusCode, string(body))
	}
	if string(body) != "/admin/settings?tab=general" {
		t.Fatalf("expected proxied request path to be preserved, got %q", body)
	}
}

func newFrontendTestRouter(t *testing.T, adminFrontend AdminFrontendConfig) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gin.Recovery())

	adminGroup := r.Group("/v1/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware(testAdminToken))
	adminGroup.GET("/announcements", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	mountAdminFrontendWithConfig(r, adminFrontend)
	return r
}

func performFrontendRequest(t *testing.T, r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func writeFrontendTestFile(t *testing.T, filePath, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create test frontend directory: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test frontend file: %v", err)
	}
}
