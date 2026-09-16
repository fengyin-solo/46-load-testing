package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerScenarioRoutes 注册场景路由。
func (s *Server) registerScenarioRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/scenarios", s.createScenario)
	mux.HandleFunc("GET /api/scenarios", s.listScenarios)
	mux.HandleFunc("GET /api/scenarios/{id}", s.getScenario)
	mux.HandleFunc("PUT /api/scenarios/{id}", s.updateScenario)
	mux.HandleFunc("DELETE /api/scenarios/{id}", s.deleteScenario)
	mux.HandleFunc("POST /api/scenarios/{id}/transition", s.transitionScenario)
}

type createScenarioRequest struct {
	Name        string `json:"name"`
	TargetID    string `json:"target_id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Concurrent  int    `json:"concurrent"`
	DurationSec int    `json:"duration_sec"`
	RampUpSec   int    `json:"ramp_up_sec"`
	Body        string `json:"body"`
}

func (s *Server) createScenario(w http.ResponseWriter, r *http.Request) {
	var req createScenarioRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.CreateScenario(model.Scenario{
		Name: req.Name, TargetID: req.TargetID, Method: req.Method, Path: req.Path,
		Concurrent: req.Concurrent, DurationSec: req.DurationSec, RampUpSec: req.RampUpSec, Body: req.Body,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, sc)
}

func (s *Server) listScenarios(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ScenarioFilter{
		TargetID: r.URL.Query().Get("target_id"),
		Method:   r.URL.Query().Get("method"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListScenarios(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getScenario(w http.ResponseWriter, r *http.Request) {
	sc, err := s.svc.GetScenario(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

type updateScenarioRequest struct {
	Name        string `json:"name"`
	TargetID    string `json:"target_id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Concurrent  int    `json:"concurrent"`
	DurationSec int    `json:"duration_sec"`
	RampUpSec   int    `json:"ramp_up_sec"`
	Body        string `json:"body"`
	Status      string `json:"status"`
}

func (s *Server) updateScenario(w http.ResponseWriter, r *http.Request) {
	var req updateScenarioRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.UpdateScenario(r.PathValue("id"), model.Scenario{
		Name: req.Name, TargetID: req.TargetID, Method: req.Method, Path: req.Path,
		Concurrent: req.Concurrent, DurationSec: req.DurationSec, RampUpSec: req.RampUpSec,
		Body: req.Body, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

func (s *Server) deleteScenario(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteScenario(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionScenarioRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionScenario(w http.ResponseWriter, r *http.Request) {
	var req transitionScenarioRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.TransitionScenario(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}
