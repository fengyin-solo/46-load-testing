package handler

import (
	"net/http"

	"loadtest/pkg/httpx"
)

// registerStatsRoutes 注册统计路由。
func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/executions/{id}", s.statsExecution)
	mux.HandleFunc("GET /api/stats/by-plan", s.statsByPlan)
	mux.HandleFunc("GET /api/stats/by-executor", s.statsByExecutor)
	mux.HandleFunc("GET /api/stats/top-scenarios", s.statsTopScenarios)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.Overview()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) statsExecution(w http.ResponseWriter, r *http.Request) {
	agg, err := s.svc.AggregateExecution(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, agg)
}

func (s *Server) statsByPlan(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.AggregationsByPlan()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) statsByExecutor(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.AggregationsByExecutor()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) statsTopScenarios(w http.ResponseWriter, r *http.Request) {
	n := httpx.ParseQueryInt(r, "n", 10)
	items, err := s.svc.TopScenarios(n)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}
