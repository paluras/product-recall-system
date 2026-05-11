package main

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type ipEntry struct {
	count     int
	windowEnd time.Time
	banned    bool
	banUntil  time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*ipEntry
	limit    int
	window   time.Duration
	banAfter int
	banFor   time.Duration
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		entries:  make(map[string]*ipEntry),
		limit:    3,
		window:   time.Hour,
		banAfter: 10,
		banFor:   24 * time.Hour,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.entries[ip]
	if !ok {
		rl.entries[ip] = &ipEntry{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}

	if e.banned {
		if now.Before(e.banUntil) {
			return false
		}
		e.banned = false
		e.count = 1
		e.windowEnd = now.Add(rl.window)
		return true
	}

	if now.After(e.windowEnd) {
		e.count = 1
		e.windowEnd = now.Add(rl.window)
		return true
	}

	e.count++
	if e.count > rl.banAfter {
		e.banned = true
		e.banUntil = now.Add(rl.banFor)
		return false
	}

	return e.count <= rl.limit
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, e := range rl.entries {
			if !e.banned && now.After(e.windowEnd) {
				delete(rl.entries, ip)
			}
			if e.banned && now.After(e.banUntil) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
