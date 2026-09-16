package handler

import (
	"net/http"
	"runtime"
	"time"

	"loadtest/pkg/httpx"
)

// registerHealthRoutes 注册健康检查路由。
func (s *Server) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.healthz)
}

type healthStatus struct {
	Status     string `json:"status"`
	Uptime     string `json:"uptime"`
	GoVersion  string `json:"go_version"`
	NumCPU     int    `json:"num_cpu"`
	NumGoroutine int  `json:"num_goroutine"`
}

var startedAt = time.Now()

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, healthStatus{
		Status:       "ok",
		Uptime:       time.Since(startedAt).String(),
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
	})
}
