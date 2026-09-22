package client

import (
	"net/http/httptest"
	"testing"
)

func TestApplyTickertapeRouteHeaders(t *testing.T) {
	tests := []struct {
		name, url, version, device string
	}{
		{"premium ratings", "https://api.tickertape.in/stocks/ratings/RELI", "8.14.0", ""},
		{"analysis ai", "https://analyze.api.tickertape.in/stocks/aiSummary/RELI", "8.14.0", ""},
		{"credit", "https://auth.api.tickertape.in/user/credit/combined/v4", "3.5.0", "web"},
		{"portfolio", "https://ecosystem.api.tickertape.in/portfolio/v4/holdings/status", "8.0.0", "web"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			applyTickertapeRouteHeaders(req)
			if got := req.Header.Get("Accept-Version"); got != tt.version {
				t.Fatalf("Accept-Version = %q, want %q", got, tt.version)
			}
			if got := req.Header.Get("X-Device-Type"); got != tt.device {
				t.Fatalf("X-Device-Type = %q, want %q", got, tt.device)
			}
		})
	}
}
