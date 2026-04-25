package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
)

const testAdminToken = "admin-test-token"

type testDBConfig struct {
	host     string
	port     string
	user     string
	password string
	name     string
	sslMode  string
	timeZone string
}

var loadTestEnvOnce sync.Once

func newAdminTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	r, _ := newAdminTestRouterWithCleanup(t, true)
	return r
}

func newAdminPersistentTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	return newAdminTestRouterWithCleanup(t, false)
}

func newAdminTestRouterWithCleanup(t *testing.T, useTransaction bool) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := openTestDB(t)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		t.Fatalf("failed to enable pgcrypto extension: %v", err)
	}
	if err := config.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	testDB := db
	if useTransaction {
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("failed to begin transaction: %v", tx.Error)
		}
		testDB = tx

		t.Cleanup(func() {
			_ = tx.Rollback().Error
		})
	}

	prevConfig := config.AppConfig
	prevDB := config.DB

	config.AppConfig = &config.Config{
		AppMode:          "test",
		JWTSecret:        "test-jwt-secret",
		SwaggerPassword:  "test-swagger-password",
		AdminBearerToken: testAdminToken,
	}
	config.DB = testDB

	t.Cleanup(func() {
		config.DB = prevDB
		config.AppConfig = prevConfig
	})

	gin.SetMode(gin.TestMode)
	return SetupRouter(), db
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg, ok := loadTestDBConfig()
	if !ok {
		t.Skip("skipping PostgreSQL integration test: set TEST_DB_HOST/TEST_DB_PORT/TEST_DB_USER/TEST_DB_PASSWORD/TEST_DB_NAME")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.host, cfg.user, cfg.password, cfg.name, cfg.port, cfg.sslMode, cfg.timeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to obtain sql.DB from gorm: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func loadTestDBConfig() (testDBConfig, bool) {
	loadTestEnvOnce.Do(func() {
		_ = godotenv.Load(".env")
	})

	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	name := os.Getenv("TEST_DB_NAME")
	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return testDBConfig{}, false
	}

	sslMode := os.Getenv("TEST_DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timeZone := os.Getenv("TEST_DB_TIMEZONE")
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	return testDBConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		name:     name,
		sslMode:  sslMode,
		timeZone: timeZone,
	}, true
}

func performRequest(t *testing.T, r *gin.Engine, method, path string, body any, adminToken string) *httptest.ResponseRecorder {
	t.Helper()

	authHeader := ""
	if adminToken != "" {
		authHeader = "Bearer " + adminToken
	}

	return performRequestWithAuthorization(t, r, method, path, body, authHeader)
}

func performRequestWithAuthorization(t *testing.T, r *gin.Engine, method, path string, body any, authorization string) *httptest.ResponseRecorder {
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func decodeJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(resp.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response body %q: %v", resp.Body.String(), err)
	}
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if resp.Code != expected {
		t.Fatalf("expected status %d, got %d, body=%s", expected, resp.Code, resp.Body.String())
	}
}
