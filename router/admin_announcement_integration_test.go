package router

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/zHElEARN/go-csust-planet/dto"
)

func TestAdminAnnouncementCRUD(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/admin/announcements/not-a-uuid", nil, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements/"+uuid.NewString(), nil, testAdminToken)
	assertStatus(t, resp, http.StatusNotFound)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/announcements", map[string]any{
		"title": "missing fields",
	}, testAdminToken)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/announcements", map[string]any{
		"title":    "公告 A",
		"content":  "内容 A",
		"isActive": false,
		"isBanner": false,
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var first dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &first)
	if first.IsActive {
		t.Fatalf("expected first announcement to be inactive")
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/announcements", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var publicList []dto.AnnouncementResponse
	decodeJSONResponse(t, resp, &publicList)
	if len(publicList) != 0 {
		t.Fatalf("expected inactive announcement to be hidden from public list, got %d items", len(publicList))
	}

	resp = performRequest(t, r, http.MethodPost, "/v1/admin/announcements", map[string]any{
		"title":    "公告 B",
		"content":  "内容 B",
		"isActive": true,
		"isBanner": true,
	}, testAdminToken)
	assertStatus(t, resp, http.StatusCreated)

	var second dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &second)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements", nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var adminList []dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &adminList)
	if len(adminList) != 2 {
		t.Fatalf("expected 2 announcements in admin list, got %d", len(adminList))
	}
	if adminList[0].ID != second.ID {
		t.Fatalf("expected latest announcement first, got %s", adminList[0].ID)
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements/"+first.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var detail dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &detail)
	if detail.Content != "内容 A" {
		t.Fatalf("expected announcement content to match, got %q", detail.Content)
	}

	resp = performRequest(t, r, http.MethodPut, "/v1/admin/announcements/"+first.ID, map[string]any{
		"title":    "公告 A 已更新",
		"content":  "内容 A 已更新",
		"isActive": true,
		"isBanner": true,
	}, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var updated dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &updated)
	if !updated.IsActive || !updated.IsBanner {
		t.Fatalf("expected updated announcement to be active banner")
	}

	resp = performRequest(t, r, http.MethodGet, "/v1/config/announcements", nil, "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSONResponse(t, resp, &publicList)
	if len(publicList) != 2 {
		t.Fatalf("expected 2 active announcements in public list, got %d", len(publicList))
	}

	resp = performRequest(t, r, http.MethodDelete, "/v1/admin/announcements/"+first.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNoContent)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements/"+first.ID, nil, testAdminToken)
	assertStatus(t, resp, http.StatusNotFound)
}
