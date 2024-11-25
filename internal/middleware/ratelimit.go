package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attempt
	max      int
	window   time.Duration
	logger   *slog.Logger
}

type attempt struct {
	count     int
	startTime time.Time
}

func NewRateLimiter(max int, window time.Duration, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		attempts: make(map[string]*attempt),
		max:      max,
		window:   window,
		logger:   logger,
	}

	go rl.cleanup()

	return rl
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if ip := net.ParseIP(clientIP); ip != nil {
				return clientIP
			}
		}
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		if ip := net.ParseIP(xrip); ip != nil {
			return xrip
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if a, exists := rl.attempts[ip]; exists {
		if now.Sub(a.startTime) > rl.window {
			a.count = 1
			a.startTime = now
			return true
		}

		if a.count >= rl.max {
			rl.logger.Warn("rate limit exceeded",
				"ip", ip,
				"count", a.count,
				"window", rl.window)
			return false
		}

		a.count++
		return true
	}

	rl.attempts[ip] = &attempt{
		count:     1,
		startTime: now,
	}
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, a := range rl.attempts {
			if now.Sub(a.startTime) > rl.window {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		rl.logger.Info("incoming request",
			"ip", ip,
			"path", r.URL.Path,
			"method", r.Method)

		if !rl.isAllowed(ip) {
			rl.logger.Warn("rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path)

			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.max))
		w.Header().Set("X-RateLimit-Window", rl.window.String())

		next.ServeHTTP(w, r)
	})
}
