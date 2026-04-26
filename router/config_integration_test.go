package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/config"
	"github.com/zHElEARN/go-csust-planet/dto"
	"github.com/zHElEARN/go-csust-planet/model"
)

func TestConfigAnnouncementsReturnsEmptyList(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/config/announcements", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var announcements []dto.AnnouncementResponse
	decodeJSONResponse(t, resp, &announcements)
	if len(announcements) != 0 {
		t.Fatalf("expected empty announcements list, got %d items", len(announcements))
	}
}

func TestConfigAnnouncementsFiltersAndOrdersActiveAnnouncements(t *testing.T) {
	r := newAdminTestRouter(t)

	older := createTestAnnouncement(t, model.Announcement{
		Title:     "较早公告",
		Content:   "较早内容",
		IsActive:  true,
		IsBanner:  false,
		CreatedAt: time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC),
	})
	createTestAnnouncement(t, model.Announcement{
		Title:     "隐藏公告",
		Content:   "不应返回",
		IsActive:  false,
		IsBanner:  true,
		CreatedAt: time.Date(2026, time.January, 3, 10, 0, 0, 0, time.UTC),
	})
	newer := createTestAnnouncement(t, model.Announcement{
		Title:     "最新公告",
		Content:   "最新内容",
		IsActive:  true,
		IsBanner:  true,
		CreatedAt: time.Date(2026, time.January, 4, 10, 0, 0, 0, time.UTC),
	})

	resp := performRequest(t, r, http.MethodGet, "/v1/config/announcements", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var announcements []dto.AnnouncementResponse
	decodeJSONResponse(t, resp, &announcements)
	if len(announcements) != 2 {
		t.Fatalf("expected 2 active announcements, got %d", len(announcements))
	}
	if announcements[0].ID != newer.ID.String() || announcements[1].ID != older.ID.String() {
		t.Fatalf("unexpected announcements order: %+v", announcements)
	}
	if announcements[0].Title != newer.Title || announcements[0].Content != newer.Content {
		t.Fatalf("unexpected latest announcement payload: %+v", announcements[0])
	}
	if !announcements[0].IsBanner || announcements[1].IsBanner {
		t.Fatalf("unexpected banner flags: %+v", announcements)
	}

	var rawAnnouncements []map[string]any
	decodeJSONResponse(t, resp, &rawAnnouncements)
	if _, exists := rawAnnouncements[0]["isActive"]; exists {
		t.Fatalf("expected public announcement payload to omit isActive field")
	}
}

func TestConfigCampusMapReturnsFeatureCollection(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/config/campus-map", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var empty dto.CampusMapResponse
	decodeJSONResponse(t, resp, &empty)
	if empty.Type != "FeatureCollection" {
		t.Fatalf("expected FeatureCollection, got %q", empty.Type)
	}
	if len(empty.Features) != 0 {
		t.Fatalf("expected empty campus map features, got %d", len(empty.Features))
	}

	first := createTestCampusMapFeature(t, model.CampusMapFeature{
		Type: "Feature",
		Properties: model.FeatureProperties{
			Name:     "图书馆",
			Campus:   "云塘",
			Category: "building",
		},
		Geometry: model.FeatureGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{{112.1, 28.1}, {112.2, 28.1}, {112.2, 28.2}, {112.1, 28.1}},
			},
		},
	})
	createTestCampusMapFeature(t, model.CampusMapFeature{
		Type: "Feature",
		Properties: model.FeatureProperties{
			Name:     "体育馆",
			Campus:   "金盆岭",
			Category: "venue",
		},
		Geometry: model.FeatureGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{{113.1, 27.1}, {113.2, 27.1}, {113.2, 27.2}, {113.1, 27.1}},
			},
		},
	})

	resp = performRequest(t, r, http.MethodGet, "/v1/config/campus-map", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var campusMap dto.CampusMapResponse
	decodeJSONResponse(t, resp, &campusMap)
	if campusMap.Type != "FeatureCollection" {
		t.Fatalf("expected FeatureCollection, got %q", campusMap.Type)
	}
	if len(campusMap.Features) != 2 {
		t.Fatalf("expected 2 campus map features, got %d", len(campusMap.Features))
	}

	featuresByName := make(map[string]dto.CampusMapFeatureResponse, len(campusMap.Features))
	for _, feature := range campusMap.Features {
		featuresByName[feature.Properties.Name] = feature
	}

	library, ok := featuresByName[first.Properties.Name]
	if !ok {
		t.Fatalf("expected campus map to contain %q, got %+v", first.Properties.Name, campusMap.Features)
	}
	if library.Type != first.Type {
		t.Fatalf("expected feature type %q, got %q", first.Type, library.Type)
	}
	if library.Properties.Campus != first.Properties.Campus ||
		library.Properties.Category != first.Properties.Category {
		t.Fatalf("unexpected feature properties: %+v", library.Properties)
	}
	if library.Geometry.Type != first.Geometry.Type {
		t.Fatalf("unexpected feature geometry type: %+v", library.Geometry)
	}
	if len(library.Geometry.Coordinates) != 1 ||
		len(library.Geometry.Coordinates[0]) != 4 {
		t.Fatalf("unexpected feature coordinates: %+v", library.Geometry.Coordinates)
	}
	if library.Geometry.Coordinates[0][1][0] != 112.2 ||
		library.Geometry.Coordinates[0][1][1] != 28.1 {
		t.Fatalf("unexpected coordinate mapping: %+v", library.Geometry.Coordinates)
	}
}

func TestConfigAppVersionsValidatesPlatformAndFiltersByPlatform(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/config/app-versions", nil, "")
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions?platform=windows", nil, "")
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions?platform=ios", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var empty []dto.AppVersionResponse
	decodeJSONResponse(t, resp, &empty)
	if len(empty) != 0 {
		t.Fatalf("expected empty ios app version list, got %d", len(empty))
	}

	createTestAppVersion(t, model.AppVersion{
		Platform:      "ios",
		VersionCode:   100,
		VersionName:   "1.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "old ios release",
		DownloadURL:   "https://example.com/ios-100",
		CreatedAt:     time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC),
	})
	latestIOS := createTestAppVersion(t, model.AppVersion{
		Platform:      "ios",
		VersionCode:   200,
		VersionName:   "2.0.0",
		IsForceUpdate: true,
		ReleaseNotes:  "latest ios release",
		DownloadURL:   "https://example.com/ios-200",
		CreatedAt:     time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC),
	})
	createTestAppVersion(t, model.AppVersion{
		Platform:      "android",
		VersionCode:   300,
		VersionName:   "3.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "android release",
		DownloadURL:   "https://example.com/android-300",
		CreatedAt:     time.Date(2026, time.January, 3, 10, 0, 0, 0, time.UTC),
	})

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions?platform=ios", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var versions []dto.AppVersionResponse
	decodeJSONResponse(t, resp, &versions)
	if len(versions) != 2 {
		t.Fatalf("expected 2 ios app versions, got %d", len(versions))
	}
	if versions[0].VersionCode != 200 || versions[1].VersionCode != 100 {
		t.Fatalf("unexpected ios app version order: %+v", versions)
	}
	if versions[0].Platform != "ios" || versions[1].Platform != "ios" {
		t.Fatalf("expected only ios versions, got %+v", versions)
	}
	if versions[0].VersionName != latestIOS.VersionName ||
		versions[0].DownloadURL != latestIOS.DownloadURL ||
		versions[0].IsForceUpdate != latestIOS.IsForceUpdate {
		t.Fatalf("unexpected latest ios version payload: %+v", versions[0])
	}
}

func TestConfigAppVersionCheckValidatesAndReturnsLatestVersionState(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check", nil, "")
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios", nil, "")
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=windows&currentVersionCode=100", nil, "")
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios&currentVersionCode=100", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var empty dto.CheckAppVersionResponse
	decodeJSONResponse(t, resp, &empty)
	if empty.HasUpdate || empty.IsForceUpdate || empty.LatestVersion != nil {
		t.Fatalf("unexpected empty check response: %+v", empty)
	}

	createTestAppVersion(t, model.AppVersion{
		Platform:      "ios",
		VersionCode:   100,
		VersionName:   "1.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "first release",
		DownloadURL:   "https://example.com/ios-100",
	})
	latest := createTestAppVersion(t, model.AppVersion{
		Platform:      "ios",
		VersionCode:   200,
		VersionName:   "2.0.0",
		IsForceUpdate: false,
		ReleaseNotes:  "minor update",
		DownloadURL:   "https://example.com/ios-200",
	})

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios&currentVersionCode=200", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var upToDate dto.CheckAppVersionResponse
	decodeJSONResponse(t, resp, &upToDate)
	if upToDate.HasUpdate || upToDate.IsForceUpdate || upToDate.LatestVersion == nil {
		t.Fatalf("unexpected up-to-date response: %+v", upToDate)
	}
	if upToDate.LatestVersion.VersionCode != latest.VersionCode {
		t.Fatalf("expected latest version code %d, got %+v", latest.VersionCode, upToDate.LatestVersion)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios&currentVersionCode=150", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var nonForce dto.CheckAppVersionResponse
	decodeJSONResponse(t, resp, &nonForce)
	if !nonForce.HasUpdate || nonForce.IsForceUpdate || nonForce.LatestVersion == nil {
		t.Fatalf("unexpected non-force update response: %+v", nonForce)
	}
	if nonForce.LatestVersion.VersionCode != latest.VersionCode {
		t.Fatalf("expected latest version code %d, got %+v", latest.VersionCode, nonForce.LatestVersion)
	}

	forceLatest := createTestAppVersion(t, model.AppVersion{
		Platform:      "ios",
		VersionCode:   300,
		VersionName:   "3.0.0",
		IsForceUpdate: true,
		ReleaseNotes:  "force update",
		DownloadURL:   "https://example.com/ios-300",
	})

	resp = performRequest(t, r, http.MethodGet, "/v1/config/app-versions/check?platform=ios&currentVersionCode=150", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var force dto.CheckAppVersionResponse
	decodeJSONResponse(t, resp, &force)
	if !force.HasUpdate || !force.IsForceUpdate || force.LatestVersion == nil {
		t.Fatalf("unexpected force update response: %+v", force)
	}
	if force.LatestVersion.VersionCode != forceLatest.VersionCode {
		t.Fatalf("expected latest version code %d, got %+v", forceLatest.VersionCode, force.LatestVersion)
	}
}

func TestConfigSemesterCalendarsListAndDetail(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var empty []dto.SemesterCalendarListResponse
	decodeJSONResponse(t, resp, &empty)
	if len(empty) != 0 {
		t.Fatalf("expected empty semester calendar list, got %d", len(empty))
	}

	older := createTestSemesterCalendar(t, model.SemesterCalendar{
		SemesterCode:  "2024-2025-1",
		Title:         "2024-2025学年度校历",
		Subtitle:      "第一学期",
		CalendarStart: time.Date(2024, time.September, 1, 0, 0, 0, 0, time.UTC),
		CalendarEnd:   time.Date(2025, time.January, 20, 0, 0, 0, 0, time.UTC),
		SemesterStart: time.Date(2024, time.September, 2, 0, 0, 0, 0, time.UTC),
		SemesterEnd:   time.Date(2025, time.January, 12, 0, 0, 0, 0, time.UTC),
	})
	newer := createTestSemesterCalendar(t, model.SemesterCalendar{
		SemesterCode:  "2024-2025-2",
		Title:         "2024-2025学年度校历",
		Subtitle:      "第二学期",
		CalendarStart: time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC),
		CalendarEnd:   time.Date(2025, time.July, 10, 0, 0, 0, 0, time.UTC),
		SemesterStart: time.Date(2025, time.February, 17, 0, 0, 0, 0, time.UTC),
		SemesterEnd:   time.Date(2025, time.June, 29, 0, 0, 0, 0, time.UTC),
		Notes: []model.CalendarNote{
			{Row: 1, Content: "开学准备"},
			{Row: 2, Content: "考试周", NeedNumber: true},
		},
		CustomWeekRanges: []model.CustomWeekRange{
			{StartRow: 3, EndRow: 4, Content: "劳动节假期"},
		},
	})

	resp = performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var calendars []dto.SemesterCalendarListResponse
	decodeJSONResponse(t, resp, &calendars)
	if len(calendars) != 2 {
		t.Fatalf("expected 2 semester calendars, got %d", len(calendars))
	}
	if calendars[0].SemesterCode != newer.SemesterCode || calendars[1].SemesterCode != older.SemesterCode {
		t.Fatalf("unexpected semester calendar order: %+v", calendars)
	}
	if calendars[0].Title != newer.Title || calendars[0].Subtitle != newer.Subtitle {
		t.Fatalf("unexpected semester calendar list payload: %+v", calendars[0])
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars/"+newer.SemesterCode, nil, "")
	assertStatus(t, resp, http.StatusOK)

	var detail dto.SemesterCalendarDetailResponse
	decodeJSONResponse(t, resp, &detail)
	if detail.SemesterCode != newer.SemesterCode ||
		detail.Title != newer.Title ||
		detail.Subtitle != newer.Subtitle {
		t.Fatalf("unexpected semester calendar detail payload: %+v", detail)
	}
	if !detail.CalendarStart.Equal(newer.CalendarStart) ||
		!detail.CalendarEnd.Equal(newer.CalendarEnd) ||
		!detail.SemesterStart.Equal(newer.SemesterStart) ||
		!detail.SemesterEnd.Equal(newer.SemesterEnd) {
		t.Fatalf("unexpected semester calendar date fields: %+v", detail)
	}
	if len(detail.Notes) != 2 || detail.Notes[1].Content != "考试周" || !detail.Notes[1].NeedNumber {
		t.Fatalf("unexpected semester calendar notes: %+v", detail.Notes)
	}
	if len(detail.CustomWeekRanges) != 1 || detail.CustomWeekRanges[0].Content != "劳动节假期" {
		t.Fatalf("unexpected custom week ranges: %+v", detail.CustomWeekRanges)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars/2099-2100-1", nil, "")
	assertStatus(t, resp, http.StatusNotFound)
}

func createTestAnnouncement(t *testing.T, announcement model.Announcement) model.Announcement {
	t.Helper()

	if announcement.ID == uuid.Nil {
		announcement.ID = uuid.New()
	}

	values := map[string]any{
		"id":        announcement.ID,
		"title":     announcement.Title,
		"content":   announcement.Content,
		"is_active": announcement.IsActive,
		"is_banner": announcement.IsBanner,
	}
	if !announcement.CreatedAt.IsZero() {
		values["created_at"] = announcement.CreatedAt
	}

	if err := config.DB.Model(&model.Announcement{}).Create(values).Error; err != nil {
		t.Fatalf("failed to create test announcement: %v", err)
	}
	if err := config.DB.First(&announcement, "id = ?", announcement.ID).Error; err != nil {
		t.Fatalf("failed to reload test announcement: %v", err)
	}

	return announcement
}

func createTestCampusMapFeature(t *testing.T, feature model.CampusMapFeature) model.CampusMapFeature {
	t.Helper()

	if err := config.DB.Create(&feature).Error; err != nil {
		t.Fatalf("failed to create test campus map feature: %v", err)
	}

	return feature
}

func createTestAppVersion(t *testing.T, version model.AppVersion) model.AppVersion {
	t.Helper()

	if err := config.DB.Create(&version).Error; err != nil {
		t.Fatalf("failed to create test app version: %v", err)
	}

	return version
}

func createTestSemesterCalendar(t *testing.T, calendar model.SemesterCalendar) model.SemesterCalendar {
	t.Helper()

	if calendar.ID == uuid.Nil {
		calendar.ID = uuid.New()
	}

	if err := config.DB.Create(&calendar).Error; err != nil {
		t.Fatalf("failed to create test semester calendar: %v", err)
	}

	return calendar
}
