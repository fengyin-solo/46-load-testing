package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerScheduleRoutes 注册调度路由。
func (s *Server) registerScheduleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/schedules", s.createSchedule)
	mux.HandleFunc("GET /api/schedules", s.listSchedules)
	mux.HandleFunc("GET /api/schedules/{id}", s.getSchedule)
	mux.HandleFunc("PUT /api/schedules/{id}", s.updateSchedule)
	mux.HandleFunc("DELETE /api/schedules/{id}", s.deleteSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/transition", s.transitionSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/trigger", s.triggerSchedule)
}

type createScheduleRequest struct {
	PlanID   string `json:"plan_id"`
	CronExpr string `json:"cron_expr"`
}

func (s *Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	var req createScheduleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.CreateSchedule(model.Schedule{PlanID: req.PlanID, CronExpr: req.CronExpr})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, sc)
}

func (s *Server) listSchedules(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ScheduleFilter{
		PlanID: r.URL.Query().Get("plan_id"),
		Status: r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListSchedules(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getSchedule(w http.ResponseWriter, r *http.Request) {
	sc, err := s.svc.GetSchedule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

type updateScheduleRequest struct {
	PlanID   string `json:"plan_id"`
	CronExpr string `json:"cron_expr"`
	Status   string `json:"status"`
}

func (s *Server) updateSchedule(w http.ResponseWriter, r *http.Request) {
	var req updateScheduleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.UpdateSchedule(r.PathValue("id"), model.Schedule{
		PlanID: req.PlanID, CronExpr: req.CronExpr, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

func (s *Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteSchedule(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionScheduleRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionSchedule(w http.ResponseWriter, r *http.Request) {
	var req transitionScheduleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sc, err := s.svc.TransitionSchedule(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sc)
}

func (s *Server) triggerSchedule(w http.ResponseWriter, r *http.Request) {
	ex, err := s.svc.TriggerSchedule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, ex)
}
