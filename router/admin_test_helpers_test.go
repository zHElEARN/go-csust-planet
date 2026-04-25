package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/controller"
	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/service"
	"github.com/zHElEARN/go-csust-planet/testsupport"
)

const testAdminToken = "admin-test-token"

type stubAuthService struct{}

func (stubAuthService) Login(string) (dto.LoginResponse, error) {
	return dto.LoginResponse{}, errors.New("not implemented")
}

type stubElectricityTaskService struct{}

func (stubElectricityTaskService) Sync(uuid.UUID, dto.SyncElectricityTaskRequest) error {
	return errors.New("not implemented")
}

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
	resetRouterTestTables(t, db)

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

	handler := controller.NewHandler(controller.Dependencies{
		DB:                     testDB,
		AuthService:            stubAuthService{},
		ElectricityTaskService: stubElectricityTaskService{},
		AdminAppVersionService: service.NewAdminAppVersionService(testDB),
	})

	gin.SetMode(gin.TestMode)
	return SetupRouter(Dependencies{
		Handler:          handler,
		AppMode:          config.AppConfig.AppMode,
		SwaggerPassword:  config.AppConfig.SwaggerPassword,
		AdminBearerToken: config.AppConfig.AdminBearerToken,
	}), db
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg, ok := testsupport.LoadTestDBConfig()
	if !ok {
		t.Skip("skipping PostgreSQL integration test: set TEST_DB_HOST/TEST_DB_PORT/TEST_DB_USER/TEST_DB_PASSWORD/TEST_DB_NAME")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode, cfg.TimeZone,
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

func resetRouterTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(
		"TRUNCATE TABLE announcements, app_versions, semester_calendars, campus_map_features RESTART IDENTITY CASCADE",
	).Error; err != nil {
		t.Fatalf("failed to reset router test tables: %v", err)
	}
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
