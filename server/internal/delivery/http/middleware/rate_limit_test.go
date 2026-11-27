package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name, remote, xff, want string
	}{
		{"direct caller from outside", "203.0.113.7:1234", "1.2.3.4", "203.0.113.7"},
		{"behind the site's Caddy", "172.18.0.5:1234", "198.51.100.9, 10.0.0.2", "198.51.100.9"},
		{"a forged leftmost hop is ignored", "172.18.0.5:1234", "1.1.1.1, 198.51.100.9", "198.51.100.9"},
		{"only private hops", "172.18.0.5:1234", "10.0.0.2", "172.18.0.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remote
			req.Header.Set("X-Forwarded-For", tt.xff)
			if got := ByClientIP(req); got != tt.want {
				t.Errorf("ByClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLimiterRefills(t *testing.T) {
	now := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	l := newLimiter(time.Minute, 2)
	l.now = func() time.Time { return now }

	if !l.allow("a") || !l.allow("a") || l.allow("a") {
		t.Fatal("want 2 allowed, then denied")
	}
	now = now.Add(time.Minute)
	if !l.allow("a") || l.allow("a") {
		t.Error("want one more after an interval")
	}
}
