// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"loadtest/internal/config"
	"loadtest/internal/model"
	"loadtest/internal/service"
	"loadtest/internal/store"
	"loadtest/pkg/httpx"
	"loadtest/pkg/logger"
)

// Server HTTP 处理器集合。
type Server struct {
	svc *service.Service
	log *logger.Logger
	cfg *config.Config
}

// NewServer 构造 Server。
func NewServer(svc *service.Service, log *logger.Logger, cfg *config.Config) *Server {
	return &Server{svc: svc, log: log, cfg: cfg}
}

// Routes 装配全部路由与中间件。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	s.registerTargetRoutes(mux)
	s.registerScenarioRoutes(mux)
	s.registerTestPlanRoutes(mux)
	s.registerExecutorRoutes(mux)
	s.registerExecutionRoutes(mux)
	s.registerMetricSampleRoutes(mux)
	s.registerAssertionRoutes(mux)
	s.registerReportRoutes(mux)
	s.registerScheduleRoutes(mux)
	s.registerStatsRoutes(mux)
	s.registerExportRoutes(mux)
	s.registerHealthRoutes(mux)

	// 挂载前端静态页面。
	mux.Handle("GET /", http.FileServer(http.Dir("web")))

	return s.loggingMiddleware(s.recoveryMiddleware(s.authMiddleware(s.rateLimitMiddleware(mux))))
}

// maxPageSize 返回分页上限。
func (s *Server) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}

// loggingMiddleware 记录请求日志。
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// recoveryMiddleware 捕获 panic 并返回 500。
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Errorf("panic: %v\n%s", rec, debug.Stack())
				httpx.InternalError(w, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// writeServiceError 将业务错误映射为 HTTP 响应。
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case model.IsValidationError(err):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, store.ErrNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, store.ErrConflict):
		httpx.Conflict(w, err.Error())
	default:
		httpx.InternalError(w, err.Error())
	}
}
