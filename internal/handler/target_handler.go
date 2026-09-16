package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerTargetRoutes 注册目标服务路由。
func (s *Server) registerTargetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/targets", s.createTarget)
	mux.HandleFunc("GET /api/targets", s.listTargets)
	mux.HandleFunc("GET /api/targets/{id}", s.getTarget)
	mux.HandleFunc("PUT /api/targets/{id}", s.updateTarget)
	mux.HandleFunc("DELETE /api/targets/{id}", s.deleteTarget)
	mux.HandleFunc("PATCH /api/targets/{id}/status", s.updateTargetStatus)
}

type createTargetRequest struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

func (s *Server) createTarget(w http.ResponseWriter, r *http.Request) {
	var req createTargetRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.CreateTarget(model.Target{Name: req.Name, URL: req.URL, Method: req.Method, Headers: req.Headers})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, t)
}

func (s *Server) listTargets(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.TargetFilter{
		Method:  r.URL.Query().Get("method"),
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListTargets(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getTarget(w http.ResponseWriter, r *http.Request) {
	t, err := s.svc.GetTarget(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}

type updateTargetRequest struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Status  string            `json:"status"`
}

func (s *Server) updateTarget(w http.ResponseWriter, r *http.Request) {
	var req updateTargetRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.UpdateTarget(r.PathValue("id"), model.Target{
		Name: req.Name, URL: req.URL, Method: req.Method, Headers: req.Headers, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}

func (s *Server) deleteTarget(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteTarget(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type updateTargetStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) updateTargetStatus(w http.ResponseWriter, r *http.Request) {
	var req updateTargetStatusRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.UpdateTargetStatus(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}
