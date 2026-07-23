package router

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zHElEARN/go-csust-planet/internal/announcement"
	"github.com/zHElEARN/go-csust-planet/internal/appversion"
	"github.com/zHElEARN/go-csust-planet/internal/campusmap"
	"github.com/zHElEARN/go-csust-planet/internal/health"
	"github.com/zHElEARN/go-csust-planet/internal/semestercalendar"
	"github.com/zHElEARN/go-csust-planet/testsupport"
)

const testAdminToken = "admin-test-token"

var activeTestDB *gorm.DB

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

	testDB, db := testsupport.OpenTestDB(t, useTransaction)
	resetRouterTestTables(t, testDB)
	previousTestDB := activeTestDB
	activeTestDB = testDB
	t.Cleanup(func() { activeTestDB = previousTestDB })

	gin.SetMode(gin.TestMode)
	return SetupRouter(Dependencies{
		HealthHandler:           health.NewHandler(testDB),
		AnnouncementHandler:     announcement.NewHandler(announcement.NewService(announcement.NewPostgresRepository(testDB))),
		AppVersionHandler:       appversion.NewHandler(appversion.NewService(appversion.NewPostgresRepository(testDB))),
		CampusMapHandler:        campusmap.NewHandler(campusmap.NewService(campusmap.NewPostgresRepository(testDB))),
		SemesterCalendarHandler: semestercalendar.NewHandler(semestercalendar.NewService(semestercalendar.NewPostgresRepository(testDB))),
		AppMode:                 "test",
		SwaggerPassword:         "test-swagger-password",
		AdminBearerToken:        testAdminToken,
		BusinessRequestTimeout:  time.Second,
	}), db
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
