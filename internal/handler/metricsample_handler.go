package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerMetricSampleRoutes 注册指标样本路由。
func (s *Server) registerMetricSampleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/metrics", s.createMetricSample)
	mux.HandleFunc("POST /api/metrics/batch", s.batchCreateMetricSamples)
	mux.HandleFunc("GET /api/metrics", s.listMetricSamples)
	mux.HandleFunc("GET /api/metrics/{id}", s.getMetricSample)
	mux.HandleFunc("DELETE /api/metrics/{id}", s.deleteMetricSample)
}

type createMetricSampleRequest struct {
	ExecutionID  string  `json:"execution_id"`
	Timestamp    string  `json:"timestamp"`
	TPS          float64 `json:"tps"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
}

func (s *Server) createMetricSample(w http.ResponseWriter, r *http.Request) {
	var req createMetricSampleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	m, err := s.svc.CreateMetricSample(model.MetricSample{
		ExecutionID:  req.ExecutionID,
		TPS:          req.TPS,
		AvgLatencyMs: req.AvgLatencyMs,
		P99LatencyMs: req.P99LatencyMs,
		ErrorRate:    req.ErrorRate,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, m)
}

type batchCreateMetricSamplesRequest struct {
	ExecutionID string                   `json:"execution_id"`
	Samples     []model.MetricSample     `json:"samples"`
}

func (s *Server) batchCreateMetricSamples(w http.ResponseWriter, r *http.Request) {
	var req batchCreateMetricSamplesRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if len(req.Samples) == 0 {
		httpx.BadRequest(w, "样本列表不能为空")
		return
	}
	items, err := s.svc.CreateMetricSamples(req.ExecutionID, req.Samples)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, items)
}

func (s *Server) listMetricSamples(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.MetricSampleFilter{
		ExecutionID:  r.URL.Query().Get("execution_id"),
		MinTPS:       httpx.ParseQueryFloat(r, "min_tps", 0),
		MaxErrorRate: httpx.ParseQueryFloat(r, "max_error_rate", 0),
	}
	items, total, err := s.svc.ListMetricSamples(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getMetricSample(w http.ResponseWriter, r *http.Request) {
	m, err := s.svc.GetMetricSample(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, m)
}

func (s *Server) deleteMetricSample(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteMetricSample(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
