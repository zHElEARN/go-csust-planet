package router

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

func TestAdminAppVersionCRUD(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/admin/app-versions/not-a-uuid", nil, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform":      "windows",
		"versionCode":   100,
		"versionName":   "1.0.0",
		"isForceUpdate": false,
		"releaseNotes":  "notes",
		"downloadUrl":   "https://example.com/windows",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform": "ios",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform":      "ios",
		"versionCode":   100,
		"versionName":   "1.0.0",
		"isForceUpdate": false,
		"releaseNotes":  "first ios release",
		"downloadUrl":   "https://example.com/ios-100",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var iosV100 dto.AdminAppVersionResponse
	decodeJSONResponse(t, resp, &iosV100)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform":      "ios",
		"versionCode":   200,
		"versionName":   "2.0.0",
		"isForceUpdate": true,
		"releaseNotes":  "force update",
		"downloadUrl":   "https://example.com/ios-200",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var iosV200 dto.AdminAppVersionResponse
	decodeJSONResponse(t, resp, &iosV200)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform":      "android",
		"versionCode":   10,
		"versionName":   "1.0.0",
		"isForceUpdate": false,
		"releaseNotes":  "android release",
		"downloadUrl":   "https://example.com/android-10",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/app-versions", nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var adminList []dto.AdminAppVersionResponse
	decodeJSONResponse(t, resp, &adminList)
	if len(adminList) != 3 {
		t.Fatalf("expected 3 app versions in admin list, got %d", len(adminList))
	}
	if adminList[0].Platform != "android" || adminList[1].ID != iosV200.ID || adminList[2].ID != iosV100.ID {
		t.Fatalf("unexpected app version ordering: %+v", adminList)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/app-versions/"+iosV100.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var detail dto.AdminAppVersionResponse
	decodeJSONResponse(t, resp, &detail)
	if detail.VersionCode != 100 {
		t.Fatalf("expected version code 100, got %d", detail.VersionCode)
	}

	resp = performRequest(t, r, http.MethodPut, "/v1/admin/app-versions/"+iosV100.ID, map[string]any{
		"platform":      "ios",
		"versionCode":   150,
		"versionName":   "1.5.0",
		"isForceUpdate": false,
		"releaseNotes":  "minor update",
		"downloadUrl":   "https://example.com/ios-150",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var updated dto.AdminAppVersionResponse
	decodeJSONResponse(t, resp, &updated)
	if updated.VersionCode != 150 || updated.VersionName != "1.5.0" {
		t.Fatalf("unexpected updated version: %+v", updated)
	}

	resp = performRequest(t, r, http.MethodPut, "/v1/admin/app-versions/"+iosV100.ID, map[string]any{
		"platform":      "ios",
		"versionCode":   200,
		"versionName":   "2.0.0-conflict",
		"isForceUpdate": false,
		"releaseNotes":  "conflict update",
		"downloadUrl":   "https://example.com/ios-conflict",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusConflict)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", map[string]any{
		"platform":      "ios",
		"versionCode":   200,
		"versionName":   "2.0.0-duplicate",
		"isForceUpdate": false,
		"releaseNotes":  "duplicate",
		"downloadUrl":   "https://example.com/ios-duplicate",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusConflict)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions?platform=ios", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var publicVersions []dto.AppVersionResponse
	decodeJSONResponse(t, resp, &publicVersions)
	if len(publicVersions) != 2 || publicVersions[0].VersionCode != 200 || publicVersions[1].VersionCode != 150 {
		t.Fatalf("unexpected public app version list: %+v", publicVersions)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios&currentVersionCode=120", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var checkResp dto.CheckAppVersionResponse
	decodeJSONResponse(t, resp, &checkResp)
	if !checkResp.HasUpdate || !checkResp.IsForceUpdate || checkResp.LatestVersion == nil || checkResp.LatestVersion.VersionCode != 200 {
		t.Fatalf("unexpected version check response: %+v", checkResp)
	}

	resp = performRequest(t, r, http.MethodDelete, "/v1/admin/app-versions/"+iosV200.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNoContent)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/app-versions/"+iosV200.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNotFound)
}

func TestAdminAppVersionCreateIsConcurrencySafe(t *testing.T) {
	r, db := newAdminPersistentTestRouter(t)

	versionCode := int(time.Now().UnixNano() % 1_000_000_000)
	t.Cleanup(func() {
		if err := db.Where("platform = ? AND version_code = ?", "ios", versionCode).Delete(&model.AppVersion{}).Error; err != nil {
			t.Fatalf("failed to cleanup concurrent app version test data: %v", err)
		}
	})

	requestBody := map[string]any{
		"platform":      "ios",
		"versionCode":   versionCode,
		"versionName":   "9.9.9",
		"isForceUpdate": false,
		"releaseNotes":  "concurrent create",
		"downloadUrl":   "https://example.com/concurrent",
	}

	start := make(chan struct{})
	results := make(chan int, 2)
	bodies := make(chan string, 2)

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			resp := performRequest(t, r, http.MethodPost, "/v1/admin/app-versions", requestBody, testAdminToken)
			results <- resp.Code
			bodies <- resp.Body.String()
		}()
	}

	close(start)
	wg.Wait()
	close(results)
	close(bodies)

	var createdCount int
	var conflictCount int
	for code := range results {
		switch code {
		case http.StatusCreated:
			createdCount++
		case http.StatusConflict:
			conflictCount++
		default:
			t.Fatalf("unexpected concurrent create status: %d", code)
		}
	}

	if createdCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one create and one conflict, got created=%d conflict=%d", createdCount, conflictCount)
	}

	var count int64
	if err := db.Model(&model.AppVersion{}).Where("platform = ? AND version_code = ?", "ios", versionCode).Count(&count).Error; err != nil {
		t.Fatalf("failed to count concurrent app versions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one app version row after concurrent create, got %d", count)
	}
}
