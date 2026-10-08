package account

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCredentialRoutesPreventCachingBeforeLimiting(t *testing.T) {
	for _, row := range []struct {
		path   string
		status int
	}{
		{"/api/v1/account", http.StatusBadRequest},
		// This parser fixture deliberately has no Game UI dependency. The existing
		// bootstrap failure must be non-storable too, without touching a database.
		{"/api/v1/bootstrap", http.StatusInternalServerError},
		{"/api/v1/session", http.StatusBadRequest},
		{"/api/v1/session/refresh", http.StatusBadRequest},
	} {
		t.Run(row.path, func(t *testing.T) {
			now := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
			handler := refreshCensusRouter(t, &now)
			for _, status := range []int{row.status, http.StatusTooManyRequests} {
				request := httptest.NewRequest(http.MethodPost, row.path, strings.NewReader("{"))
				request.RemoteAddr = "192.0.2.42:1234"
				request.Header.Set("Content-Type", "application/json")
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				if response.Code != status {
					t.Fatalf("status=%d want=%d (body withheld)", response.Code, status)
				}
				if response.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("credential route status=%d is not protected by no-store", status)
				}
			}
		})
	}
}
