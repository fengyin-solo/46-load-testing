package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerExecutionRoutes 注册执行记录路由。
func (s *Server) registerExecutionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/executions", s.createExecution)
	mux.HandleFunc("GET /api/executions", s.listExecutions)
	mux.HandleFunc("GET /api/executions/{id}", s.getExecution)
	mux.HandleFunc("PUT /api/executions/{id}", s.updateExecution)
	mux.HandleFunc("DELETE /api/executions/{id}", s.deleteExecution)
	mux.HandleFunc("POST /api/executions/{id}/start", s.startExecution)
	mux.HandleFunc("POST /api/executions/{id}/complete", s.completeExecution)
	mux.HandleFunc("POST /api/executions/{id}/fail", s.failExecution)
	mux.HandleFunc("POST /api/executions/{id}/cancel", s.cancelExecution)
}

type createExecutionRequest struct {
	PlanID     string `json:"plan_id"`
	ExecutorID string `json:"executor_id"`
}

func (s *Server) createExecution(w http.ResponseWriter, r *http.Request) {
	var req createExecutionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.CreateExecution(model.Execution{PlanID: req.PlanID, ExecutorID: req.ExecutorID})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, e)
}

func (s *Server) listExecutions(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ExecutionFilter{
		PlanID:     r.URL.Query().Get("plan_id"),
		ExecutorID: r.URL.Query().Get("executor_id"),
		Status:     r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListExecutions(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getExecution(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetExecution(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

type updateExecutionRequest struct {
	TotalRequests int64  `json:"total_requests"`
	ErrorCount    int64  `json:"error_count"`
	Status        string `json:"status"`
}

func (s *Server) updateExecution(w http.ResponseWriter, r *http.Request) {
	var req updateExecutionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.UpdateExecution(r.PathValue("id"), model.Execution{
		TotalRequests: req.TotalRequests, ErrorCount: req.ErrorCount, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) deleteExecution(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteExecution(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) startExecution(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.StartExecution(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

type completeExecutionRequest struct {
	TotalRequests int64 `json:"total_requests"`
	ErrorCount    int64 `json:"error_count"`
}

func (s *Server) completeExecution(w http.ResponseWriter, r *http.Request) {
	var req completeExecutionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.CompleteExecution(r.PathValue("id"), req.TotalRequests, req.ErrorCount)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

type failExecutionRequest struct {
	TotalRequests int64 `json:"total_requests"`
	ErrorCount    int64 `json:"error_count"`
}

func (s *Server) failExecution(w http.ResponseWriter, r *http.Request) {
	var req failExecutionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.FailExecution(r.PathValue("id"), req.TotalRequests, req.ErrorCount)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) cancelExecution(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.CancelExecution(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}
