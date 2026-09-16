package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerExecutorRoutes 注册执行器路由。
func (s *Server) registerExecutorRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/executors", s.createExecutor)
	mux.HandleFunc("GET /api/executors", s.listExecutors)
	mux.HandleFunc("GET /api/executors/{id}", s.getExecutor)
	mux.HandleFunc("PUT /api/executors/{id}", s.updateExecutor)
	mux.HandleFunc("DELETE /api/executors/{id}", s.deleteExecutor)
	mux.HandleFunc("POST /api/executors/{id}/transition", s.transitionExecutor)
	mux.HandleFunc("POST /api/executors/{id}/heartbeat", s.heartbeat)
}

type createExecutorRequest struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Capacity int    `json:"capacity"`
}

func (s *Server) createExecutor(w http.ResponseWriter, r *http.Request) {
	var req createExecutorRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.CreateExecutor(model.Executor{Name: req.Name, Addr: req.Addr, Capacity: req.Capacity})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, e)
}

func (s *Server) listExecutors(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ExecutorFilter{
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListExecutors(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getExecutor(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetExecutor(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

type updateExecutorRequest struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

func (s *Server) updateExecutor(w http.ResponseWriter, r *http.Request) {
	var req updateExecutorRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.UpdateExecutor(r.PathValue("id"), model.Executor{
		Name: req.Name, Addr: req.Addr, Capacity: req.Capacity, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) deleteExecutor(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteExecutor(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionExecutorRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionExecutor(w http.ResponseWriter, r *http.Request) {
	var req transitionExecutorRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.TransitionExecutor(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.Heartbeat(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}
