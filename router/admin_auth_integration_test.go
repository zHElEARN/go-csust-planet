package router

import (
	"net/http"
	"testing"

	"github.com/zHElEARN/go-csust-planet/dto"
)

func TestAdminAuthMiddleware(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodGet, "/v1/admin/announcements", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements", nil, "wrong-token")
	assertStatus(t, resp, http.StatusUnauthorized)

	resp = performRequestWithAuthorization(t, r, http.MethodGet, "/v1/admin/announcements", nil, "bearer "+testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	resp = performRequest(t, r, http.MethodGet, "/v1/admin/announcements", nil, testAdminToken)
	assertStatus(t, resp, http.StatusOK)

	var list []dto.AdminAnnouncementResponse
	decodeJSONResponse(t, resp, &list)
	if len(list) != 0 {
		t.Fatalf("expected empty announcement list, got %d items", len(list))
	}
}
