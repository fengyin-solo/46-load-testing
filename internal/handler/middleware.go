package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"loadtest/pkg/httpx"
)

// authMiddleware 校验 X-API-Key，仅对 /api/ 路径生效。
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-API-Key")
		if s.cfg == nil || key == "" || key != s.cfg.APIKey {
			httpx.Unauthorized(w, "缺少或无效的 API Key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateBucket 固定窗口限流计数器。
type rateBucket struct {
	count int
	reset time.Time
}

// rateLimitMiddleware 基于客户端 IP 的固定窗口限流。
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	limit := 1000
	window := time.Minute
	if s.cfg != nil {
		if s.cfg.RateLimit > 0 {
			limit = s.cfg.RateLimit
		}
		if s.cfg.RateWindowSec > 0 {
			window = time.Duration(s.cfg.RateWindowSec) * time.Second
		}
	}
	var mu sync.Mutex
	buckets := make(map[string]*rateBucket)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		now := time.Now()
		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.reset) {
			buckets[ip] = &rateBucket{count: 1, reset: now.Add(window)}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		b.count++
		mu.Unlock()
		if b.count > limit {
			httpx.TooManyRequests(w, "请求过于频繁，请稍后再试")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP 提取客户端 IP。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		return host[:i]
	}
	return host
}
