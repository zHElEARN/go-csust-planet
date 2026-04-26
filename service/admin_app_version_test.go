package service

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

func TestAdminAppVersionServiceCRUDAndConflict(t *testing.T) {
	db := openServiceTestDB(t)
	appVersionService := NewAdminAppVersionService(db)

	req := dto.AdminAppVersionUpsertRequest{
		Platform:      "ios",
		VersionCode:   intPtr(100),
		VersionName:   "1.0.0",
		IsForceUpdate: boolPtr(false),
		ReleaseNotes:  "initial",
		DownloadURL:   "https://example.com/ios-100",
	}

	created, err := appVersionService.Create(req)
	if err != nil {
		t.Fatalf("expected create to succeed: %v", err)
	}

	_, err = appVersionService.Create(req)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected duplicate create to return ErrConflict, got %v", err)
	}

	got, err := appVersionService.Get(created.ID)
	if err != nil {
		t.Fatalf("expected get to succeed: %v", err)
	}
	if got.VersionCode != 100 || got.Platform != "ios" {
		t.Fatalf("unexpected created version: %+v", got)
	}

	updated, err := appVersionService.Update(created.ID, dto.AdminAppVersionUpsertRequest{
		Platform:      "ios",
		VersionCode:   intPtr(101),
		VersionName:   "1.0.1",
		IsForceUpdate: boolPtr(true),
		ReleaseNotes:  "patch",
		DownloadURL:   "https://example.com/ios-101",
	})
	if err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}
	if updated.VersionCode != 101 || !updated.IsForceUpdate {
		t.Fatalf("unexpected updated version: %+v", updated)
	}

	list, err := appVersionService.List()
	if err != nil {
		t.Fatalf("expected list to succeed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("unexpected version list: %+v", list)
	}

	if err := appVersionService.Delete(created.ID); err != nil {
		t.Fatalf("expected delete to succeed: %v", err)
	}
	if err := appVersionService.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleting missing row to return ErrNotFound, got %v", err)
	}
}

func TestAdminAppVersionServiceCreateIsConcurrencySafe(t *testing.T) {
	db := openPersistentServiceTestDB(t)
	appVersionService := NewAdminAppVersionService(db)
	versionCode := int(time.Now().UnixNano() % 1_000_000_000)

	req := dto.AdminAppVersionUpsertRequest{
		Platform:      "android",
		VersionCode:   intPtr(versionCode),
		VersionName:   "10.0.0",
		IsForceUpdate: boolPtr(false),
		ReleaseNotes:  "concurrent",
		DownloadURL:   "https://example.com/android-1000",
	}
	t.Cleanup(func() {
		if err := db.Where("platform = ? AND version_code = ?", req.Platform, *req.VersionCode).Delete(&model.AppVersion{}).Error; err != nil {
			t.Fatalf("failed to cleanup concurrent app version test data: %v", err)
		}
	})

	start := make(chan struct{})
	results := make(chan error, 2)

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := appVersionService.Create(req)
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var successCount int
	var conflictCount int
	for err := range results {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, ErrConflict):
			conflictCount++
		default:
			t.Fatalf("unexpected concurrent create error: %v", err)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one success and one conflict, got success=%d conflict=%d", successCount, conflictCount)
	}

	var count int64
	if err := db.Model(&model.AppVersion{}).Where("platform = ? AND version_code = ?", req.Platform, *req.VersionCode).Count(&count).Error; err != nil {
		t.Fatalf("failed to count app versions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one row after concurrent create, got %d", count)
	}
}

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func TestAdminAppVersionServiceGetMissingVersion(t *testing.T) {
	db := openServiceTestDB(t)
	appVersionService := NewAdminAppVersionService(db)

	_, err := appVersionService.Get(uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdminAppVersionServiceListByPlatformAndCheckUpdate(t *testing.T) {
	db := openServiceTestDB(t)
	appVersionService := NewAdminAppVersionService(db)

	createServiceTestAppVersion(t, db, model.AppVersion{
		Platform:      "ios",
		VersionCode:   100,
		VersionName:   "1.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "initial",
		DownloadURL:   "https://example.com/ios-100",
	})
	mid := createServiceTestAppVersion(t, db, model.AppVersion{
		Platform:      "ios",
		VersionCode:   200,
		VersionName:   "2.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "minor",
		DownloadURL:   "https://example.com/ios-200",
	})
	createServiceTestAppVersion(t, db, model.AppVersion{
		Platform:      "android",
		VersionCode:   50,
		VersionName:   "5.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "android",
		DownloadURL:   "https://example.com/android-50",
	})

	versions, err := appVersionService.ListByPlatform("ios")
	if err != nil {
		t.Fatalf("expected list by platform to succeed: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 ios versions, got %d", len(versions))
	}
	if versions[0].VersionCode != mid.VersionCode || versions[1].VersionCode != 100 {
		t.Fatalf("unexpected ios version order: %+v", versions)
	}

	emptyResult, err := appVersionService.CheckUpdate("android", 100)
	if err != nil {
		t.Fatalf("expected empty platform check to succeed: %v", err)
	}
	if emptyResult.HasUpdate || emptyResult.IsForceUpdate || emptyResult.LatestVersion == nil {
		t.Fatalf("expected latest android version without update, got %+v", emptyResult)
	}

	noVersionResult, err := appVersionService.CheckUpdate("web", 1)
	if err != nil {
		t.Fatalf("expected missing platform check to succeed: %v", err)
	}
	if noVersionResult.HasUpdate || noVersionResult.IsForceUpdate || noVersionResult.LatestVersion != nil {
		t.Fatalf("expected missing platform to return empty result, got %+v", noVersionResult)
	}

	upToDateResult, err := appVersionService.CheckUpdate("ios", mid.VersionCode)
	if err != nil {
		t.Fatalf("expected up-to-date check to succeed: %v", err)
	}
	if upToDateResult.HasUpdate || upToDateResult.IsForceUpdate || upToDateResult.LatestVersion == nil {
		t.Fatalf("unexpected up-to-date result: %+v", upToDateResult)
	}

	nonForceResult, err := appVersionService.CheckUpdate("ios", 150)
	if err != nil {
		t.Fatalf("expected non-force check to succeed: %v", err)
	}
	if !nonForceResult.HasUpdate || nonForceResult.IsForceUpdate || nonForceResult.LatestVersion == nil {
		t.Fatalf("unexpected non-force result: %+v", nonForceResult)
	}

	latest := createServiceTestAppVersion(t, db, model.AppVersion{
		Platform:      "ios",
		VersionCode:   300,
		VersionName:   "3.0.0",
		IsForceUpdate: true,
		ReleaseNotes:  "force",
		DownloadURL:   "https://example.com/ios-300",
	})

	forceResult, err := appVersionService.CheckUpdate("ios", 150)
	if err != nil {
		t.Fatalf("expected force check to succeed: %v", err)
	}
	if !forceResult.HasUpdate || !forceResult.IsForceUpdate || forceResult.LatestVersion == nil {
		t.Fatalf("unexpected force result: %+v", forceResult)
	}
	if forceResult.LatestVersion.VersionCode != latest.VersionCode {
		t.Fatalf("expected latest version code %d, got %+v", latest.VersionCode, forceResult.LatestVersion)
	}
}
