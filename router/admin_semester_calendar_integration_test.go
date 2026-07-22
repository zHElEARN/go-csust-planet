package router

import (
	"net/http"
	"testing"
)

func TestAdminSemesterCalendarCRUD(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/admin/semester-calendars/2099-2100-1", nil, testAdminToken)
	assertStatus(t, resp, http.StatusNotFound)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/semester-calendars", map[string]any{
		"semesterCode": "2025-2026-1",
		"title":        "missing fields",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/semester-calendars", map[string]any{
		"semesterCode":  "2025-2026-1",
		"title":         "2025-2026学年度校历",
		"subtitle":      "第一学期",
		"calendarStart": "2025-09-01T00:00:00Z",
		"calendarEnd":   "2026-01-20T00:00:00Z",
		"semesterStart": "2025-09-02T00:00:00Z",
		"semesterEnd":   "2026-01-11T00:00:00Z",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var first AdminSemesterCalendarResponse
	decodeJSONResponse(t, resp, &first)
	if first.SemesterCode != "2025-2026-1" {
		t.Fatalf("expected created semester code to match, got %q", first.SemesterCode)
	}
	if len(first.Notes) != 0 || len(first.CustomWeekRanges) != 0 {
		t.Fatalf("expected empty notes and custom week ranges, got %+v %+v", first.Notes, first.CustomWeekRanges)
	}

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/semester-calendars", map[string]any{
		"semesterCode":  "2025-2026-2",
		"title":         "2025-2026学年度校历",
		"subtitle":      "第二学期",
		"calendarStart": "2026-02-23T00:00:00Z",
		"calendarEnd":   "2026-07-10T00:00:00Z",
		"semesterStart": "2026-03-02T00:00:00Z",
		"semesterEnd":   "2026-06-28T00:00:00Z",
		"notes": []map[string]any{
			{"row": 1, "content": "开学准备"},
			{"row": 2, "content": "考试周", "needNumber": true},
		},
		"customWeekRanges": []map[string]any{
			{"startRow": 8, "endRow": 8, "content": "清明节假期"},
		},
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var second AdminSemesterCalendarResponse
	decodeJSONResponse(t, resp, &second)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/semester-calendars", nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var adminList []AdminSemesterCalendarResponse
	decodeJSONResponse(t, resp, &adminList)
	if len(adminList) != 2 {
		t.Fatalf("expected 2 semester calendars in admin list, got %d", len(adminList))
	}
	if adminList[0].SemesterCode != second.SemesterCode || adminList[1].SemesterCode != first.SemesterCode {
		t.Fatalf("unexpected semester calendar ordering: %+v", adminList)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/semester-calendars/"+second.SemesterCode, nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var detail AdminSemesterCalendarResponse
	decodeJSONResponse(t, resp, &detail)
	if detail.Title != "2025-2026学年度校历" || len(detail.Notes) != 2 || len(detail.CustomWeekRanges) != 1 {
		t.Fatalf("unexpected semester calendar detail payload: %+v", detail)
	}

	resp = performRequest(t, r, http.MethodPut, "/v1/admin/semester-calendars/"+first.SemesterCode, map[string]any{
		"semesterCode":  "2025-2026-1A",
		"title":         "2025-2026学年度校历（调整）",
		"subtitle":      "第一学期",
		"calendarStart": "2025-09-01T00:00:00Z",
		"calendarEnd":   "2026-01-21T00:00:00Z",
		"semesterStart": "2025-09-03T00:00:00Z",
		"semesterEnd":   "2026-01-12T00:00:00Z",
		"notes": []map[string]any{
			{"row": 1, "content": "课程补退选"},
		},
	}, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var updated AdminSemesterCalendarResponse
	decodeJSONResponse(t, resp, &updated)
	if updated.SemesterCode != "2025-2026-1A" || updated.Title != "2025-2026学年度校历（调整）" {
		t.Fatalf("unexpected updated semester calendar payload: %+v", updated)
	}

	resp = performRequest(t, r, http.MethodPut, "/v1/admin/semester-calendars/"+updated.SemesterCode, map[string]any{
		"semesterCode":  second.SemesterCode,
		"title":         updated.Title,
		"subtitle":      updated.Subtitle,
		"calendarStart": "2025-09-01T00:00:00Z",
		"calendarEnd":   "2026-01-21T00:00:00Z",
		"semesterStart": "2025-09-03T00:00:00Z",
		"semesterEnd":   "2026-01-12T00:00:00Z",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusConflict)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/semester-calendars", map[string]any{
		"semesterCode":  second.SemesterCode,
		"title":         "duplicate",
		"subtitle":      "duplicate",
		"calendarStart": "2026-02-23T00:00:00Z",
		"calendarEnd":   "2026-07-10T00:00:00Z",
		"semesterStart": "2026-03-02T00:00:00Z",
		"semesterEnd":   "2026-06-28T00:00:00Z",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusConflict)

	resp = performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var publicList []SemesterCalendarListResponse
	decodeJSONResponse(t, resp, &publicList)
	if len(publicList) != 2 || publicList[0].SemesterCode != second.SemesterCode || publicList[1].SemesterCode != updated.SemesterCode {
		t.Fatalf("unexpected public semester calendar list: %+v", publicList)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/semester-calendars/"+updated.SemesterCode, nil, "")
	assertStatus(t, resp, http.StatusOK)

	var publicDetail SemesterCalendarDetailResponse
	decodeJSONResponse(t, resp, &publicDetail)
	if publicDetail.Title != updated.Title || len(publicDetail.Notes) != 1 || publicDetail.Notes[0].Content != "课程补退选" {
		t.Fatalf("unexpected public semester calendar detail: %+v", publicDetail)
	}

	resp = performRequest(t, r, http.MethodDelete, "/v1/admin/semester-calendars/"+updated.SemesterCode, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNoContent)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/semester-calendars/"+updated.SemesterCode, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNotFound)
}
