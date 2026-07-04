package router

import (
	"net/http"
	"testing"
)

func TestRemovedUserRoutesReturnNotFound(t *testing.T) {
	r := newAdminTestRouter(t)

	resp := performRequest(t, r, http.MethodPost, "/v1/auth/login", nil, "")
	assertStatus(t, resp, http.StatusNotFound)

	resp = performRequest(t, r, http.MethodPost, "/v1/task/electricity", nil, "")
	assertStatus(t, resp, http.StatusNotFound)
}
