// Package middleware wraps the router's handlers
package middleware

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/syauqiahmd/masjid-digital-site/server/pkg/apperror"
)

// Handler is an http.HandlerFunc that returns its error to the server's error handler
type Handler func(w http.ResponseWriter, r *http.Request) error

// RateLimit lets each key make burst requests at once, then one more every interval; past that
// it answers TOO_MANY_REQUESTS (429). Counts live in memory and restart from zero with the server.
func RateLimit(interval time.Duration, burst int, key func(*http.Request) string) func(Handler) Handler {
	l := newLimiter(interval, burst)
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if !l.allow(key(r)) {
				return apperror.New(apperror.CodeTooManyRequests, errors.New("rate limit exceeded"), "terlalu banyak percobaan, silakan coba lagi nanti")
			}
			return next(w, r)
		}
	}
}

// Everyone shares one rate limit, capping a route however many clients call it
func Everyone(*http.Request) string {
	return "everyone"
}

// limiter lets each key make burst requests at once, then one more every interval. Counts live in
// memory and restart from zero with the server.
type limiter struct {
	interval time.Duration
	burst    int
	now      func() time.Time

	mu      sync.Mutex
	buckets map[string]bucket
}

type bucket struct {
	tokens float64
	at     time.Time
}

func newLimiter(interval time.Duration, burst int) *limiter {
	return &limiter{interval: interval, burst: burst, now: time.Now, buckets: map[string]bucket{}}
}

// allow takes one token from key's bucket, refilled at one per interval
func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = bucket{tokens: float64(l.burst), at: now}
	}
	b.tokens = min(float64(l.burst), b.tokens+float64(now.Sub(b.at))/float64(l.interval))
	b.at = now

	allowed := b.tokens >= 1
	if allowed {
		b.tokens--
	}
	l.buckets[key] = b

	// a full bucket is the same as none, so drop those now and then to keep the map small
	if len(l.buckets) > 10_000 {
		full := time.Duration(l.burst) * l.interval
		for k, v := range l.buckets {
			if now.Sub(v.at) >= full {
				delete(l.buckets, k)
			}
		}
	}

	return allowed
}

// ByClientIP is the address a rate limit counts: the rightmost X-Forwarded-For address outside
// the private, loopback and link-local ranges, so a client cannot pick its own by writing the
// header. Only proxies on a private network (the site's Caddy) are trusted to append to it; a
// direct caller from outside is counted by its own address.
func ByClientIP(r *http.Request) string {
	remote, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remote = r.RemoteAddr
	}
	if !internal(remote) {
		return remote
	}

	hops := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
	for i := len(hops) - 1; i >= 0; i-- {
		hop := strings.TrimSpace(hops[i])
		if hop != "" && !internal(hop) {
			return hop
		}
	}

	return remote
}

func internal(addr string) bool {
	ip := net.ParseIP(addr)
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast())
}
